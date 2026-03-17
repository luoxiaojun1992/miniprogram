package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/response"
)

// UploadController handles file upload requests.
type UploadController struct {
	uploadDir string
	baseURL   string
	log       *logrus.Logger
	cos       *cosUploader
	auditRepo repository.AuditLogRepository
}

// NewUploadController creates a new UploadController.
func NewUploadController(uploadDir, baseURL string, log *logrus.Logger) *UploadController {
	return &UploadController{uploadDir: uploadDir, baseURL: baseURL, log: log}
}

// NewUploadControllerWithCOS creates a new UploadController that uploads to COS.
func NewUploadControllerWithCOS(uploadDir, baseURL, endpoint, bucket string, log *logrus.Logger) *UploadController {
	if baseURL == "" {
		baseURL = strings.TrimRight(endpoint, "/")
	}
	return &UploadController{
		uploadDir: uploadDir,
		baseURL:   baseURL,
		log:       log,
		cos: &cosUploader{
			endpoint:      strings.TrimRight(endpoint, "/"),
			publicBaseURL: strings.TrimRight(baseURL, "/"),
			bucket:        bucket,
			client:        &http.Client{Timeout: 30 * time.Second},
		},
	}
}

func (c *UploadController) WithAuditRepo(auditRepo repository.AuditLogRepository) *UploadController {
	c.auditRepo = auditRepo
	return c
}

// UploadAvatar handles POST /upload/avatar.
func (c *UploadController) UploadAvatar(ctx *gin.Context) {
	c.uploadImageWithPrefix(ctx, "avatar")
}

// UploadArticleImage handles POST /upload/article/image.
func (c *UploadController) UploadArticleImage(ctx *gin.Context) {
	if !isAdminUser(ctx) {
		ctx.Error(apperrors.NewForbidden("仅管理员可上传文章图片", nil))
		return
	}
	c.uploadImageWithPrefix(ctx, "article-image")
}

// GenerateCourseVideoPresignURL handles GET /upload/course/video/presign.
func (c *UploadController) GenerateCourseVideoPresignURL(ctx *gin.Context) {
	if !isAdminUser(ctx) {
		ctx.Error(apperrors.NewForbidden("仅管理员可上传课程视频", nil))
		return
	}
	c.generatePresignUploadURL(ctx, "course-video", ".mp4")
}

// GenerateArticleAttachmentPresignURL handles GET /upload/article/attachment/presign.
func (c *UploadController) GenerateArticleAttachmentPresignURL(ctx *gin.Context) {
	if !isAdminUser(ctx) {
		ctx.Error(apperrors.NewForbidden("仅管理员可上传文章附件", nil))
		return
	}
	c.generatePresignUploadURL(ctx, "article-attachment", "")
}

// GenerateCourseAttachmentPresignURL handles GET /upload/course/attachment/presign.
func (c *UploadController) GenerateCourseAttachmentPresignURL(ctx *gin.Context) {
	if !isAdminUser(ctx) {
		ctx.Error(apperrors.NewForbidden("仅管理员可上传课程附件", nil))
		return
	}
	c.generatePresignUploadURL(ctx, "course-attachment", "")
}

// GenerateCourseVideoDownloadURL handles GET /download/course/video.
func (c *UploadController) GenerateCourseVideoDownloadURL(ctx *gin.Context) {
	c.generateTemporaryDownloadURL(ctx, "course-video")
}

// GenerateArticleAttachmentDownloadURL handles GET /download/article/attachment.
func (c *UploadController) GenerateArticleAttachmentDownloadURL(ctx *gin.Context) {
	c.generateTemporaryDownloadURL(ctx, "article-attachment")
}

// GenerateCourseAttachmentDownloadURL handles GET /download/course/attachment.
func (c *UploadController) GenerateCourseAttachmentDownloadURL(ctx *gin.Context) {
	c.generateTemporaryDownloadURL(ctx, "course-attachment")
}

