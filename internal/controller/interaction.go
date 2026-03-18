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

// StudyRecordController handles study record requests.
type StudyRecordController struct {
	svc service.StudyRecordService
	log *logrus.Logger
}

// NewStudyRecordController creates a new StudyRecordController.
func NewStudyRecordController(svc service.StudyRecordService, log *logrus.Logger) *StudyRecordController {
	return &StudyRecordController{svc: svc, log: log}
}

// List handles GET /study-records.
func (c *StudyRecordController) List(ctx *gin.Context) {
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
	records, total, err := c.svc.List(ctx, userID, q.GetPage(), q.GetPageSize())
	if err != nil {
		ctx.Error(err)
		return
	}
	response.PaginatedSuccess(ctx, records, total, q.GetPage(), q.GetPageSize())
}

// Update handles POST /study-records.
func (c *StudyRecordController) Update(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		ctx.Error(apperrors.NewUnauthorized("未登录", nil))
		return
	}
	var req dto.UpdateStudyRecordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if err := c.svc.Update(ctx, userID, &req); err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, nil)
}
