package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/service"
)

type mockUploadFileService struct {
	generateAdminPresignFn            func(ctx context.Context, userID uint64, filename, usage, expiresInRaw string) (*service.AdminPresignResult, error)
	generateProtectedBusinessPresign  func(ctx context.Context, userID uint64, filename, business, expiresInRaw string, allowedCategories []string) (*service.AdminPresignResult, error)
	generateBusinessDownloadFn        func(ctx context.Context, fileID uint64, allowedCategories []string, expiresInRaw string) (*service.BusinessDownloadResult, error)
	generateStaticURLFn               func(ctx context.Context, fileID uint64) (*service.StaticURLResult, error)
}

func (m *mockUploadFileService) GenerateAdminPresign(ctx context.Context, userID uint64, filename, usage, expiresInRaw string) (*service.AdminPresignResult, error) {
	if m.generateAdminPresignFn != nil {
		return m.generateAdminPresignFn(ctx, userID, filename, usage, expiresInRaw)
	}
	return &service.AdminPresignResult{}, nil
}

func (m *mockUploadFileService) GenerateProtectedBusinessPresign(ctx context.Context, userID uint64, filename, business, expiresInRaw string, allowedCategories []string) (*service.AdminPresignResult, error) {
	if m.generateProtectedBusinessPresign != nil {
		return m.generateProtectedBusinessPresign(ctx, userID, filename, business, expiresInRaw, allowedCategories)
	}
	return &service.AdminPresignResult{}, nil
}

func (m *mockUploadFileService) GenerateBusinessDownload(ctx context.Context, fileID uint64, allowedCategories []string, expiresInRaw string) (*service.BusinessDownloadResult, error) {
	if m.generateBusinessDownloadFn != nil {
		return m.generateBusinessDownloadFn(ctx, fileID, allowedCategories, expiresInRaw)
	}
	return &service.BusinessDownloadResult{}, nil
}

func (m *mockUploadFileService) GenerateStaticURL(ctx context.Context, fileID uint64) (*service.StaticURLResult, error) {
	if m.generateStaticURLFn != nil {
		return m.generateStaticURLFn(ctx, fileID)
	}
	return &service.StaticURLResult{}, nil
}

type errAuditRepo struct {
	getErr    error
	createErr error
}

func (r *errAuditRepo) GetByID(_ context.Context, _ uint64) (*entity.AuditLog, error) {
	return nil, r.getErr
}

func (r *errAuditRepo) List(_ context.Context, _, _ int, _, _ string, _, _ *string) ([]*entity.AuditLog, int64, error) {
	return nil, 0, nil
}

func (r *errAuditRepo) Create(_ context.Context, _ *entity.AuditLog) error {
	return r.createErr
}

func withAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("user_id", uint64(99))
		ctx.Set("user_type", int8(2))
		ctx.Next()
	}
}

func withFrontUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("user_id", uint64(1))
		ctx.Set("user_type", int8(1))
		ctx.Next()
	}
}

