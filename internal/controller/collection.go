package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/response"
	"github.com/luoxiaojun1992/miniprogram/internal/service"
)

// CollectionController handles collection requests.
type CollectionController struct {
	svc service.CollectionService
	log *logrus.Logger
}

// NewCollectionController creates a new CollectionController.
func NewCollectionController(svc service.CollectionService, log *logrus.Logger) *CollectionController {
	return &CollectionController{svc: svc, log: log}
}

// List handles GET /collections.
func (c *CollectionController) List(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		ctx.Error(apperrors.NewUnauthorized("未登录", nil))
		return
	}
	var q dto.ListQuery
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	var contentType *int8
	if ct := ctx.Query("content_type"); ct != "" {
		v, _ := strconv.ParseInt(ct, 10, 8)
		t := int8(v)
		contentType = &t
	}
	cols, total, err := c.svc.List(ctx, userID, q.GetPage(), q.GetPageSize(), contentType)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.PaginatedSuccess(ctx, cols, total, q.GetPage(), q.GetPageSize())
}

// Add handles POST /collections/:content_type/:content_id.
func (c *CollectionController) Add(ctx *gin.Context) {
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
	response.SuccessWithStatus(ctx, http.StatusCreated, nil)
}

// Remove handles DELETE /collections/:content_type/:content_id.
func (c *CollectionController) Remove(ctx *gin.Context) {
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
	ctx.Status(http.StatusNoContent)
}
