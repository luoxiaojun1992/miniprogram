package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/response"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
	"github.com/luoxiaojun1992/miniprogram/internal/service"
)

// ModuleController handles module and module page requests.
type ModuleController struct {
	svc           service.ModuleService
	log           *logrus.Logger
	accessChecker *accessChecker
}

// NewModuleController creates a new ModuleController.
func NewModuleController(
	svc service.ModuleService,
	log *logrus.Logger,
	deps ...interface{},
) *ModuleController {
	var contentPermRepo repository.ContentPermissionRepository
	var roleRepo repository.RoleRepository
	for _, dep := range deps {
		switch v := dep.(type) {
		case repository.ContentPermissionRepository:
			contentPermRepo = v
		case repository.RoleRepository:
			roleRepo = v
		}
	}
	return &ModuleController{
		svc:           svc,
		log:           log,
		accessChecker: newAccessChecker(contentPermRepo, roleRepo),
	}
}

// List handles GET /modules.
func (c *ModuleController) List(ctx *gin.Context) {
	var status *int8
	if s := ctx.Query("status"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 8)
		t := int8(v)
		status = &t
	}
	modules, err := c.svc.List(ctx, status)
	if err != nil {
		ctx.Error(err)
		return
	}
	userID, _ := middleware.GetCurrentUserID(ctx)
	var uid *uint64
	if userID > 0 {
		uid = &userID
	}
	filtered := make([]*entity.Module, 0, len(modules))
	for _, item := range modules {
		allowed, accessErr := c.accessChecker.canAccess(ctx, 3, uint64(item.ID), uid, nil)
		if accessErr != nil {
			ctx.Error(accessErr)
			return
		}
		if allowed {
			filtered = append(filtered, item)
		}
	}
	response.Success(ctx, filtered)
}

// Create handles POST /admin/modules.
func (c *ModuleController) Create(ctx *gin.Context) {
	var req dto.CreateModuleRequest
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

// Update handles PUT /admin/modules/:id.
func (c *ModuleController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的模块ID", err))
		return
	}
	var req dto.CreateModuleRequest
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

// Delete handles DELETE /admin/modules/:id.
func (c *ModuleController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的模块ID", err))
		return
	}
	if svcErr := c.svc.Delete(ctx, uint(id)); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// GetPages handles GET /admin/modules/:id/pages.
func (c *ModuleController) GetPages(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的模块ID", err))
		return
	}
	pages, svcErr := c.svc.GetPages(ctx, uint(id))
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, pages)
}

// CreatePage handles POST /admin/modules/:id/pages.
func (c *ModuleController) CreatePage(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的模块ID", err))
		return
	}
	var req dto.CreateModulePageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	pageID, svcErr := c.svc.CreatePage(ctx, uint(id), &req)
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.SuccessWithStatus(ctx, http.StatusCreated, gin.H{"id": pageID})
}

// UpdatePage handles PUT /admin/modules/:id/pages/:page_id.
func (c *ModuleController) UpdatePage(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的模块ID", err))
		return
	}
	pageID, err := strconv.ParseUint(ctx.Param("page_id"), 10, 32)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的页面ID", err))
		return
	}
	var req dto.CreateModulePageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if svcErr := c.svc.UpdatePage(ctx, uint(id), uint(pageID), &req); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, nil)
}

// DeletePage handles DELETE /admin/modules/:id/pages/:page_id.
func (c *ModuleController) DeletePage(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的模块ID", err))
		return
	}
	pageID, err := strconv.ParseUint(ctx.Param("page_id"), 10, 32)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的页面ID", err))
		return
	}
	if svcErr := c.svc.DeletePage(ctx, uint(id), uint(pageID)); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	ctx.Status(http.StatusNoContent)
}
