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
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/response"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
	"github.com/luoxiaojun1992/miniprogram/internal/service"
)

// UploadController handles file upload requests.
type UploadController struct {
	uploadDir string
	baseURL   string
	log       *logrus.Logger
	cos       *cosUploader
	auditRepo repository.AuditLogRepository
	uploadSvc service.UploadFileService
	access    *accessChecker

	articleRepo          repository.ArticleRepository
	courseRepo           repository.CourseRepository
	courseUnitRepo       repository.CourseUnitRepository
	articleAttachRepo    repository.ArticleAttachmentRepository
	courseAttachRepo     repository.CourseAttachmentRepository
	courseUnitAttachRepo repository.CourseUnitAttachmentRepository
}

type fileRecordPayload struct {
	Key        string `json:"key"`
	Filename   string `json:"filename"`
	Usage      string `json:"usage"`
	Category   string `json:"category"`
	StaticURL  string `json:"static_url,omitempty"`
	Business   string `json:"business,omitempty"`
	ExpiresIn  int    `json:"expires_in"`
	Protected  bool   `json:"protected"`
	ObjectPath string `json:"object_path,omitempty"`
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

func (c *UploadController) WithUploadService(uploadSvc service.UploadFileService) *UploadController {
	c.uploadSvc = uploadSvc
	return c
}

func (c *UploadController) WithPermissionRepos(
	contentPermRepo repository.ContentPermissionRepository,
	roleRepo repository.RoleRepository,
	articleRepo repository.ArticleRepository,
	courseRepo repository.CourseRepository,
	courseUnitRepo repository.CourseUnitRepository,
	articleAttachRepo repository.ArticleAttachmentRepository,
	courseAttachRepo repository.CourseAttachmentRepository,
	courseUnitAttachRepo repository.CourseUnitAttachmentRepository,
) *UploadController {
	c.access = newAccessChecker(contentPermRepo, roleRepo)
	c.articleRepo = articleRepo
	c.courseRepo = courseRepo
	c.courseUnitRepo = courseUnitRepo
	c.articleAttachRepo = articleAttachRepo
	c.courseAttachRepo = courseAttachRepo
	c.courseUnitAttachRepo = courseUnitAttachRepo
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

// GenerateAdminUploadPresignURL handles GET /admin/upload/files/presign.
func (c *UploadController) GenerateAdminUploadPresignURL(ctx *gin.Context) {
	if !isAdminUser(ctx) {
		ctx.Error(apperrors.NewForbidden("仅管理员可上传素材", nil))
		return
	}
	if c.uploadSvc == nil {
		if c.cos == nil {
			ctx.Error(apperrors.NewBadRequest("当前存储不支持预签名上传", nil))
			return
		}
		filename := strings.TrimSpace(ctx.Query("filename"))
		if filename == "" {
			ctx.Error(apperrors.NewBadRequest("filename不能为空", nil))
			return
		}
		usage := strings.ToLower(strings.TrimSpace(ctx.DefaultQuery("usage", "protected")))
		if usage != "embedded" && usage != "protected" {
			ctx.Error(apperrors.NewBadRequest("usage仅支持embedded或protected", nil))
			return
		}
		ext := strings.ToLower(filepath.Ext(filename))
		category := classifyFileCategory(ext)
		if category == "" {
			ctx.Error(apperrors.NewBadRequest("不支持的文件扩展名", nil))
			return
		}
		if usage == "embedded" && category == "attachment" {
			ctx.Error(apperrors.NewBadRequest("内嵌素材仅支持图片或视频", nil))
			return
		}
		expiresIn := 900
		if raw := strings.TrimSpace(ctx.Query("expires_in")); raw != "" {
			v, convErr := strconv.Atoi(raw)
			if convErr != nil || v < 60 || v > 3600 {
				ctx.Error(apperrors.NewBadRequest("expires_in必须在60-3600秒之间", convErr))
				return
			}
			expiresIn = v
		}
		prefix := usage + "-" + category
		key := generateObjectKey(prefix, ext)
		staticURL := ""
		if usage == "embedded" {
			staticURL = c.cos.objectURL(key)
		}
		payload := fileRecordPayload{
			Key:       key,
			Filename:  filename,
			Usage:     usage,
			Category:  category,
			StaticURL: staticURL,
			ExpiresIn: expiresIn,
			Protected: usage == "protected",
		}
		fileID, recordErr := c.recordFileUpload(ctx, payload)
		if recordErr != nil {
			ctx.Error(recordErr)
			return
		}
		resp := gin.H{
			"file_id":    fileID,
			"key":        key,
			"put_url":    c.cos.presignPutURL(key, expiresIn),
			"expires_in": expiresIn,
			"expire_at":  time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
		}
		if staticURL != "" {
			resp["static_url"] = staticURL
		}
		response.Success(ctx, resp)
		return
	}
	userID, _ := middleware.GetCurrentUserID(ctx)
	result, err := c.uploadSvc.GenerateAdminPresign(
		ctx.Request.Context(),
		userID,
		ctx.Query("filename"),
		ctx.DefaultQuery("usage", "protected"),
		ctx.Query("expires_in"),
	)
	if err != nil {
		ctx.Error(err)
		return
	}
	resp := gin.H{
		"file_id":    result.FileID,
		"key":        result.Key,
		"put_url":    result.PutURL,
		"expires_in": result.ExpiresIn,
		"expire_at":  result.ExpireAt,
	}
	if result.StaticURL != "" {
		resp["static_url"] = result.StaticURL
	}
	response.Success(ctx, resp)
}

// GenerateCourseVideoPresignURL handles GET /upload/course/video/presign.
func (c *UploadController) GenerateCourseVideoPresignURL(ctx *gin.Context) {
	if !isAdminUser(ctx) {
		ctx.Error(apperrors.NewForbidden("仅管理员可上传课程视频", nil))
		return
	}
	c.generatePresignUploadURL(ctx, "course-video", ".mp4", nil, "")
}

// GenerateBannerMediaPresignURL handles GET /upload/banner/media/presign.
func (c *UploadController) GenerateBannerMediaPresignURL(ctx *gin.Context) {
	if !isAdminUser(ctx) {
		ctx.Error(apperrors.NewForbidden("仅管理员可上传轮播素材", nil))
		return
	}
	if c.uploadSvc == nil {
		ctx.Error(apperrors.NewBadRequest("上传服务未初始化", nil))
		return
	}
	userID, _ := middleware.GetCurrentUserID(ctx)
	result, err := c.uploadSvc.GenerateProtectedBusinessPresign(
		ctx.Request.Context(),
		userID,
		ctx.Query("filename"),
		"banner_media",
		ctx.Query("expires_in"),
		[]string{"image", "video"},
	)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, gin.H{
		"file_id":    result.FileID,
		"key":        result.Key,
		"put_url":    result.PutURL,
		"expires_in": result.ExpiresIn,
		"expire_at":  result.ExpireAt,
	})
}

// GenerateArticleAttachmentPresignURL handles GET /upload/article/attachment/presign.
func (c *UploadController) GenerateArticleAttachmentPresignURL(ctx *gin.Context) {
	if !isAdminUser(ctx) {
		ctx.Error(apperrors.NewForbidden("仅管理员可上传文章附件", nil))
		return
	}
	c.generatePresignUploadURL(ctx, "article-attachment", "", nil, "附件扩展名不支持")
}

// GenerateCourseAttachmentPresignURL handles GET /upload/course/attachment/presign.
func (c *UploadController) GenerateCourseAttachmentPresignURL(ctx *gin.Context) {
	if !isAdminUser(ctx) {
		ctx.Error(apperrors.NewForbidden("仅管理员可上传课程附件", nil))
		return
	}
	c.generatePresignUploadURL(ctx, "course-attachment", "", courseAttachmentTypePattern, "课程附件扩展名不支持")
}

// GenerateCourseUnitAttachmentPresignURL handles GET /upload/course/unit/attachment/presign.
func (c *UploadController) GenerateCourseUnitAttachmentPresignURL(ctx *gin.Context) {
	if !isAdminUser(ctx) {
		ctx.Error(apperrors.NewForbidden("仅管理员可上传课程单元附件", nil))
		return
	}
	c.generatePresignUploadURL(ctx, "course-unit-attachment", "", courseAttachmentTypePattern, "课程单元附件扩展名不支持")
}

// GenerateCourseVideoDownloadURL handles GET /download/course/video.
func (c *UploadController) GenerateCourseVideoDownloadURL(ctx *gin.Context) {
	if c.uploadSvc == nil {
		c.generateBusinessTemporaryDownloadURL(ctx, "course_video", "video")
		return
	}
	c.generateServiceBusinessDownloadURL(ctx, "video")
}

// GenerateArticleAttachmentDownloadURL handles GET /download/article/attachment.
func (c *UploadController) GenerateArticleAttachmentDownloadURL(ctx *gin.Context) {
	if err := c.checkArticleAttachmentDownloadAccess(ctx); err != nil {
		ctx.Error(err)
		return
	}
	if c.uploadSvc == nil {
		c.generateBusinessTemporaryDownloadURL(ctx, "article_attachment", "attachment")
		return
	}
	c.generateServiceBusinessDownloadURL(ctx, "attachment")
}

// GenerateCourseAttachmentDownloadURL handles GET /download/course/attachment.
func (c *UploadController) GenerateCourseAttachmentDownloadURL(ctx *gin.Context) {
	if err := c.checkCourseAttachmentDownloadAccess(ctx); err != nil {
		ctx.Error(err)
		return
	}
	if c.uploadSvc == nil {
		c.generateBusinessTemporaryDownloadURL(ctx, "course_attachment", "attachment")
		return
	}
	c.generateServiceBusinessDownloadURL(ctx, "attachment")
}

// GenerateCourseUnitAttachmentDownloadURL handles GET /download/course/unit/attachment/:file_id.
func (c *UploadController) GenerateCourseUnitAttachmentDownloadURL(ctx *gin.Context) {
	if err := c.checkCourseUnitAttachmentDownloadAccess(ctx); err != nil {
		ctx.Error(err)
		return
	}
	if c.uploadSvc == nil {
		c.generateBusinessTemporaryDownloadURL(ctx, "course_unit_attachment", "attachment")
		return
	}
	c.generateServiceBusinessDownloadURL(ctx, "attachment")
}

// GenerateBannerMediaDownloadURL handles GET /download/banner/media/:file_id.
func (c *UploadController) GenerateBannerMediaDownloadURL(ctx *gin.Context) {
	if c.uploadSvc == nil {
		ctx.Error(apperrors.NewBadRequest("上传服务未初始化", nil))
		return
	}
	c.generateServiceBusinessDownloadURL(ctx, "image", "video")
}

// GenerateStaticMaterialURL handles GET /download/static/:file_id.
func (c *UploadController) GenerateStaticMaterialURL(ctx *gin.Context) {
	if c.uploadSvc == nil {
		file, payload, err := c.loadFileRecord(ctx)
		if err != nil {
			ctx.Error(err)
			return
		}
		if payload.Protected || payload.Usage != "embedded" || (payload.Category != "image" && payload.Category != "video") {
			ctx.Error(apperrors.NewForbidden("该文件不支持静态访问", nil))
			return
		}
		if c.cos == nil {
			ctx.Error(apperrors.NewBadRequest("当前存储不支持静态链接校验", nil))
			return
		}
		ok, checkErr := c.cos.isStaticMediaObject(ctx.Request.Context(), payload.Key)
		if checkErr != nil {
			ctx.Error(apperrors.NewInternal("校验文件类型失败", checkErr))
			return
		}
		if !ok {
			ctx.Error(apperrors.NewForbidden("静态访问仅支持图片和视频文件", nil))
			return
		}
		response.Success(ctx, gin.H{
			"file_id":    file.ID,
			"static_url": c.cos.objectURL(payload.Key),
			"category":   payload.Category,
		})
		return
	}
	fileID, err := strconv.ParseUint(strings.TrimSpace(ctx.Param("file_id")), 10, 64)
	if err != nil || fileID == 0 {
		ctx.Error(apperrors.NewBadRequest("file_id不合法", err))
		return
	}
	result, err := c.uploadSvc.GenerateStaticURL(ctx.Request.Context(), fileID)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, gin.H{
		"file_id":    result.FileID,
		"static_url": result.StaticURL,
		"category":   result.Category,
	})
}

