// Package testutil provides test helpers for HTTP request/response testing.
package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

const TestJWTSecret = "test-secret"

// NewTestEngine creates a Gin engine suitable for controller tests.
func NewTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	log := logrus.New()
	log.SetOutput(new(bytes.Buffer)) // suppress log output in tests
	r.Use(middleware.ErrorMiddleware(log))
	return r
}

// NewTestEngineWithAuth creates a Gin engine with JWT auth middleware injected.
func NewTestEngineWithAuth(userID uint64, userType int8) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	log := logrus.New()
	log.SetOutput(new(bytes.Buffer))
	r.Use(middleware.ErrorMiddleware(log))
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("user_type", userType)
		c.Next()
	})
	return r
}

// GenerateTestToken generates a JWT token for test purposes.
func GenerateTestToken(userID uint64, userType int8) string {
	claims := &middleware.JWTClaims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, _ := token.SignedString([]byte(TestJWTSecret))
	return str
}

// PerformRequest performs an HTTP request on a Gin engine and returns the recorded response.
func PerformRequest(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req, _ = http.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ParseResponse parses a JSON response body into a map.
func ParseResponse(w *httptest.ResponseRecorder) map[string]interface{} {
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return resp
}

// ErrInternal is a convenience error for tests.
var ErrInternal = errors.NewInternal("internal error", nil)