func makePNGMultipartBody(t *testing.T, field, filename string) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile(field, filename)
	require.NoError(t, err)
	_, err = part.Write([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'})
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return &body, w.FormDataContentType()
}

func TestUploadCtrl_WithUploadServiceAndServiceDownloadWrappers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	calls := map[uint64][]string{}
	svc := &mockUploadFileService{
		generateBusinessDownloadFn: func(_ context.Context, fileID uint64, allowedCategories []string, expiresInRaw string) (*service.BusinessDownloadResult, error) {
			calls[fileID] = append([]string(nil), allowedCategories...)
			return &service.BusinessDownloadResult{
				FileID:    fileID,
				Download:  "http://cos:9000/miniapp-test/protected/x",
				ExpiresIn: 300,
				ExpireAt:  1,
			}, nil
		},
	}
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New()).WithUploadService(svc)
	require.NotNil(t, ctrl.uploadSvc)

	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.GET("/download/course/video/:file_id", ctrl.GenerateCourseVideoDownloadURL)
	r.GET("/download/article/attachment/:file_id", ctrl.GenerateArticleAttachmentDownloadURL)
	r.GET("/download/course/attachment/:file_id", ctrl.GenerateCourseAttachmentDownloadURL)
	r.GET("/download/banner/media/:file_id", ctrl.GenerateBannerMediaDownloadURL)

	for _, p := range []string{
		"/download/course/video/1",
		"/download/article/attachment/2",
		"/download/course/attachment/3",
		"/download/banner/media/4",
	} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, p, nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, p)
	}

	assert.Equal(t, []string{"video"}, calls[1])
	assert.Equal(t, []string{"attachment"}, calls[2])
	assert.Equal(t, []string{"attachment"}, calls[3])
	assert.Equal(t, []string{"image", "video"}, calls[4])

	// Bad file_id path exercises generateServiceBusinessDownloadURL input validation branch.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/download/banner/media/not-number", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadCtrl_UploadArticleImage_ForbiddenAndAdminOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCOS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockCOS.Close()
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", mockCOS.URL, mockCOS.URL, "miniapp-test", logrus.New())

	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.Use(withFrontUser())
	r.POST("/upload/article/image", ctrl.UploadArticleImage)
	body, ctype := makePNGMultipartBody(t, "file", "article.png")
	req, _ := http.NewRequest(http.MethodPost, "/upload/article/image", body)
	req.Header.Set("Content-Type", ctype)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	r2 := gin.New()
	r2.Use(middleware.ErrorMiddleware(logrus.New()))
	r2.Use(withAdmin())
	r2.POST("/upload/article/image", ctrl.UploadArticleImage)
	body2, ctype2 := makePNGMultipartBody(t, "file", "article.png")
	req2, _ := http.NewRequest(http.MethodPost, "/upload/article/image", body2)
	req2.Header.Set("Content-Type", ctype2)
	w2 := httptest.NewRecorder()
	r2.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestUploadCtrl_GenerateCourseVideoPresignURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New())

	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.Use(withFrontUser())
	r.GET("/upload/course/video/presign", ctrl.GenerateCourseVideoPresignURL)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/upload/course/video/presign?filename=test.mp4", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	r2 := gin.New()
	r2.Use(middleware.ErrorMiddleware(logrus.New()))
	r2.Use(withAdmin())
	r2.GET("/upload/course/video/presign", ctrl.GenerateCourseVideoPresignURL)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/upload/course/video/presign?filename=test.mp4", nil)
	r2.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestUploadCtrl_GenerateBannerMediaPresignURL_ServicePath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrlNilSvc := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New())
	r0 := gin.New()
	r0.Use(middleware.ErrorMiddleware(logrus.New()))
	r0.Use(withAdmin())
	r0.GET("/upload/banner/media/presign", ctrlNilSvc.GenerateBannerMediaPresignURL)
	w0 := httptest.NewRecorder()
	req0, _ := http.NewRequest(http.MethodGet, "/upload/banner/media/presign?filename=b.png", nil)
	r0.ServeHTTP(w0, req0)
	assert.Equal(t, http.StatusBadRequest, w0.Code)

	called := false
	svc := &mockUploadFileService{
		generateProtectedBusinessPresign: func(_ context.Context, userID uint64, filename, business, expiresInRaw string, allowedCategories []string) (*service.AdminPresignResult, error) {
			called = true
			assert.Equal(t, uint64(99), userID)
			assert.Equal(t, "banner_media", business)
			assert.Equal(t, []string{"image", "video"}, allowedCategories)
			return &service.AdminPresignResult{
				FileID:    1,
				Key:       "protected-image/a.png",
				PutURL:    "http://cos:9000/miniapp-test/protected-image/a.png?expires_in=600",
				ExpiresIn: 600,
				ExpireAt:  1,
			}, nil
		},
	}
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New()).WithUploadService(svc)
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.Use(withAdmin())
	r.GET("/upload/banner/media/presign", ctrl.GenerateBannerMediaPresignURL)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/upload/banner/media/presign?filename=b.png", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, called)
}

