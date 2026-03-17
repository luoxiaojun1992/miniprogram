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
