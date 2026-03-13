package controller

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/response"
)

// UploadController handles file upload requests.
type UploadController struct {
	uploadDir string
	baseURL   string
	log       *logrus.Logger
}

// NewUploadController creates a new UploadController.
func NewUploadController(uploadDir, baseURL string, log *logrus.Logger) *UploadController {
	return &UploadController{uploadDir: uploadDir, baseURL: baseURL, log: log}
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

	imgType := ctx.PostForm("type")
	if imgType == "" {
		imgType = "article"
	}
	key := fmt.Sprintf("%s/%d%s", imgType, time.Now().UnixNano(), ext)
	savePath := filepath.Join(c.uploadDir, key)

	if err := ctx.SaveUploadedFile(header, savePath); err != nil {
		ctx.Error(apperrors.NewInternal("保存文件失败", err))
		return
	}

	url := fmt.Sprintf("%s/%s", strings.TrimRight(c.baseURL, "/"), key)
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

	key := fmt.Sprintf("video/%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(c.uploadDir, key)

	if err := ctx.SaveUploadedFile(header, savePath); err != nil {
		ctx.Error(apperrors.NewInternal("保存文件失败", err))
		return
	}

	url := fmt.Sprintf("%s/%s", strings.TrimRight(c.baseURL, "/"), key)
	response.Success(ctx, gin.H{"url": url, "duration": 0, "cover_url": ""})
}