func (c *UploadController) generateServiceBusinessDownloadURL(ctx *gin.Context, expectedCategories ...string) {
	if c.uploadSvc == nil {
		ctx.Error(apperrors.NewBadRequest("上传服务未初始化", nil))
		return
	}
	fileID, err := strconv.ParseUint(strings.TrimSpace(ctx.Param("file_id")), 10, 64)
	if err != nil || fileID == 0 {
		ctx.Error(apperrors.NewBadRequest("file_id不合法", err))
		return
	}
	result, svcErr := c.uploadSvc.GenerateBusinessDownload(ctx.Request.Context(), fileID, expectedCategories, ctx.Query("expires_in"))
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, gin.H{
		"file_id":    result.FileID,
		"download":   result.Download,
		"expires_in": result.ExpiresIn,
		"expire_at":  result.ExpireAt,
	})
}

func (c *UploadController) checkArticleAttachmentDownloadAccess(ctx *gin.Context) error {
	if c.access == nil || c.articleAttachRepo == nil || c.articleRepo == nil {
		return nil
	}
	fileID, err := strconv.ParseUint(strings.TrimSpace(ctx.Param("file_id")), 10, 64)
	if err != nil || fileID == 0 {
		return apperrors.NewBadRequest("file_id不合法", err)
	}
	row, err := c.articleAttachRepo.GetByFileID(ctx, fileID)
	if err != nil {
		return err
	}
	if row == nil {
		return apperrors.NewNotFound("附件不存在", nil)
	}
	article, err := c.articleRepo.GetByID(ctx, row.ArticleID)
	if err != nil {
		return err
	}
	if article == nil {
		return apperrors.NewNotFound("文章不存在", nil)
	}
	var uid *uint64
	if userID, ok := middleware.GetCurrentUserID(ctx); ok && userID > 0 {
		uid = &userID
	}
	allowed, accessErr := c.access.canAccess(ctx, 1, article.ID, uid, &article.AuthorID)
	if accessErr != nil {
		return accessErr
	}
	if !allowed {
		return apperrors.NewForbidden("无权限访问该附件", nil)
	}
	allowed, accessErr = c.access.canAccess(ctx, 4, row.ID, uid, &article.AuthorID)
	if accessErr != nil {
		return accessErr
	}
	if !allowed {
		return apperrors.NewForbidden("无权限访问该附件", nil)
	}
	return nil
}

