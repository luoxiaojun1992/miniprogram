package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestUploadCtrl_UploadImage_COS_OK(t *testing.T) {
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
	r.POST("/upload/image", ctrl.UploadImage)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "test.png")
	_, _ = part.Write([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'})
	_ = writer.WriteField("type", "article")
	_ = writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/upload/image", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, gotPath, "/miniapp-test/article/")

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data, _ := resp["data"].(map[string]interface{})
	url, _ := data["url"].(string)
	assert.Contains(t, url, "/miniapp-test/article/")
}

func TestUploadCtrl_GeneratePresignURL_COS_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New())
	r := gin.New()
	r.GET("/upload/presign", ctrl.GeneratePresignURL)

	req, _ := http.NewRequest(http.MethodGet, "/upload/presign?filename=video.mp4&expires_in=600", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data, _ := resp["data"].(map[string]interface{})
	assert.Contains(t, data["put_url"], "http://cos:9000/miniapp-test/video/")
	assert.Contains(t, data["url"], "http://cos:9000/miniapp-test/video/")
}

func TestUploadCtrl_GeneratePresignURL_InvalidFilename(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New())
	r := gin.New()
	r.GET("/upload/presign", ctrl.GeneratePresignURL)

	req, _ := http.NewRequest(http.MethodGet, "/upload/presign?filename=video.avi", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