// UploadImage handles POST /upload/image.
func (c *UploadController) UploadImage(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("获取文件失败", err))
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	if !allowed[ext] {
		ctx.Error(apperrors.NewBadRequest("不支持的文件类型", nil))
		return
	}
	if header.Size > 5*1024*1024 {
		ctx.Error(apperrors.NewBadRequest("文件大小超过5MB限制", nil))
		return
	}
	contentType := header.Header.Get("Content-Type")
	if contentType != "" && contentType != "application/octet-stream" && !strings.HasPrefix(contentType, "image/") {
		ctx.Error(apperrors.NewBadRequest("仅支持图片MIME类型", nil))
		return
	}
	if err := validateImageMagic(file, ext); err != nil {
		ctx.Error(apperrors.NewBadRequest("图片文件内容非法", err))
		return
	}

	imgType := ctx.PostForm("type")
	if imgType == "" {
		imgType = "article"
	}
	imgType = sanitizeUploadType(imgType)
	key := generateObjectKey(imgType, ext)
	url, err := c.saveFile(ctx.Request.Context(), file, header, key)
	if err != nil {
		ctx.Error(err)
		return
	}

	response.SuccessWithStatus(ctx, http.StatusOK, gin.H{"url": url, "key": key})
}

// UploadVideo handles POST /upload/video.
func (c *UploadController) UploadVideo(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("获取文件失败", err))
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".mp4" {
		ctx.Error(apperrors.NewBadRequest("仅支持mp4格式", nil))
		return
	}
	if header.Size > 500*1024*1024 {
		ctx.Error(apperrors.NewBadRequest("文件大小超过500MB限制", nil))
		return
	}

	key := generateObjectKey("video", ext)
	url, err := c.saveFile(ctx.Request.Context(), file, header, key)
	if err != nil {
		ctx.Error(err)
		return
	}

	response.Success(ctx, gin.H{"url": url, "duration": 0, "cover_url": ""})
}

// GeneratePresignURL handles GET /upload/presign.
func (c *UploadController) GeneratePresignURL(ctx *gin.Context) {
	if c.cos == nil {
		ctx.Error(apperrors.NewBadRequest("当前存储不支持预签名上传", nil))
		return
	}

	filename := strings.TrimSpace(ctx.Query("filename"))
	if filename == "" {
		ctx.Error(apperrors.NewBadRequest("filename不能为空", nil))
		return
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".mp4" {
		ctx.Error(apperrors.NewBadRequest("仅支持mp4预签名上传", nil))
		return
	}

	expiresIn := 900
	if raw := strings.TrimSpace(ctx.Query("expires_in")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 60 || v > 3600 {
			ctx.Error(apperrors.NewBadRequest("expires_in必须在60-3600秒之间", err))
			return
		}
		expiresIn = v
	}
	key := generateObjectKey("video", ext)
	expireAt := time.Now().Add(time.Duration(expiresIn) * time.Second).Unix()

	response.Success(ctx, gin.H{
		"key":        key,
		"url":        c.cos.objectURL(key),
		"put_url":    c.cos.presignPutURL(key, expiresIn),
		"expires_in": expiresIn,
		"expire_at":  expireAt,
	})
}

func (c *UploadController) uploadImageWithPrefix(ctx *gin.Context, prefix string) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("获取文件失败", err))
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	if !allowed[ext] {
		ctx.Error(apperrors.NewBadRequest("不支持的文件类型", nil))
		return
	}
	if header.Size > 5*1024*1024 {
		ctx.Error(apperrors.NewBadRequest("文件大小超过5MB限制", nil))
		return
	}
	contentType := header.Header.Get("Content-Type")
	if contentType != "" && contentType != "application/octet-stream" && !strings.HasPrefix(contentType, "image/") {
		ctx.Error(apperrors.NewBadRequest("仅支持图片MIME类型", nil))
		return
	}
	if err := validateImageMagic(file, ext); err != nil {
		ctx.Error(apperrors.NewBadRequest("图片文件内容非法", err))
		return
	}

	key := generateObjectKey(prefix, ext)
	uploadedURL, err := c.saveFile(ctx.Request.Context(), file, header, key)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.SuccessWithStatus(ctx, http.StatusOK, gin.H{"url": uploadedURL, "key": key})
}

