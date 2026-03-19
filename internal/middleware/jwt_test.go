package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/testutil"
)

func makeJWTToken(t *testing.T, secret string, userID uint64, userType int8) string {
	t.Helper()
	claims := &middleware.JWTClaims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return s
}

func TestJWTAuthMiddleware_BlocksFrozenUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "secret"
	token := makeJWTToken(t, secret, 1, 1)

	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return nil, nil
		},
	}
	uaRepo := &testutil.MockUserAttributeRepository{}

	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.Use(middleware.JWTAuthMiddleware(secret, userRepo, uaRepo))
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthMiddleware_BlocksFrozenAttribute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "secret"
	token := makeJWTToken(t, secret, 2, 1)

	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return &entity.User{ID: id, Status: 1}, nil
		},
	}
	v := int64(1)
	uaRepo := &testutil.MockUserAttributeRepository{
		ListByUserIDFn: func(_ context.Context, userID uint64) ([]*entity.UserAttribute, error) {
			return []*entity.UserAttribute{
				{Attribute: &entity.Attribute{Name: "is_frozen"}, ValueBigint: &v},
			}, nil
		},
	}

	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.Use(middleware.JWTAuthMiddleware(secret, userRepo, uaRepo))
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthMiddleware_AllowsNormalUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "secret"
	token := makeJWTToken(t, secret, 3, 1)

	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return &entity.User{ID: id, Status: 1}, nil
		},
	}
	uaRepo := &testutil.MockUserAttributeRepository{
		ListByUserIDFn: func(_ context.Context, userID uint64) ([]*entity.UserAttribute, error) {
			return nil, nil
		},
	}

	r := gin.New()
	r.Use(middleware.ErrorMiddleware(logrus.New()))
	r.Use(middleware.JWTAuthMiddleware(secret, userRepo, uaRepo))
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
