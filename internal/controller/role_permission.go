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

// RoleController handles role management requests.
type RoleController struct {
	svc service.RoleService
	log *logrus.Logger
}

// NewRoleController creates a new RoleController.
func NewRoleController(svc service.RoleService, log *logrus.Logger) *RoleController {
	return &RoleController{svc: svc, log: log}
}

// List handles GET /admin/roles.
func (c *RoleController) List(ctx *gin.Context) {
	roles, err := c.svc.List(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, roles)
}

// GetByID handles GET /admin/roles/:id.
func (c *RoleController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的角色ID", err))
		return
	}
	role, svcErr := c.svc.GetByID(ctx, uint(id))
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, role)
}

// Create handles POST /admin/roles.
func (c *RoleController) Create(ctx *gin.Context) {
	var req dto.CreateRoleRequest
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

// Update handles PUT /admin/roles/:id.
func (c *RoleController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的角色ID", err))
		return
	}
	var req dto.UpdateRoleRequest
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

// Delete handles DELETE /admin/roles/:id.
func (c *RoleController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的角色ID", err))
		return
	}
	if svcErr := c.svc.Delete(ctx, uint(id)); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// PermissionController handles permission requests.
type PermissionController struct {
	svc service.PermissionService
	log *logrus.Logger
}

// NewPermissionController creates a new PermissionController.
func NewPermissionController(svc service.PermissionService, log *logrus.Logger) *PermissionController {
	return &PermissionController{svc: svc, log: log}
}

// GetTree handles GET /admin/permissions.
func (c *PermissionController) GetTree(ctx *gin.Context) {
	tree, err := c.svc.GetTree(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, tree)
}
