package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type fakeRateLimitStore struct {
	counts map[string]int64
	err    error
}

func (f *fakeRateLimitStore) Increment(_ context.Context, key string, _ time.Duration) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	if f.counts == nil {
		f.counts = map[string]int64{}
	}
	f.counts[key]++
	return f.counts[key], nil
}

func TestRateLimitMiddleware_ExceedsLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeRateLimitStore{}
	r := gin.New()
	r.Use(RateLimitMiddleware(store, 1, time.Minute, nil))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	req1 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req1.RemoteAddr = "10.0.0.1:1234"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	req2 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req2.RemoteAddr = "10.0.0.1:1234"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRateLimitMiddleware_BackendErrorAllowsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeRateLimitStore{err: errors.New("redis down")}
	r := gin.New()
	r.Use(RateLimitMiddleware(store, 1, time.Minute, nil))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
