package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
)

type testAuditRepo struct {
	created []*entity.AuditLog
	byID    map[uint64]*entity.AuditLog
}

func (r *testAuditRepo) GetByID(_ context.Context, id uint64) (*entity.AuditLog, error) {
	if r.byID == nil {
		return nil, nil
	}
	return r.byID[id], nil
}

func (r *testAuditRepo) List(_ context.Context, _, _ int, _, _ string, _, _ *string) ([]*entity.AuditLog, int64, error) {
	return nil, 0, nil
}

func (r *testAuditRepo) Create(_ context.Context, log *entity.AuditLog) error {
	log.ID = uint64(len(r.created) + 1)
	r.created = append(r.created, log)
	if r.byID == nil {
		r.byID = map[uint64]*entity.AuditLog{}
	}
	r.byID[log.ID] = log
	return nil
}

func TestUploadCtrl_UploadAvatar_COS_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotPath string
	mockCOS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockCOS.Close()

	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", mockCOS.URL, mockCOS.URL, "miniapp-test", logrus.New())
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.POST("/upload/avatar", ctrl.UploadAvatar)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "test.png")
	_, _ = part.Write([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'})
	_ = writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/upload/avatar", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, gotPath, "/miniapp-test/avatar/")

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data, _ := resp["data"].(map[string]interface{})
	url, _ := data["url"].(string)
	assert.Contains(t, url, "/miniapp-test/avatar/")
}

func TestUploadCtrl_GenerateAdminUploadPresignURL_EmbeddedImage_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auditRepo := &testAuditRepo{}
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New()).WithAuditRepo(auditRepo)
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint64(99))
		ctx.Set("user_type", int8(2))
		ctx.Next()
	})
	r.GET("/admin/upload/files/presign", ctrl.GenerateAdminUploadPresignURL)

	req, _ := http.NewRequest(http.MethodGet, "/admin/upload/files/presign?filename=embed.png&usage=embedded&expires_in=600", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data, _ := resp["data"].(map[string]interface{})
	assert.NotZero(t, data["file_id"])
	assert.Contains(t, data["put_url"], "http://cos:9000/miniapp-test/embedded-image/")
	assert.Contains(t, data["static_url"], "http://cos:9000/miniapp-test/embedded-image/")
	assert.Len(t, auditRepo.created, 1)
	assert.Contains(t, auditRepo.created[0].RequestData, "\"usage\":\"embedded\"")
}

func TestUploadCtrl_GenerateAdminUploadPresignURL_ForbiddenForFrontUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New())
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint64(1))
		ctx.Set("user_type", int8(1))
		ctx.Next()
	})
	r.GET("/admin/upload/files/presign", ctrl.GenerateAdminUploadPresignURL)

	req, _ := http.NewRequest(http.MethodGet, "/admin/upload/files/presign?filename=video.mp4", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUploadCtrl_GenerateStaticMaterialURL_ValidatesCosContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCOS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Content-Type", "application/pdf")
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockCOS.Close()
	auditRepo := &testAuditRepo{
		byID: map[uint64]*entity.AuditLog{
			1: {
				ID:     1,
				Action: "file_asset",
				Module: "file_upload",
				RequestData: `{"key":"embedded-image/20260317/a.png","usage":"embedded","category":"image","protected":false}`,
			},
		},
	}
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", mockCOS.URL, mockCOS.URL, "miniapp-test", logrus.New()).WithAuditRepo(auditRepo)
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.GET("/download/static/:file_id", ctrl.GenerateStaticMaterialURL)

	req, _ := http.NewRequest(http.MethodGet, "/download/static/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUploadCtrl_GenerateCourseVideoDownloadURL_ByFileID_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auditRepo := &testAuditRepo{
		byID: map[uint64]*entity.AuditLog{
			11: {
				ID:     11,
				Action: "file_asset",
				Module: "file_upload",
				RequestData: `{"key":"protected-video/20260317/test.mp4","usage":"protected","category":"video","protected":true}`,
			},
		},
	}
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New()).WithAuditRepo(auditRepo)
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.GET("/download/course/video/:file_id", ctrl.GenerateCourseVideoDownloadURL)

	req, _ := http.NewRequest(http.MethodGet, "/download/course/video/11", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data, _ := resp["data"].(map[string]interface{})
	assert.Contains(t, data["download"], "http://cos:9000/miniapp-test/protected-video/20260317/test.mp4")
}

func TestUploadCtrl_GenerateCourseAttachmentPresignURL_RejectsZip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New())
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint64(99))
		ctx.Set("user_type", int8(2))
		ctx.Next()
	})
	r.GET("/upload/course/attachment/presign", ctrl.GenerateCourseAttachmentPresignURL)

	req, _ := http.NewRequest(http.MethodGet, "/upload/course/attachment/presign?filename=course.zip", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadCtrl_GenerateArticleAttachmentPresignURL_AllowsZip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New())
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint64(99))
		ctx.Set("user_type", int8(2))
		ctx.Next()
	})
	r.GET("/upload/article/attachment/presign", ctrl.GenerateArticleAttachmentPresignURL)

	req, _ := http.NewRequest(http.MethodGet, "/upload/article/attachment/presign?filename=article.zip", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
