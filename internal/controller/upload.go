package controller

import (
	"bytes"
	"context"
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

	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/response"
)

// UploadController handles file upload requests.
type UploadController struct {
	uploadDir string
	baseURL   string
	log       *logrus.Logger
	cos       *cosUploader
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