func (c *UploadController) checkCourseAttachmentDownloadAccess(ctx *gin.Context) error {
	if c.access == nil || c.courseAttachRepo == nil || c.courseRepo == nil {
		return nil
	}
	fileID, err := strconv.ParseUint(strings.TrimSpace(ctx.Param("file_id")), 10, 64)
	if err != nil || fileID == 0 {
		return apperrors.NewBadRequest("file_id不合法", err)
	}
	row, err := c.courseAttachRepo.GetByFileID(ctx, fileID)
	if err != nil {
		return err
	}
	if row == nil {
		return apperrors.NewNotFound("附件不存在", nil)
	}
	course, err := c.courseRepo.GetByID(ctx, row.CourseID)
	if err != nil {
		return err
	}
	if course == nil {
		return apperrors.NewNotFound("课程不存在", nil)
	}
	var uid *uint64
	if userID, ok := middleware.GetCurrentUserID(ctx); ok && userID > 0 {
		uid = &userID
	}
	allowed, accessErr := c.access.canAccess(ctx, 2, course.ID, uid, &course.AuthorID)
	if accessErr != nil {
		return accessErr
	}
	if !allowed {
		return apperrors.NewForbidden("无权限访问该附件", nil)
	}
	allowed, accessErr = c.access.canAccess(ctx, 5, row.ID, uid, &course.AuthorID)
	if accessErr != nil {
		return accessErr
	}
	if !allowed {
		return apperrors.NewForbidden("无权限访问该附件", nil)
	}
	return nil
}

