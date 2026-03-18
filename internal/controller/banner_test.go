package controller

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type mockBannerService struct {
	listFn      func(ctx context.Context) ([]*entity.Banner, error)
	adminListFn func(ctx context.Context, status *int8) ([]*entity.Banner, error)
	createFn    func(ctx context.Context, req *dto.CreateBannerRequest) (uint64, error)
	updateFn    func(ctx context.Context, id uint64, req *dto.CreateBannerRequest) error
	deleteFn    func(ctx context.Context, id uint64) error
}

func (m *mockBannerService) List(ctx context.Context) ([]*entity.Banner, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}
func (m *mockBannerService) AdminList(ctx context.Context, status *int8) ([]*entity.Banner, error) {
	if m.adminListFn != nil {
		return m.adminListFn(ctx, status)
	}
	return nil, nil
}
func (m *mockBannerService) Create(ctx context.Context, req *dto.CreateBannerRequest) (uint64, error) {
	if m.createFn != nil {
		return m.createFn(ctx, req)
	}
	return 1, nil
}
func (m *mockBannerService) Update(ctx context.Context, id uint64, req *dto.CreateBannerRequest) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, req)
	}
	return nil
}
func (m *mockBannerService) Delete(ctx context.Context, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func TestBannerCtrl_List_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	ctrl := NewBannerController(&mockBannerService{
		listFn: func(_ context.Context) ([]*entity.Banner, error) {
			return []*entity.Banner{{ID: 1, Title: "B1"}}, nil
		},
	}, logrus.New())
	r.GET("/banners", ctrl.List)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/banners", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBannerCtrl_AdminCreate_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	ctrl := NewBannerController(&mockBannerService{}, logrus.New())
	r.POST("/admin/banners", ctrl.AdminCreate)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/admin/banners", bytes.NewBufferString(`{"title":""}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBannerCtrl_AdminDelete_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	ctrl := NewBannerController(&mockBannerService{
		deleteFn: func(_ context.Context, _ uint64) error {
			return apperrors.NewNotFound("nf", nil)
		},
	}, logrus.New())
	r.DELETE("/admin/banners/:id", ctrl.AdminDelete)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/admin/banners/1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestBannerCtrl_AdminList_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	ctrl := NewBannerController(&mockBannerService{
		adminListFn: func(_ context.Context, status *int8) ([]*entity.Banner, error) {
			require.NotNil(t, status)
			assert.Equal(t, int8(1), *status)
			return []*entity.Banner{{ID: 1, Title: "B1", Status: 1}}, nil
		},
	}, logrus.New())
	r.GET("/admin/banners", ctrl.AdminList)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/banners?status=1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBannerCtrl_AdminUpdate_BadID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	ctrl := NewBannerController(&mockBannerService{}, logrus.New())
	r.PUT("/admin/banners/:id", ctrl.AdminUpdate)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/admin/banners/not-a-number", bytes.NewBufferString(`{"title":"ok","image_file_id":1}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBannerCtrl_AdminUpdate_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	ctrl := NewBannerController(&mockBannerService{
		updateFn: func(_ context.Context, id uint64, req *dto.CreateBannerRequest) error {
			assert.Equal(t, uint64(1), id)
			assert.Equal(t, "ok", req.Title)
			return nil
		},
	}, logrus.New())
	r.PUT("/admin/banners/:id", ctrl.AdminUpdate)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/admin/banners/1", bytes.NewBufferString(`{"title":"ok","image_file_id":1}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBannerCtrl_List_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	ctrl := NewBannerController(&mockBannerService{
		listFn: func(_ context.Context) ([]*entity.Banner, error) {
			return nil, apperrors.NewInternal("list failed", nil)
		},
	}, logrus.New())
	r.GET("/banners", ctrl.List)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/banners", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestBannerCtrl_AdminCreate_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	ctrl := NewBannerController(&mockBannerService{
		createFn: func(_ context.Context, req *dto.CreateBannerRequest) (uint64, error) {
			assert.Equal(t, "banner", req.Title)
			return 10, nil
		},
	}, logrus.New())
	r.POST("/admin/banners", ctrl.AdminCreate)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/admin/banners", bytes.NewBufferString(`{"title":"banner","image_file_id":1}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestBannerCtrl_AdminUpdate_BindErrorAndServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r1 := gin.New()
	r1.Use(middleware.ErrorMiddleware(logrus.New()))
	ctrl1 := NewBannerController(&mockBannerService{}, logrus.New())
	r1.PUT("/admin/banners/:id", ctrl1.AdminUpdate)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodPut, "/admin/banners/1", bytes.NewBufferString("{"))
	req1.Header.Set("Content-Type", "application/json")
	r1.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusBadRequest, w1.Code)

	r2 := gin.New()
	r2.Use(middleware.ErrorMiddleware(logrus.New()))
	ctrl2 := NewBannerController(&mockBannerService{
		updateFn: func(_ context.Context, _ uint64, _ *dto.CreateBannerRequest) error {
			return apperrors.NewInternal("update failed", nil)
		},
	}, logrus.New())
	r2.PUT("/admin/banners/:id", ctrl2.AdminUpdate)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPut, "/admin/banners/1", bytes.NewBufferString(`{"title":"ok","image_file_id":1}`))
	req2.Header.Set("Content-Type", "application/json")
	r2.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusInternalServerError, w2.Code)
}

func TestBannerCtrl_AdminDelete_BadIDAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r1 := gin.New()
	r1.Use(middleware.ErrorMiddleware(logrus.New()))
	ctrl1 := NewBannerController(&mockBannerService{}, logrus.New())
	r1.DELETE("/admin/banners/:id", ctrl1.AdminDelete)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodDelete, "/admin/banners/xx", nil)
	r1.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusBadRequest, w1.Code)

	r2 := gin.New()
	r2.Use(middleware.ErrorMiddleware(logrus.New()))
	called := false
	ctrl2 := NewBannerController(&mockBannerService{
		deleteFn: func(_ context.Context, id uint64) error {
			called = true
			assert.Equal(t, uint64(1), id)
			return nil
		},
	}, logrus.New())
	r2.DELETE("/admin/banners/:id", ctrl2.AdminDelete)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodDelete, "/admin/banners/1", nil)
	r2.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNoContent, w2.Code)
	assert.True(t, called)
}
