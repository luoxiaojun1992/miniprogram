package controller

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
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

	imgType := ctx.PostForm("type")
	if imgType == "" {
		imgType = "article"
	}
	key := fmt.Sprintf("%s/%d%s", imgType, time.Now().UnixNano(), ext)
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

	key := fmt.Sprintf("video/%d%s", time.Now().UnixNano(), ext)
	url, err := c.saveFile(ctx.Request.Context(), file, header, key)
	if err != nil {
		ctx.Error(err)
		return
	}

	response.Success(ctx, gin.H{"url": url, "duration": 0, "cover_url": ""})
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
	targetURL := fmt.Sprintf("%s/%s/%s", u.endpoint, url.PathEscape(u.bucket), escapedKey)
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
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(u.publicBaseURL, "/"), url.PathEscape(u.bucket), escapedKey), nil
}