func (c *UploadController) checkCourseUnitAttachmentDownloadAccess(ctx *gin.Context) error {
	if c.access == nil || c.courseUnitAttachRepo == nil || c.courseUnitRepo == nil || c.courseRepo == nil {
		return nil
	}
	fileID, err := strconv.ParseUint(strings.TrimSpace(ctx.Param("file_id")), 10, 64)
	if err != nil || fileID == 0 {
		return apperrors.NewBadRequest("file_id不合法", err)
	}
	row, err := c.courseUnitAttachRepo.GetByFileID(ctx, fileID)
	if err != nil {
		return err
	}
	if row == nil {
		return apperrors.NewNotFound("附件不存在", nil)
	}
	unit, err := c.courseUnitRepo.GetByID(ctx, row.UnitID)
	if err != nil {
		return err
	}
	if unit == nil {
		return apperrors.NewNotFound("课程单元不存在", nil)
	}
	course, err := c.courseRepo.GetByID(ctx, unit.CourseID)
	if err != nil {
		return err
	}
	if course == nil {
		return apperrors.NewNotFound("课程不存在", nil)
	}
	var uid *uint64
	if userID, ok := middleware.GetCurrentUserID(ctx); ok && userID > 0 {
		uid = &userID
	}
	allowed, accessErr := c.access.canAccess(ctx, 6, unit.ID, uid, &course.AuthorID)
	if accessErr != nil {
		return accessErr
	}
	if !allowed {
		return apperrors.NewForbidden("无权限访问该单元附件", nil)
	}
	allowed, accessErr = c.access.canAccess(ctx, 7, row.ID, uid, &course.AuthorID)
	if accessErr != nil {
		return accessErr
	}
	if !allowed {
		return apperrors.NewForbidden("无权限访问该单元附件", nil)
	}
	return nil
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

func (c *UploadController) generatePresignUploadURL(ctx *gin.Context, businessType, forceExt string, attachmentPattern *regexp.Regexp, extErrMsg string) {
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
	if forceExt == "" {
		pattern := attachmentTypePattern
		if attachmentPattern != nil {
			pattern = attachmentPattern
		}
		if !pattern.MatchString(ext) {
			ctx.Error(apperrors.NewBadRequest(extErrMsg, nil))
			return
		}
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

func (c *UploadController) generateBusinessTemporaryDownloadURL(ctx *gin.Context, business, expectedCategory string) {
	if c.cos == nil {
		ctx.Error(apperrors.NewBadRequest("当前存储不支持临时下载链接", nil))
		return
	}
	file, payload, err := c.loadFileRecord(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if payload.Usage != "protected" {
		ctx.Error(apperrors.NewForbidden("该文件无需临时链接下载", nil))
		return
	}
	if payload.Category != expectedCategory {
		ctx.Error(apperrors.NewForbidden("文件类型与业务不匹配", nil))
		return
	}

	expiresIn := 300
	if raw := strings.TrimSpace(ctx.Query("expires_in")); raw != "" {
		v, convErr := strconv.Atoi(raw)
		if convErr != nil || v < 60 || v > 3600 {
			ctx.Error(apperrors.NewBadRequest("expires_in必须在60-3600秒之间", convErr))
			return
		}
		expiresIn = v
	}
	response.Success(ctx, gin.H{
		"file_id":    file.ID,
		"business":   business,
		"download":   c.cos.presignGetURL(payload.Key, expiresIn),
		"expires_in": expiresIn,
		"expire_at":  time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
	})
}

func (c *UploadController) loadFileRecord(ctx *gin.Context) (*entity.AuditLog, *fileRecordPayload, error) {
	if c.auditRepo == nil {
		return nil, nil, apperrors.NewInternal("文件记录仓储未初始化", nil)
	}
	fileID, err := strconv.ParseUint(strings.TrimSpace(ctx.Param("file_id")), 10, 64)
	if err != nil || fileID == 0 {
		return nil, nil, apperrors.NewBadRequest("file_id不合法", err)
	}
	record, err := c.auditRepo.GetByID(ctx.Request.Context(), fileID)
	if err != nil {
		return nil, nil, err
	}
	if record == nil || record.Module != "file_upload" || record.Action != "file_asset" {
		return nil, nil, apperrors.NewNotFound("文件不存在", nil)
	}
	var payload fileRecordPayload
	if unmarshalErr := json.Unmarshal([]byte(record.RequestData), &payload); unmarshalErr != nil {
		return nil, nil, apperrors.NewInternal("文件记录数据异常", unmarshalErr)
	}
	if normalizeObjectKey(payload.Key) == "" {
		return nil, nil, apperrors.NewInternal("文件记录数据异常", nil)
	}
	return record, &payload, nil
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

func (c *UploadController) recordFileUpload(ctx *gin.Context, payload fileRecordPayload) (uint64, error) {
	if c.auditRepo == nil {
		return 0, nil
	}
	userID, _ := middleware.GetCurrentUserID(ctx)
	requestData, _ := json.Marshal(payload)
	logRecord := &entity.AuditLog{
		UserID:      userID,
		Action:      "file_asset",
		Module:      "file_upload",
		Description: "创建文件上传记录",
		IPAddress:   ctx.ClientIP(),
		UserAgent:   ctx.Request.UserAgent(),
		RequestData: string(requestData),
	}
	if err := c.auditRepo.Create(ctx.Request.Context(), logRecord); err != nil {
		return 0, apperrors.NewInternal("记录文件上传信息失败", err)
	}
	return logRecord.ID, nil
}

func (c *UploadController) saveFile(ctx context.Context, file io.Reader, header *multipart.FileHeader, key string) (string, error) {
	normalizedKey := normalizeObjectKey(key)
	if normalizedKey == "" {
		return "", apperrors.NewBadRequest("文件key不合法", nil)
	}

	if c.cos != nil {
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		url, err := c.cos.upload(ctx, normalizedKey, file, contentType)
		if err != nil {
			return "", apperrors.NewInternal("上传到对象存储失败", err)
		}
		return url, nil
	}

	savePath := filepath.Join(c.uploadDir, normalizedKey)
	relPath, err := filepath.Rel(c.uploadDir, savePath)
	if err != nil || relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
		return "", apperrors.NewBadRequest("文件key不合法", err)
	}
	if err := os.MkdirAll(filepath.Dir(savePath), 0o755); err != nil {
		return "", apperrors.NewInternal("创建上传目录失败", err)
	}
	if err := saveUploadedFile(header, savePath, 500*1024*1024); err != nil {
		return "", apperrors.NewInternal("保存文件失败", err)
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(c.baseURL, "/"), normalizedKey), nil
}

func saveUploadedFile(header *multipart.FileHeader, savePath string, maxBytes int64) error {
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

	limitReader := io.LimitReader(src, maxBytes+1)
	written, err := io.Copy(dst, limitReader)
	if err != nil {
		return err
	}
	if written > maxBytes {
		return fmt.Errorf("文件过大")
	}
	return nil
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
var courseAttachmentTypePattern = regexp.MustCompile(`^\.(pdf|doc|docx|xls|xlsx|ppt|pptx)$`)

func classifyFileCategory(ext string) string {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		return "image"
	case ".mp4":
		return "video"
	default:
		if attachmentTypePattern.MatchString(ext) {
			return "attachment"
		}
		return ""
	}
}

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

func (u *cosUploader) isStaticMediaObject(ctx context.Context, key string) (bool, error) {
	targetURL := fmt.Sprintf("%s/%s/%s", u.endpoint, url.PathEscape(u.bucket), escapeObjectKey(key))
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, targetURL, nil)
	if err != nil {
		return false, err
	}
	resp, err := u.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return false, fmt.Errorf("head object failed status=%d", resp.StatusCode)
	}
	contentType := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type")))
	if strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/") {
		return true, nil
	}
	return false, nil
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
	return ok && (userType == 2 || userType == 3)
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
