package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
)

func TestRequireAdmin_RejectsFrontUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint64(1))
		ctx.Set("user_type", int8(1))
		ctx.Next()
	})
	r.Use(middleware.RequireAdmin())
	r.GET("/admin/ping", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/admin/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireAdmin_AllowsNormalAdminAndSystemAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, ut := range []int8{2, 3} {
		r := gin.New()
		r.Use(func(ctx *gin.Context) {
			ctx.Set("user_id", uint64(1))
			ctx.Set("user_type", ut)
			ctx.Next()
		})
		r.Use(middleware.RequireAdmin())
		r.GET("/admin/ping", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })

		req := httptest.NewRequest(http.MethodGet, "/admin/ping", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestRequireAdmin_RejectsUnexpectedType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint64(1))
		ctx.Set("user_type", int8(4))
		ctx.Next()
	})
	r.Use(middleware.RequireAdmin())
	r.GET("/admin/ping", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/admin/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
