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

// StudyRecordController handles study record requests.
// LikeController handles like requests.
type LikeController struct {
	svc service.LikeService
	log *logrus.Logger
}

// NewLikeController creates a new LikeController.
func NewLikeController(svc service.LikeService, log *logrus.Logger) *LikeController {
	return &LikeController{svc: svc, log: log}
}

// Add handles POST /likes/:content_type/:content_id.
func (c *LikeController) Add(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		ctx.Error(apperrors.NewUnauthorized("未登录", nil))
		return
	}
	ct, err := strconv.ParseInt(ctx.Param("content_type"), 10, 8)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的内容类型", err))
		return
	}
	cid, err := strconv.ParseUint(ctx.Param("content_id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的内容ID", err))
		return
	}
	if svcErr := c.svc.Add(ctx, userID, int8(ct), cid); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.SuccessWithStatus(ctx, http.StatusCreated, gin.H{"content_type": ct, "content_id": cid, "is_liked": true})
}

// Remove handles DELETE /likes/:content_type/:content_id.
func (c *LikeController) Remove(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		ctx.Error(apperrors.NewUnauthorized("未登录", nil))
		return
	}
	ct, err := strconv.ParseInt(ctx.Param("content_type"), 10, 8)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的内容类型", err))
		return
	}
	cid, err := strconv.ParseUint(ctx.Param("content_id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的内容ID", err))
		return
	}
	if svcErr := c.svc.Remove(ctx, userID, int8(ct), cid); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, gin.H{"content_type": ct, "content_id": cid, "is_liked": false})
}
