package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/response"
	"github.com/luoxiaojun1992/miniprogram/internal/service"
)

// FollowController handles follow requests.
type FollowController struct {
	svc service.FollowService
	log *logrus.Logger
}

// NewFollowController creates a new FollowController.
func NewFollowController(svc service.FollowService, log *logrus.Logger) *FollowController {
	return &FollowController{svc: svc, log: log}
}

// Add handles POST /follows/:user_id.
func (c *FollowController) Add(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		ctx.Error(apperrors.NewUnauthorized("未登录", nil))
		return
	}
	targetID, err := strconv.ParseUint(ctx.Param("user_id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的用户ID", err))
		return
	}
	if svcErr := c.svc.Add(ctx, userID, targetID); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.SuccessWithStatus(ctx, http.StatusCreated, gin.H{"user_id": targetID, "is_followed": true})
}

// Remove handles DELETE /follows/:user_id.
func (c *FollowController) Remove(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		ctx.Error(apperrors.NewUnauthorized("未登录", nil))
		return
	}
	targetID, err := strconv.ParseUint(ctx.Param("user_id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的用户ID", err))
		return
	}
	if svcErr := c.svc.Remove(ctx, userID, targetID); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, gin.H{"user_id": targetID, "is_followed": false})
}

