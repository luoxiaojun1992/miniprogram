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
}

func (r *testAuditRepo) List(_ context.Context, _, _ int, _, _ string, _, _ *string) ([]*entity.AuditLog, int64, error) {
	return nil, 0, nil
}

func (r *testAuditRepo) Create(_ context.Context, log *entity.AuditLog) error {
	r.created = append(r.created, log)
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

func TestUploadCtrl_GenerateCourseVideoPresignURL_COS_OK(t *testing.T) {
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
	r.GET("/upload/course/video/presign", ctrl.GenerateCourseVideoPresignURL)

	req, _ := http.NewRequest(http.MethodGet, "/upload/course/video/presign?filename=video.mp4&expires_in=600", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data, _ := resp["data"].(map[string]interface{})
	assert.Contains(t, data["put_url"], "http://cos:9000/miniapp-test/course-video/")
	assert.Contains(t, data["url"], "http://cos:9000/miniapp-test/course-video/")
	assert.Len(t, auditRepo.created, 1)
	assert.Contains(t, auditRepo.created[0].RequestData, "course-video")
}

func TestUploadCtrl_GenerateCourseVideoPresignURL_ForbiddenForFrontUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New())
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint64(1))
		ctx.Set("user_type", int8(1))
		ctx.Next()
	})
	r.GET("/upload/course/video/presign", ctrl.GenerateCourseVideoPresignURL)

	req, _ := http.NewRequest(http.MethodGet, "/upload/course/video/presign?filename=video.mp4", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUploadCtrl_GenerateCourseVideoDownloadURL_COS_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New())
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.GET("/download/course/video", ctrl.GenerateCourseVideoDownloadURL)

	req, _ := http.NewRequest(http.MethodGet, "/download/course/video?key=course-video/20260317/test.mp4", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data, _ := resp["data"].(map[string]interface{})
	assert.Contains(t, data["download"], "http://cos:9000/miniapp-test/course-video/20260317/test.mp4")
}
