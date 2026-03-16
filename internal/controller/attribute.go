package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/response"
	"github.com/luoxiaojun1992/miniprogram/internal/service"
)

// AttributeController handles attribute-related requests.
type AttributeController struct {
	svc service.AttributeService
	log *logrus.Logger
}

// NewAttributeController creates a new AttributeController.
func NewAttributeController(svc service.AttributeService, log *logrus.Logger) *AttributeController {
	return &AttributeController{svc: svc, log: log}
}

// List handles GET /admin/attributes.
func (c *AttributeController) List(ctx *gin.Context) {
	attrs, err := c.svc.List(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, attrs)
}

// Create handles POST /admin/attributes.
func (c *AttributeController) Create(ctx *gin.Context) {
	var req dto.CreateAttributeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	id, err := c.svc.Create(ctx, &req)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.SuccessWithStatus(ctx, http.StatusCreated, gin.H{"id": id})
}

// Update handles PUT /admin/attributes/:id.
func (c *AttributeController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的属性ID", err))
		return
	}
	var req dto.UpdateAttributeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if svcErr := c.svc.Update(ctx, uint(id), &req); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, nil)
}

// Delete handles DELETE /admin/attributes/:id.
func (c *AttributeController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的属性ID", err))
		return
	}
	if svcErr := c.svc.Delete(ctx, uint(id)); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// ListUserAttributes handles GET /admin/users/:id/attributes.
func (c *AttributeController) ListUserAttributes(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的用户ID", err))
		return
	}
	uas, svcErr := c.svc.ListUserAttributes(ctx, userID)
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, uas)
}

// SetUserAttribute handles POST /admin/users/:id/attributes.
func (c *AttributeController) SetUserAttribute(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的用户ID", err))
		return
	}
	var req dto.SetUserAttributeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if svcErr := c.svc.SetUserAttribute(ctx, userID, &req); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, nil)
}

// DeleteUserAttribute handles DELETE /admin/users/:id/attributes.
func (c *AttributeController) DeleteUserAttribute(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的用户ID", err))
		return
	}
	attrIDStr := ctx.Query("attribute_id")
	attrID, err := strconv.ParseUint(attrIDStr, 10, 32)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的属性ID", err))
		return
	}
	if svcErr := c.svc.DeleteUserAttribute(ctx, userID, uint(attrID)); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	ctx.Status(http.StatusNoContent)
}
