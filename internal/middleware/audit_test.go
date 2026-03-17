package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
)

type testAuditLogSvc struct {
	logs []*entity.AuditLog
}

func (s *testAuditLogSvc) List(ctx context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error) {
	return nil, 0, nil
}

func (s *testAuditLogSvc) Log(ctx context.Context, log *entity.AuditLog) {
	s.logs = append(s.logs, log)
}

func TestAuditLogMiddleware_AdminWriteSuccess_Logs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &testAuditLogSvc{}
	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("user_id", uint64(99))
		ctx.Next()
	})
	r.Use(AuditLogMiddleware(svc))
	r.POST("/v1/admin/modules", func(ctx *gin.Context) {
		body, _ := ctx.GetRawData()
		assert.Contains(t, string(body), "title")
		ctx.JSON(http.StatusCreated, gin.H{"ok": true})
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/modules", strings.NewReader(`{"title":"a"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Len(t, svc.logs, 1)
	assert.Equal(t, "create", svc.logs[0].Action)
	assert.Equal(t, "modules", svc.logs[0].Module)
	assert.Contains(t, svc.logs[0].Description, "POST /v1/admin/modules")
}

func TestAuditLogMiddleware_AdminGet_NoLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &testAuditLogSvc{}
	r := gin.New()
	r.Use(AuditLogMiddleware(svc))
	r.GET("/v1/admin/modules", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	})
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/modules", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, svc.logs, 0)
}
