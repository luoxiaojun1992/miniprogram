package service

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/cosutil"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

type UploadFileService interface {
	GenerateAdminPresign(ctx context.Context, userID uint64, filename, usage, expiresInRaw string) (*AdminPresignResult, error)
	GenerateProtectedBusinessPresign(ctx context.Context, userID uint64, filename, business, expiresInRaw string, allowedCategories []string) (*AdminPresignResult, error)
	GenerateBusinessDownload(ctx context.Context, fileID uint64, allowedCategories []string, expiresInRaw string) (*BusinessDownloadResult, error)
	GenerateStaticURL(ctx context.Context, fileID uint64) (*StaticURLResult, error)
}

type AdminPresignResult struct {
	FileID    uint64 `json:"file_id"`
	Key       string `json:"key"`
	PutURL    string `json:"put_url"`
	ExpiresIn int    `json:"expires_in"`
	ExpireAt  int64  `json:"expire_at"`
	StaticURL string `json:"static_url,omitempty"`
}

type BusinessDownloadResult struct {
	FileID    uint64 `json:"file_id"`
	Download  string `json:"download"`
	ExpiresIn int    `json:"expires_in"`
	ExpireAt  int64  `json:"expire_at"`
}

type StaticURLResult struct {
	FileID    uint64 `json:"file_id"`
	StaticURL string `json:"static_url"`
	Category  string `json:"category"`
}

const (
	usageEmbedded  = "embedded"
	usageProtected = "protected"

	categoryImage      = "image"
	categoryVideo      = "video"
	categoryAttachment = "attachment"

	defaultAdminExpiresIn    = 900
	defaultDownloadExpiresIn = 300
	minExpiresIn             = 60
	maxExpiresIn             = 3600
)

type uploadFileService struct {
	fileRepo repository.FileRepository
	cos      *cosutil.Client
	log      *logrus.Logger
}

func NewUploadFileService(fileRepo repository.FileRepository, cosClient *cosutil.Client, log *logrus.Logger) UploadFileService {
	return &uploadFileService{
		fileRepo: fileRepo,
		cos:      cosClient,
		log:      log,
	}
}

func (s *uploadFileService) GenerateAdminPresign(ctx context.Context, userID uint64, filename, usage, expiresInRaw string) (*AdminPresignResult, error) {
	if err := s.validateCOS("当前存储不支持预签名上传"); err != nil {
		return nil, err
	}
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return nil, errors.NewBadRequest("filename不能为空", nil)
	}
	usage = strings.ToLower(strings.TrimSpace(usage))
	if usage == "" {
		usage = usageProtected
	}
	if usage != usageEmbedded && usage != usageProtected {
		return nil, errors.NewBadRequest("usage仅支持embedded或protected", nil)
	}
	ext := strings.ToLower(filepath.Ext(filename))
	category := classifyFileCategory(ext)
	if category == "" {
		return nil, errors.NewBadRequest("不支持的文件扩展名", nil)
	}
	if usage == usageEmbedded && category == categoryAttachment {
		return nil, errors.NewBadRequest("内嵌素材仅支持图片或视频", nil)
	}
	expiresIn, err := parseExpiresIn(expiresInRaw, defaultAdminExpiresIn)
	if err != nil {
		return nil, err
	}
	key := generateObjectKey(usage+"-"+category, ext)
	staticURL := ""
	if usage == usageEmbedded {
		staticURL = s.cos.ObjectURL(key)
	}
	file := &entity.File{
		Key:       key,
		Filename:  filename,
		Usage:     usage,
		Category:  category,
		Business:  usage + "_" + category,
		StaticURL: staticURL,
		CreatedBy: userID,
	}
	if createErr := s.fileRepo.Create(ctx, file); createErr != nil {
		return nil, createErr
	}
	return &AdminPresignResult{
		FileID:    file.ID,
		Key:       key,
		PutURL:    s.cos.PresignPutURL(key, expiresIn),
		ExpiresIn: expiresIn,
		ExpireAt:  time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
		StaticURL: staticURL,
	}, nil
}

func (s *uploadFileService) GenerateProtectedBusinessPresign(ctx context.Context, userID uint64, filename, business, expiresInRaw string, allowedCategories []string) (*AdminPresignResult, error) {
	if err := s.validateCOS("当前存储不支持预签名上传"); err != nil {
		return nil, err
	}
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return nil, errors.NewBadRequest("filename不能为空", nil)
	}
	ext := strings.ToLower(filepath.Ext(filename))
	category := classifyFileCategory(ext)
	if category == "" {
		return nil, errors.NewBadRequest("不支持的文件扩展名", nil)
	}
	if !containsCategory(allowedCategories, category) {
		return nil, errors.NewBadRequest("该业务仅支持图片或视频文件", nil)
	}
	expiresIn, err := parseExpiresIn(expiresInRaw, defaultAdminExpiresIn)
	if err != nil {
		return nil, err
	}
	key := generateObjectKey(usageProtected+"-"+category, ext)
	file := &entity.File{
		Key:       key,
		Filename:  filename,
		Usage:     usageProtected,
		Category:  category,
		Business:  business,
		CreatedBy: userID,
	}
	if createErr := s.fileRepo.Create(ctx, file); createErr != nil {
		return nil, createErr
	}
	return &AdminPresignResult{
		FileID:    file.ID,
		Key:       key,
		PutURL:    s.cos.PresignPutURL(key, expiresIn),
		ExpiresIn: expiresIn,
		ExpireAt:  time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
	}, nil
}

