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

// CreateMultipartFile creates a multipart form body containing a single file field.
// It returns the body buffer and the Content-Type header value.
func CreateMultipartFile(t testing.TB, fieldName, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("CreateMultipartFile: create form file: %v", err)
	}
	if _, err = fw.Write(content); err != nil {
		t.Fatalf("CreateMultipartFile: write content: %v", err)
	}
	w.Close()
	return &buf, w.FormDataContentType()
}

// CreateMultipartFileWithFields creates a multipart form body with a file and extra string fields.
func CreateMultipartFileWithFields(t testing.TB, fieldName, filename string, content []byte, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("CreateMultipartFileWithFields: write field %q: %v", k, err)
		}
	}
	fw, err := w.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("CreateMultipartFileWithFields: create form file: %v", err)
	}
	if _, err = fw.Write(content); err != nil {
		t.Fatalf("CreateMultipartFileWithFields: write content: %v", err)
	}
	w.Close()
	return &buf, w.FormDataContentType()
}