func TestUploadCtrl_GeneratePresignURL_Branches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rNoCOS := gin.New()
	rNoCOS.Use(middleware.ErrorMiddleware(logrus.New()))
	rNoCOS.GET("/upload/presign", NewUploadController("/tmp/uploads_test", "http://localhost:8080", logrus.New()).GeneratePresignURL)
	wNoCOS := httptest.NewRecorder()
	reqNoCOS, _ := http.NewRequest(http.MethodGet, "/upload/presign?filename=a.mp4", nil)
	rNoCOS.ServeHTTP(wNoCOS, reqNoCOS)
	assert.Equal(t, http.StatusBadRequest, wNoCOS.Code)

	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New())
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.GET("/upload/presign", ctrl.GeneratePresignURL)

	wBad := httptest.NewRecorder()
	reqBad, _ := http.NewRequest(http.MethodGet, "/upload/presign?filename=a.png", nil)
	r.ServeHTTP(wBad, reqBad)
	assert.Equal(t, http.StatusBadRequest, wBad.Code)

	wOK := httptest.NewRecorder()
	reqOK, _ := http.NewRequest(http.MethodGet, "/upload/presign?filename=a.mp4&expires_in=600", nil)
	r.ServeHTTP(wOK, reqOK)
	assert.Equal(t, http.StatusOK, wOK.Code)
}

func TestUploadCtrl_GenerateStaticMaterialURL_ServicePath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New()).WithUploadService(&mockUploadFileService{
		generateStaticURLFn: func(_ context.Context, fileID uint64) (*service.StaticURLResult, error) {
			return &service.StaticURLResult{
				FileID:    fileID,
				StaticURL: "http://cos:9000/miniapp-test/embedded-image/a.png",
				Category:  "image",
			}, nil
		},
	})
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.GET("/download/static/:file_id", ctrl.GenerateStaticMaterialURL)

	wBad := httptest.NewRecorder()
	reqBad, _ := http.NewRequest(http.MethodGet, "/download/static/not-a-number", nil)
	r.ServeHTTP(wBad, reqBad)
	assert.Equal(t, http.StatusBadRequest, wBad.Code)

	wOK := httptest.NewRecorder()
	reqOK, _ := http.NewRequest(http.MethodGet, "/download/static/12", nil)
	r.ServeHTTP(wOK, reqOK)
	assert.Equal(t, http.StatusOK, wOK.Code)
}

func TestUploadCtrl_InternalHelpersAndRecordBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	c := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New())
	r.GET("/tmp-download", func(ctx *gin.Context) { c.generateTemporaryDownloadURL(ctx, "course-video") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/tmp-download?key=course-video/20260318/a.mp4&expires_in=300", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/tmp-download?key=avatar/a.png", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusForbidden, w2.Code)

	wNil := httptest.NewRecorder()
	ctxNil, _ := gin.CreateTestContext(wNil)
	ctxNil.Request, _ = http.NewRequest(http.MethodGet, "/x", nil)
	_, _, err := c.loadFileRecord(ctxNil)
	assert.Error(t, err) // audit repo nil

	c.auditRepo = &testAuditRepo{byID: map[uint64]*entity.AuditLog{
		1: {ID: 1, Action: "file_asset", Module: "file_upload", RequestData: "{"},
	}}
	w3 := httptest.NewRecorder()
	ctx2, _ := gin.CreateTestContext(w3)
	ctx2.Request, _ = http.NewRequest(http.MethodGet, "/x", nil)
	ctx2.Params = gin.Params{{Key: "file_id", Value: "1"}}
	_, _, err = c.loadFileRecord(ctx2)
	assert.Error(t, err)

	c.auditRepo = &errAuditRepo{createErr: errors.New("create failed")}
	err = c.recordUploadBusiness(ctx2, "course-video", "course-video/1.mp4", "1.mp4", 300)
	assert.Error(t, err)
	_, err = c.recordFileUpload(ctx2, fileRecordPayload{Key: "course-video/1.mp4", Usage: "protected", Category: "video"})
	assert.Error(t, err)

	_ = r // keep style parity
}