func (s *uploadFileService) GenerateBusinessDownload(ctx context.Context, fileID uint64, allowedCategories []string, expiresInRaw string) (*BusinessDownloadResult, error) {
	if err := s.validateCOS("当前存储不支持临时下载链接"); err != nil {
		return nil, err
	}
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, errors.NewNotFound("文件不存在", nil)
	}
	if file.Usage != usageProtected {
		return nil, errors.NewForbidden("该文件无需临时链接下载", nil)
	}
	if !containsCategory(allowedCategories, file.Category) {
		return nil, errors.NewForbidden("文件类型与业务不匹配", nil)
	}
	contentType, contentTypeErr := s.cos.ObjectContentType(ctx, file.Key)
	if contentTypeErr != nil {
		return nil, errors.NewInternal("校验文件类型失败", contentTypeErr)
	}
	if !contentTypeMatchesCategory(file.Category, contentType, file.Filename, file.Key) {
		return nil, errors.NewForbidden(fmt.Sprintf("文件MIME类型(%s)与业务类别(%s)不匹配", contentType, file.Category), nil)
	}
	expiresIn, parseErr := parseExpiresIn(expiresInRaw, defaultDownloadExpiresIn)
	if parseErr != nil {
		return nil, parseErr
	}
	return &BusinessDownloadResult{
		FileID:    file.ID,
		Download:  s.cos.PresignGetURL(file.Key, expiresIn),
		ExpiresIn: expiresIn,
		ExpireAt:  time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
	}, nil
}

func (s *uploadFileService) GenerateStaticURL(ctx context.Context, fileID uint64) (*StaticURLResult, error) {
	if err := s.validateCOS("当前存储不支持静态链接校验"); err != nil {
		return nil, err
	}
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, errors.NewNotFound("文件不存在", nil)
	}
	if file.Usage != usageEmbedded || (file.Category != categoryImage && file.Category != categoryVideo) {
		return nil, errors.NewForbidden("该文件不支持静态访问", nil)
	}
	ok, checkErr := s.cos.IsStaticMediaObject(ctx, file.Key)
	if checkErr != nil {
		return nil, errors.NewInternal("校验文件类型失败", checkErr)
	}
	if !ok {
		return nil, errors.NewForbidden("静态访问仅支持图片和视频文件", nil)
	}
	return &StaticURLResult{
		FileID:    file.ID,
		StaticURL: s.cos.ObjectURL(file.Key),
		Category:  file.Category,
	}, nil
}

// Validation and helper functions.
func (s *uploadFileService) validateCOS(errMsg string) error {
	if s.cos == nil {
		return errors.NewBadRequest(errMsg, nil)
	}
	return nil
}

func containsCategory(allowed []string, category string) bool {
	if len(allowed) == 0 {
		return false
	}
	for _, item := range allowed {
		if strings.EqualFold(strings.TrimSpace(item), category) {
			return true
		}
	}
	return false
}

var fileAttachmentTypePattern = regexp.MustCompile(`^\.(pdf|doc|docx|xls|xlsx|ppt|pptx|zip|rar|7z|txt)$`)

func classifyFileCategory(ext string) string {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		return categoryImage
	case ".mp4":
		return categoryVideo
	default:
		if fileAttachmentTypePattern.MatchString(ext) {
			return categoryAttachment
		}
		return ""
	}
}

func parseExpiresIn(raw string, defaultValue int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v < minExpiresIn || v > maxExpiresIn {
		return 0, errors.NewBadRequest(fmt.Sprintf("expires_in必须在%d-%d秒之间", minExpiresIn, maxExpiresIn), err)
	}
	return v, nil
}

func generateObjectKey(prefix, ext string) string {
	return fmt.Sprintf("%s/%s/%s%s", prefix, time.Now().Format("20060102"), uuid.NewString(), ext)
}

func contentTypeMatchesCategory(category, contentType, filename, key string) bool {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if ct == "" {
		return false
	}
	if idx := strings.Index(ct, ";"); idx >= 0 {
		ct = strings.TrimSpace(ct[:idx])
	}
	switch category {
	case categoryImage:
		return strings.HasPrefix(ct, "image/")
	case categoryVideo:
		return strings.HasPrefix(ct, "video/")
	case categoryAttachment:
		ext := resolveFileExtension(filename, key)
		return attachmentContentTypeAllowed(ext, ct)
	default:
		return false
	}
}

func attachmentContentTypeAllowed(extension, contentType string) bool {
	if contentType == "application/octet-stream" {
		return true
	}
	switch extension {
	case ".pdf":
		return contentType == "application/pdf"
	case ".doc":
		return contentType == "application/msword"
	case ".docx":
		return contentType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return contentType == "application/vnd.ms-excel"
	case ".xlsx":
		return contentType == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return contentType == "application/vnd.ms-powerpoint"
	case ".pptx":
		return contentType == "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".zip":
		return contentType == "application/zip" || contentType == "application/x-zip-compressed"
	case ".rar":
		return contentType == "application/vnd.rar" || contentType == "application/x-rar-compressed"
	case ".7z":
		return contentType == "application/x-7z-compressed"
	case ".txt":
		return contentType == "text/plain"
	default:
		return false
	}
}

func resolveFileExtension(filename, key string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != "" {
		return ext
	}
	keyExt := filepath.Ext(key)
	return strings.ToLower(keyExt)
}
