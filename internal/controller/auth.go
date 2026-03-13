package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/response"
	"github.com/luoxiaojun1992/miniprogram/internal/service"
)

// AuthController handles authentication requests.
type AuthController struct {
	svc service.AuthService
	log *logrus.Logger
}

// NewAuthController creates a new AuthController.
func NewAuthController(svc service.AuthService, log *logrus.Logger) *AuthController {
	return &AuthController{svc: svc, log: log}
}

// WechatLogin handles POST /auth/wechat-login.
func (c *AuthController) WechatLogin(ctx *gin.Context) {
	var req dto.WechatLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	data, err := c.svc.WechatLogin(ctx, &req)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, data)
}

// AdminLogin handles POST /auth/admin-login.
func (c *AuthController) AdminLogin(ctx *gin.Context) {
	var req dto.AdminLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	data, err := c.svc.AdminLogin(ctx, &req)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, data)
}

// RefreshToken handles POST /auth/refresh.
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		ctx.Error(apperrors.NewUnauthorized("未登录", nil))
		return
	}
	userType, _ := middleware.GetCurrentUserType(ctx)
	data, err := c.svc.RefreshToken(ctx, userID, userType)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, data)
}
