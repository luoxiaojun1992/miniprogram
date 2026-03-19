package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/userstate"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

// JWTClaims represents the JWT claims.
type JWTClaims struct {
	UserID   uint64 `json:"user_id"`
	UserType int8   `json:"user_type"`
	jwt.RegisteredClaims
}

// JWTAuthMiddleware validates the JWT token from the Authorization header.
func JWTAuthMiddleware(
	secret string,
	userRepo repository.UserRepository,
	uaRepo repository.UserAttributeRepository,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.Error(errors.NewUnauthorized("缺少认证Token", nil))
			ctx.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			ctx.Error(errors.NewUnauthorized("Token格式错误", nil))
			ctx.Abort()
			return
		}

		tokenStr := parts[1]
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.NewUnauthorized("无效的签名方法", nil)
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			ctx.Error(errors.NewUnauthorized("Token无效或已过期", err))
			ctx.Abort()
			return
		}
		if userRepo != nil {
			if err = ensureNotFrozen(ctx.Request.Context(), userRepo, uaRepo, claims.UserID); err != nil {
				ctx.Error(err)
				ctx.Abort()
				return
			}
		}

		ctx.Set("user_id", claims.UserID)
		ctx.Set("user_type", claims.UserType)
		ctx.Next()
	}
}

// OptionalJWTAuthMiddleware validates the JWT token if present, but does not abort if missing.
func OptionalJWTAuthMiddleware(
	secret string,
	userRepo repository.UserRepository,
	uaRepo repository.UserAttributeRepository,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			ctx.Next()
			return
		}

		tokenStr := parts[1]
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.NewUnauthorized("无效的签名方法", nil)
			}
			return []byte(secret), nil
		})
		if err == nil && token.Valid {
			if userRepo != nil {
				if stateErr := ensureNotFrozen(ctx.Request.Context(), userRepo, uaRepo, claims.UserID); stateErr != nil {
					ctx.Next()
					return
				}
			}
			ctx.Set("user_id", claims.UserID)
			ctx.Set("user_type", claims.UserType)
		}
		ctx.Next()
	}
}

func ensureNotFrozen(
	ctx context.Context,
	userRepo repository.UserRepository,
	uaRepo repository.UserAttributeRepository,
	userID uint64,
) error {
	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.NewUnauthorized("用户不存在", nil)
	}
	var attrs []*entity.UserAttribute
	if uaRepo != nil {
		attrs, err = uaRepo.ListByUserID(ctx, userID)
		if err != nil {
			return err
		}
	}
	if userstate.IsFrozen(user, attrs, time.Now()) {
		return errors.NewUnauthorized("账号已被冻结", nil)
	}
	return nil
}

// PermissionMiddleware checks if the current user has the required permission code.
func PermissionMiddleware(permissionCode string, checker func(ctx *gin.Context, userID uint64, code string) bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userIDVal, exists := ctx.Get("user_id")
		if !exists {
			ctx.Error(errors.NewUnauthorized("未登录", nil))
			ctx.Abort()
			return
		}
		userID, ok := userIDVal.(uint64)
		if !ok {
			ctx.Error(errors.NewUnauthorized("用户信息异常", nil))
			ctx.Abort()
			return
		}
		if !checker(ctx, userID, permissionCode) {
			ctx.Error(errors.NewForbidden("权限不足", nil))
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

// GetCurrentUserID returns the current user ID from the Gin context.
func GetCurrentUserID(ctx *gin.Context) (uint64, bool) {
	val, exists := ctx.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := val.(uint64)
	return id, ok
}

// GetCurrentUserType returns the current user type from the Gin context.
func GetCurrentUserType(ctx *gin.Context) (int8, bool) {
	val, exists := ctx.Get("user_type")
	if !exists {
		return 0, false
	}
	t, ok := val.(int8)
	return t, ok
}

// RequireAdmin returns 403 if the current user is not an admin (user_type >= 2).
func RequireAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userType, ok := GetCurrentUserType(ctx)
		if !ok || userType < 2 {
			ctx.JSON(http.StatusForbidden, gin.H{
				"code":    403001,
				"message": "需要管理员权限",
				"data":    nil,
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