func TestUploadCtrl_HelperFunctions(t *testing.T) {
	assert.Equal(t, "image", classifyFileCategory(".png"))
	assert.Equal(t, "video", classifyFileCategory(".mp4"))
	assert.Equal(t, "attachment", classifyFileCategory(".pdf"))
	assert.Equal(t, "", classifyFileCategory(".exe"))

	assert.Equal(t, "avatar", sanitizeUploadType(" Avatar "))
	assert.Equal(t, "article", sanitizeUploadType("unknown"))

	assert.Equal(t, "a/b.png", normalizeObjectKey(" /a//b.png "))
	assert.Equal(t, "x", normalizeObjectKey("../x"))
	assert.Equal(t, "a%20b/c+d.png", escapeObjectKey("a b/c+d.png"))

	assert.NoError(t, validateImageMagic(newTestMultipartFile(t, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'}), ".png"))
	assert.NoError(t, validateImageMagic(newTestMultipartFile(t, []byte{'G', 'I', 'F', '8', '9', 'a'}), ".gif"))
	assert.NoError(t, validateImageMagic(newTestMultipartFile(t, []byte{0xFF, 0xD8, 0xAA}), ".jpg"))
	assert.Error(t, validateImageMagic(newTestMultipartFile(t, []byte("not-image")), ".png"))
}

func TestUploadCtrl_SaveFile_InvalidKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewUploadController("/tmp/uploads_test", "http://localhost:8080", logrus.New())
	fileBody := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'}
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", "a.png")
	require.NoError(t, err)
	_, err = part.Write(fileBody)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(1<<20))
	file, header, err := req.FormFile("file")
	require.NoError(t, err)
	t.Cleanup(func() { _ = file.Close() })

	_, err = ctrl.saveFile(context.Background(), file, header, "")
	require.Error(t, err)
}

func TestSaveUploadedFile_RejectOversizedContent(t *testing.T) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", "a.bin")
	require.NoError(t, err)
	_, err = part.Write([]byte("123456"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(1<<20))
	_, header, err := req.FormFile("file")
	require.NoError(t, err)

	err = saveUploadedFile(header, "/tmp/uploads_test/oversized.bin", 5)
	require.Error(t, err)
}

func newTestMultipartFile(t *testing.T, b []byte) multipart.File {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", "x.bin")
	require.NoError(t, err)
	_, err = part.Write(b)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(1<<20))
	f, _, err := req.FormFile("file")
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	return f
}

func TestUploadCtrl_GenerateBannerMediaDownloadURL_NoUploadService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New())
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.GET("/download/banner/media/:file_id", ctrl.GenerateBannerMediaDownloadURL)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/download/banner/media/1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadCtrl_GenerateBusinessTemporaryDownloadURL_Variants(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auditRepo := &testAuditRepo{
		byID: map[uint64]*entity.AuditLog{
			1: {
				ID:     1,
				Action: "file_asset",
				Module: "file_upload",
				RequestData: `{"key":"protected-video/20260318/a.mp4","usage":"protected","category":"video","protected":true}`,
			},
			2: {
				ID:     2,
				Action: "file_asset",
				Module: "file_upload",
				RequestData: `{"key":"protected-attachment/20260318/a.pdf","usage":"protected","category":"attachment","protected":true}`,
			},
			3: {
				ID:     3,
				Action: "file_asset",
				Module: "file_upload",
				RequestData: `{"key":"embedded-image/20260318/a.png","usage":"embedded","category":"image","protected":false}`,
			},
		},
	}
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New()).WithAuditRepo(auditRepo)
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.GET("/download/course/video/:file_id", ctrl.GenerateCourseVideoDownloadURL)
	r.GET("/download/article/attachment/:file_id", ctrl.GenerateArticleAttachmentDownloadURL)
	r.GET("/download/course/attachment/:file_id", ctrl.GenerateCourseAttachmentDownloadURL)

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/download/course/video/1?expires_in=300", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/download/article/attachment/2", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/download/course/attachment/2", nil)
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)

	// usage/category mismatch branches
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodGet, "/download/course/video/3", nil)
	r.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusForbidden, w4.Code)
}

func TestUploadCtrl_GenerateAttachmentPresign_ForbiddenAndValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New())

	r1 := gin.New()
	r1.Use(middleware.ErrorMiddleware(logrus.New()))
	r1.Use(withFrontUser())
	r1.GET("/upload/article/attachment/presign", ctrl.GenerateArticleAttachmentPresignURL)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/upload/article/attachment/presign?filename=a.pdf", nil)
	r1.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusForbidden, w1.Code)

	r2 := gin.New()
	r2.Use(middleware.ErrorMiddleware(logrus.New()))
	r2.Use(withAdmin())
	r2.GET("/upload/article/attachment/presign", ctrl.GenerateArticleAttachmentPresignURL)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/upload/article/attachment/presign?filename=", nil)
	r2.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	r3 := gin.New()
	r3.Use(middleware.ErrorMiddleware(logrus.New()))
	r3.Use(withAdmin())
	r3.GET("/upload/course/attachment/presign", ctrl.GenerateCourseAttachmentPresignURL)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/upload/course/attachment/presign?filename=a.pdf&expires_in=10", nil)
	r3.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusBadRequest, w3.Code)
}