func (c *UploadController) generatePresignUploadURL(ctx *gin.Context, businessType, forceExt string) {
	if c.cos == nil {
		ctx.Error(apperrors.NewBadRequest("当前存储不支持预签名上传", nil))
		return
	}

	filename := strings.TrimSpace(ctx.Query("filename"))
	if filename == "" {
		ctx.Error(apperrors.NewBadRequest("filename不能为空", nil))
		return
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if forceExt != "" && ext != forceExt {
		ctx.Error(apperrors.NewBadRequest("文件扩展名不合法", nil))
		return
	}
	if forceExt == "" && !attachmentTypePattern.MatchString(ext) {
		ctx.Error(apperrors.NewBadRequest("附件扩展名不支持", nil))
		return
	}

	expiresIn := 900
	if raw := strings.TrimSpace(ctx.Query("expires_in")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 60 || v > 3600 {
			ctx.Error(apperrors.NewBadRequest("expires_in必须在60-3600秒之间", err))
			return
		}
		expiresIn = v
	}

	key := generateObjectKey(businessType, ext)
	expireAt := time.Now().Add(time.Duration(expiresIn) * time.Second).Unix()
	if err := c.recordUploadBusiness(ctx, businessType, key, filename, expiresIn); err != nil {
		ctx.Error(err)
		return
	}

	response.Success(ctx, gin.H{
		"key":        key,
		"url":        c.cos.objectURL(key),
		"put_url":    c.cos.presignPutURL(key, expiresIn),
		"expires_in": expiresIn,
		"expire_at":  expireAt,
	})
}

func (c *UploadController) generateTemporaryDownloadURL(ctx *gin.Context, businessType string) {
	if c.cos == nil {
		ctx.Error(apperrors.NewBadRequest("当前存储不支持临时下载链接", nil))
		return
	}
	key := normalizeObjectKey(ctx.Query("key"))
	if key == "" {
		ctx.Error(apperrors.NewBadRequest("key不能为空", nil))
		return
	}
	if !strings.HasPrefix(key, businessType+"/") {
		ctx.Error(apperrors.NewForbidden("无权访问该文件", nil))
		return
	}

	expiresIn := 300
	if raw := strings.TrimSpace(ctx.Query("expires_in")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 60 || v > 3600 {
			ctx.Error(apperrors.NewBadRequest("expires_in必须在60-3600秒之间", err))
			return
		}
		expiresIn = v
	}
	response.Success(ctx, gin.H{
		"key":        key,
		"download":   c.cos.presignGetURL(key, expiresIn),
		"expires_in": expiresIn,
		"expire_at":  time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
	})
}

func (c *UploadController) recordUploadBusiness(ctx *gin.Context, businessType, key, filename string, expiresIn int) error {
	if c.auditRepo == nil {
		return nil
	}
	userID, _ := middleware.GetCurrentUserID(ctx)
	payload, _ := json.Marshal(gin.H{
		"business_type": businessType,
		"key":           key,
		"filename":      filename,
		"expires_in":    expiresIn,
	})
	err := c.auditRepo.Create(ctx.Request.Context(), &entity.AuditLog{
		UserID:      userID,
		Action:      "upload_presign",
		Module:      "file_upload",
		Description: "生成上传预签名链接",
		IPAddress:   ctx.ClientIP(),
		UserAgent:   ctx.Request.UserAgent(),
		RequestData: string(payload),
	})
	if err != nil {
		return apperrors.NewInternal("记录上传业务类型失败", err)
	}
	return nil
}

func (c *UploadController) saveFile(ctx context.Context, file io.Reader, header *multipart.FileHeader, key string) (string, error) {
	if c.cos != nil {
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		url, err := c.cos.upload(ctx, key, file, contentType)
		if err != nil {
			return "", apperrors.NewInternal("上传到对象存储失败", err)
		}
		return url, nil
	}

	savePath := filepath.Join(c.uploadDir, key)
	if err := os.MkdirAll(filepath.Dir(savePath), 0o755); err != nil {
		return "", apperrors.NewInternal("创建上传目录失败", err)
	}
	if err := saveUploadedFile(header, savePath); err != nil {
		return "", apperrors.NewInternal("保存文件失败", err)
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(c.baseURL, "/"), key), nil
}

func saveUploadedFile(header *multipart.FileHeader, savePath string) error {
	src, err := header.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

type cosUploader struct {
	endpoint      string
	publicBaseURL string
	bucket        string
	client        *http.Client
}

func (u *cosUploader) upload(ctx context.Context, key string, file io.Reader, contentType string) (string, error) {
	escapedKey := path.Clean("/" + key)
	if escapedKey == "/" {
		return "", fmt.Errorf("invalid object key")
	}
	escapedKey = strings.TrimPrefix(escapedKey, "/")
	if strings.Contains(escapedKey, "..") {
		return "", fmt.Errorf("invalid object key")
	}
	escapedPath := escapeObjectKey(escapedKey)
	targetURL := fmt.Sprintf("%s/%s/%s", u.endpoint, url.PathEscape(u.bucket), escapedPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, targetURL, file)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := u.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("cos upload failed status=%d body=%s", resp.StatusCode, string(body))
	}
	return u.objectURL(key), nil
}

var uploadTypePattern = regexp.MustCompile(`^(avatar|article|course|cover)$`)
var attachmentTypePattern = regexp.MustCompile(`^\.(pdf|doc|docx|xls|xlsx|ppt|pptx|zip|rar|7z|txt)$`)

func sanitizeUploadType(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if uploadTypePattern.MatchString(raw) {
		return raw
	}
	return "article"
}

func generateObjectKey(prefix, ext string) string {
	return fmt.Sprintf("%s/%s/%s%s", prefix, time.Now().Format("20060102"), uuid.NewString(), ext)
}

func (u *cosUploader) objectURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(u.publicBaseURL, "/"), url.PathEscape(u.bucket), escapeObjectKey(key))
}

