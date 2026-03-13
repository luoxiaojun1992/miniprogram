package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/response"
	"github.com/luoxiaojun1992/miniprogram/internal/service"
)

// NotificationController handles notification requests.
type NotificationController struct {
	svc service.NotificationService
	log *logrus.Logger
}

// NewNotificationController creates a new NotificationController.
func NewNotificationController(svc service.NotificationService, log *logrus.Logger) *NotificationController {
	return &NotificationController{svc: svc, log: log}
}

// List handles GET /notifications.
func (c *NotificationController) List(ctx *gin.Context) {
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
	var isRead *bool
	if r := ctx.Query("is_read"); r != "" {
		b := r == "true" || r == "1"
		isRead = &b
	}
	notifs, total, unreadCount, err := c.svc.List(ctx, userID, q.GetPage(), q.GetPageSize(), isRead)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list": notifs,
			"pagination": gin.H{
				"total":     total,
				"page":      q.GetPage(),
				"page_size": q.GetPageSize(),
			},
			"unread_count": unreadCount,
		},
	})
}

// MarkRead handles PUT /notifications/:id/read.
func (c *NotificationController) MarkRead(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的通知ID", err))
		return
	}
	if svcErr := c.svc.MarkRead(ctx, id); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, nil)
}

// MarkAllRead handles PUT /notifications/read-all.
func (c *NotificationController) MarkAllRead(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		ctx.Error(apperrors.NewUnauthorized("未登录", nil))
		return
	}
	if err := c.svc.MarkAllRead(ctx, userID); err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, nil)
}