func TestUploadCtrl_GenerateCourseVideoPresignURL_NoCOS_AndAuditError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrlNoCOS := NewUploadController("/tmp/uploads_test", "http://localhost:8080", logrus.New())
	r1 := gin.New()
	r1.Use(middleware.ErrorMiddleware(logrus.New()))
	r1.Use(withAdmin())
	r1.GET("/upload/course/video/presign", ctrlNoCOS.GenerateCourseVideoPresignURL)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/upload/course/video/presign?filename=a.mp4", nil)
	r1.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusBadRequest, w1.Code)

	ctrlAuditErr := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New()).WithAuditRepo(&errAuditRepo{createErr: errors.New("audit fail")})
	r2 := gin.New()
	r2.Use(middleware.ErrorMiddleware(logrus.New()))
	r2.Use(withAdmin())
	r2.GET("/upload/course/video/presign", ctrlAuditErr.GenerateCourseVideoPresignURL)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/upload/course/video/presign?filename=a.mp4", nil)
	r2.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusInternalServerError, w2.Code)
}

func TestUploadCtrl_GenerateCourseAttachmentPresignURL_ForbiddenAndInvalidExpires(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New())

	r1 := gin.New()
	r1.Use(middleware.ErrorMiddleware(logrus.New()))
	r1.Use(withFrontUser())
	r1.GET("/upload/course/attachment/presign", ctrl.GenerateCourseAttachmentPresignURL)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/upload/course/attachment/presign?filename=a.pdf", nil)
	r1.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusForbidden, w1.Code)

	r2 := gin.New()
	r2.Use(middleware.ErrorMiddleware(logrus.New()))
	r2.Use(withAdmin())
	r2.GET("/upload/course/attachment/presign", ctrl.GenerateCourseAttachmentPresignURL)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/upload/course/attachment/presign?filename=a.pdf&expires_in=10", nil)
	r2.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

func TestUploadCtrl_GenerateStaticMaterialURL_NoCOS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auditRepo := &testAuditRepo{
		byID: map[uint64]*entity.AuditLog{
			1: {
				ID:          1,
				Action:      "file_asset",
				Module:      "file_upload",
				RequestData: `{"key":"embedded-image/20260318/a.png","usage":"embedded","category":"image","protected":false}`,
			},
		},
	}
	ctrl := NewUploadController("/tmp/uploads_test", "http://localhost:8080", logrus.New()).WithAuditRepo(auditRepo)
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.GET("/download/static/:file_id", ctrl.GenerateStaticMaterialURL)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/download/static/1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadCtrl_RecordUploadBusiness_NilAuditRepoNoop(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest(http.MethodGet, "/x", nil)
	ctx.Set("user_id", uint64(99))
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New())
	err := ctrl.recordUploadBusiness(ctx, "course-video", "course-video/a.mp4", "a.mp4", 900)
	require.NoError(t, err)
}

func TestUploadCtrl_GenerateAdminUploadPresignURL_ServicePath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewUploadControllerWithCOS("/tmp/uploads_test", "http://cos:9000", "http://cos:9000", "miniapp-test", logrus.New()).WithUploadService(&mockUploadFileService{
		generateAdminPresignFn: func(_ context.Context, userID uint64, filename, usage, expiresInRaw string) (*service.AdminPresignResult, error) {
			assert.Equal(t, uint64(99), userID)
			assert.Equal(t, "embedded", usage)
			return &service.AdminPresignResult{
				FileID:    3,
				Key:       "embedded-image/1.png",
				PutURL:    "put",
				ExpiresIn: 600,
				ExpireAt:  1,
				StaticURL: "static",
			}, nil
		},
	})
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.Use(withAdmin())
	r.GET("/admin/upload/files/presign", ctrl.GenerateAdminUploadPresignURL)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/upload/files/presign?filename=1.png&usage=embedded&expires_in=600", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
}
