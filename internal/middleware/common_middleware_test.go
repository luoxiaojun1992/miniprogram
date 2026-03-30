package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestIDMiddleware_UsesIncomingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestIDMiddleware())
	r.GET("/ping", func(ctx *gin.Context) {
		assert.Equal(t, "rid-from-client", ctx.GetString("request_id"))
		ctx.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Request-ID", "rid-from-client")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "rid-from-client", w.Header().Get("X-Request-ID"))
}

func TestRequestIDMiddleware_GeneratesIDWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestIDMiddleware())
	r.GET("/ping", func(ctx *gin.Context) {
		assert.NotEmpty(t, ctx.GetString("request_id"))
		ctx.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestCorsMiddleware_OptionsRequestAbortsWith204(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CorsMiddleware())
	nextCalled := false
	r.OPTIONS("/ping", func(ctx *gin.Context) {
		nextCalled = true
		ctx.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.False(t, nextCalled)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS, PATCH", w.Header().Get("Access-Control-Allow-Methods"))
}

func TestCorsMiddleware_NonOptionsRequestPassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CorsMiddleware())
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Content-Type, Authorization, X-Request-ID", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
}

func TestRecoveryMiddleware_RecoversFromPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	r := gin.New()
	r.Use(RecoveryMiddleware(logger))
	r.GET("/panic", func(ctx *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"code":500000`)
	assert.Contains(t, w.Body.String(), `"message":"服务器内部错误"`)
}

func TestLoggerMiddleware_LogsRequestFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.JSONFormatter{DisableTimestamp: true})

	r := gin.New()
	r.Use(LoggerMiddleware(logger))
	r.Use(func(ctx *gin.Context) {
		ctx.Set("request_id", "rid-logger")
		ctx.Next()
	})
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.Status(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "10.1.2.3:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	require.NotEmpty(t, buf.String())

	var entry map[string]any
	err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry)
	require.NoError(t, err)
	assert.Equal(t, "request", entry["msg"])
	assert.Equal(t, "GET", entry["method"])
	assert.Equal(t, "/ping", entry["path"])
	assert.Equal(t, float64(http.StatusCreated), entry["status"])
	assert.Equal(t, "rid-logger", entry["request_id"])
}