func (u *cosUploader) presignPutURL(key string, expiresIn int) string {
	return fmt.Sprintf("%s/%s/%s?expires_in=%d", strings.TrimRight(u.endpoint, "/"), url.PathEscape(u.bucket), escapeObjectKey(key), expiresIn)
}

func (u *cosUploader) presignGetURL(key string, expiresIn int) string {
	return fmt.Sprintf("%s/%s/%s?expires_in=%d", strings.TrimRight(u.endpoint, "/"), url.PathEscape(u.bucket), escapeObjectKey(key), expiresIn)
}

func normalizeObjectKey(key string) string {
	clean := path.Clean("/" + strings.TrimSpace(key))
	if clean == "/" {
		return ""
	}
	clean = strings.TrimPrefix(clean, "/")
	if strings.Contains(clean, "..") {
		return ""
	}
	return clean
}

func isAdminUser(ctx *gin.Context) bool {
	userType, ok := middleware.GetCurrentUserType(ctx)
	return ok && userType >= 2
}

func escapeObjectKey(key string) string {
	parts := strings.Split(key, "/")
	escaped := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		escaped = append(escaped, url.PathEscape(part))
	}
	return strings.Join(escaped, "/")
}

func validateImageMagic(file multipart.File, ext string) error {
	header := make([]byte, 12)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return err
	}
	if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
		return seekErr
	}
	header = header[:n]

	switch ext {
	case ".jpg", ".jpeg":
		if len(header) >= 2 && header[0] == 0xFF && header[1] == 0xD8 {
			return nil
		}
	case ".png":
		if len(header) >= 8 && bytes.Equal(header[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'}) {
			return nil
		}
	case ".gif":
		if len(header) >= 6 && (bytes.Equal(header[:6], []byte("GIF87a")) || bytes.Equal(header[:6], []byte("GIF89a"))) {
			return nil
		}
	}
	return fmt.Errorf("invalid image magic")
}
