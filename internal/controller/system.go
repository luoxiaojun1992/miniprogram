package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/response"
	"github.com/luoxiaojun1992/miniprogram/internal/service"
)

// SystemController handles system configuration requests.
type SystemController struct {
	wechatSvc    service.WechatConfigService
	auditSvc     service.AuditLogService
	logConfigSvc service.LogConfigService
	log          *logrus.Logger
}

// NewSystemController creates a new SystemController.
func NewSystemController(
	wechatSvc service.WechatConfigService,
	auditSvc service.AuditLogService,
	logConfigSvc service.LogConfigService,
	log *logrus.Logger,
) *SystemController {
	return &SystemController{
		wechatSvc:    wechatSvc,
		auditSvc:     auditSvc,
		logConfigSvc: logConfigSvc,
		log:          log,
	}
}

// GetWechatConfig handles GET /admin/wechat-config.
func (c *SystemController) GetWechatConfig(ctx *gin.Context) {
	cfg, err := c.wechatSvc.Get(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, cfg)
}

// UpdateWechatConfig handles PUT /admin/wechat-config.
func (c *SystemController) UpdateWechatConfig(ctx *gin.Context) {
	var req dto.UpdateWechatConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if err := c.wechatSvc.Update(ctx, &req); err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, nil)
}

// ListAuditLogs handles GET /admin/audit-logs.
func (c *SystemController) ListAuditLogs(ctx *gin.Context) {
	var q dto.ListQuery
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	module := ctx.Query("module")
	action := ctx.Query("action")
	var startTime, endTime *string
	if st := ctx.Query("start_time"); st != "" {
		startTime = &st
	}
	if et := ctx.Query("end_time"); et != "" {
		endTime = &et
	}
	logs, total, err := c.auditSvc.List(ctx, q.GetPage(), q.GetPageSize(), module, action, startTime, endTime)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.PaginatedSuccess(ctx, logs, total, q.GetPage(), q.GetPageSize())
}

// GetLogConfig handles GET /admin/log-config.
func (c *SystemController) GetLogConfig(ctx *gin.Context) {
	cfg, err := c.logConfigSvc.Get(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, cfg)
}

// UpdateLogConfig handles PUT /admin/log-config.
func (c *SystemController) UpdateLogConfig(ctx *gin.Context) {
	var req dto.UpdateLogConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if err := c.logConfigSvc.Update(ctx, &req); err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, nil)
}
