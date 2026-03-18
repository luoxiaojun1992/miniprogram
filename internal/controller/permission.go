package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/pkg/response"
	"github.com/luoxiaojun1992/miniprogram/internal/service"
)

// RoleController handles role management requests.
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
