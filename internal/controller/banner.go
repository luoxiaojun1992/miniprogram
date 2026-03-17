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

// BannerController handles banner requests.
type BannerController struct {
	svc service.BannerService
	log *logrus.Logger
}

// NewBannerController creates a new BannerController.
func NewBannerController(svc service.BannerService, log *logrus.Logger) *BannerController {
	return &BannerController{svc: svc, log: log}
}

// List handles GET /banners.
func (c *BannerController) List(ctx *gin.Context) {
	banners, err := c.svc.List(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, banners)
}

// AdminList handles GET /admin/banners.
func (c *BannerController) AdminList(ctx *gin.Context) {
	var status *int8
	if s := ctx.Query("status"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 8)
		t := int8(v)
		status = &t
	}
	banners, err := c.svc.AdminList(ctx, status)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, banners)
}

// AdminCreate handles POST /admin/banners.
func (c *BannerController) AdminCreate(ctx *gin.Context) {
	var req dto.CreateBannerRequest
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

// AdminUpdate handles PUT /admin/banners/:id.
func (c *BannerController) AdminUpdate(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的轮播图ID", err))
		return
	}
	var req dto.CreateBannerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if svcErr := c.svc.Update(ctx, id, &req); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, nil)
}

// AdminDelete handles DELETE /admin/banners/:id.
func (c *BannerController) AdminDelete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的轮播图ID", err))
		return
	}
	if svcErr := c.svc.Delete(ctx, id); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	ctx.Status(http.StatusNoContent)
}
