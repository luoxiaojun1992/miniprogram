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

// ModuleController handles module and module page requests.
type ModuleController struct {
	svc service.ModuleService
	log *logrus.Logger
}

// NewModuleController creates a new ModuleController.
func NewModuleController(svc service.ModuleService, log *logrus.Logger) *ModuleController {
	return &ModuleController{svc: svc, log: log}
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
	response.Success(ctx, modules)
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
