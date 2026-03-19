package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/response"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

// GenerateTestTokenRequest is the request body for the debug token endpoint.
// Only user_id is required; the real user_type is looked up from the database
// so the token faithfully represents the target user.
// user_type_override may be set to force a specific type (for edge-case tests).
type GenerateTestTokenRequest struct {
	UserID           uint64 `json:"user_id" binding:"required"`
	UserTypeOverride int8   `json:"user_type_override"`
	// ExpirySeconds overrides the default JWT expiry. Zero means use the
	// server default configured in jwt.expiry.
	ExpirySeconds int `json:"expiry_seconds"`
}

// GenerateTestTokenResponse is the response body for the debug token endpoint.
type GenerateTestTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	UserID      uint64 `json:"user_id"`
	UserType    int8   `json:"user_type"`
}

// DebugController exposes debug/development-only endpoints.
// It must only be registered when the debug.enable_test_token config flag is true.
type DebugController struct {
	userRepo      repository.UserRepository
	jwtSecret     string
	jwtExpiry     int
	log           *logrus.Logger
	signingMethod jwt.SigningMethod // nil uses jwt.SigningMethodHS256
}

// NewDebugController creates a new DebugController.
func NewDebugController(userRepo repository.UserRepository, jwtSecret string, jwtExpiry int, log *logrus.Logger) *DebugController {
	return &DebugController{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
		log:       log,
	}
}

// GenerateTestToken handles POST /v1/debug/token.
//
// It looks up the user by user_id in the database and issues a signed JWT
// token carrying the user's real user_type, so every downstream permission
// check behaves exactly as it would in production.
//
// Pass user_type_override != 0 to force a specific user_type (useful for
// edge-case tests where the DB row has a different type).
//
// WARNING: This endpoint is for local/integration testing only. It MUST be
// disabled (debug.enable_test_token: false) in production environments.
func (c *DebugController) GenerateTestToken(ctx *gin.Context) {
	var req GenerateTestTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}

	userType := req.UserTypeOverride
	if userType == 0 {
		user, err := c.userRepo.GetByID(context.Background(), req.UserID)
		if err != nil {
			ctx.Error(apperrors.NewInternal("查询用户失败", err))
			return
		}
		if user == nil {
			ctx.Error(apperrors.NewNotFound("用户不存在", nil))
			return
		}
		userType = user.UserType
	}
	if !isValidDebugUserType(userType) {
		ctx.Error(apperrors.NewBadRequest("用户类型不合法，仅支持1/2/3", nil))
		return
	}

	expiry := c.jwtExpiry
	if req.ExpirySeconds > 0 {
		expiry = req.ExpirySeconds
	}

	tokenStr, err := c.generateToken(req.UserID, userType, expiry)
	if err != nil {
		ctx.Error(apperrors.NewInternal("生成Token失败", err))
		return
	}

	response.Success(ctx, &GenerateTestTokenResponse{
		AccessToken: tokenStr,
		TokenType:   "Bearer",
		ExpiresIn:   expiry,
		UserID:      req.UserID,
		UserType:    userType,
	})
}

func isValidDebugUserType(userType int8) bool {
	return userType == 1 || userType == 2 || userType == 3
}

func (c *DebugController) generateToken(userID uint64, userType int8, expirySeconds int) (string, error) {
	claims := &middleware.JWTClaims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expirySeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}
	method := c.signingMethod
	if method == nil {
		method = jwt.SigningMethodHS256
	}
	token := jwt.NewWithClaims(method, claims)
	return token.SignedString([]byte(c.jwtSecret))
}
