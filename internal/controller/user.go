package controller

import (
	"math/bits"
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

// UserController handles user-related requests.
type UserController struct {
	svc       service.UserService
	log       *logrus.Logger
	uploadSvc service.UploadFileService
}

// NewUserController creates a new UserController.
func NewUserController(svc service.UserService, log *logrus.Logger, uploadSvc ...service.UploadFileService) *UserController {
	var svcUpload service.UploadFileService
	if len(uploadSvc) > 0 {
		svcUpload = uploadSvc[0]
	}
	return &UserController{svc: svc, log: log, uploadSvc: svcUpload}
}

// GetProfile handles GET /users/profile.
func (c *UserController) GetProfile(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		ctx.Error(apperrors.NewUnauthorized("未登录", nil))
		return
	}
	user, err := c.svc.GetProfile(ctx, userID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if c.uploadSvc != nil && user.AvatarFileID != nil && *user.AvatarFileID > 0 {
		if download, downloadErr := c.uploadSvc.GenerateBusinessDownload(ctx.Request.Context(), *user.AvatarFileID, []string{"image"}, "300"); downloadErr == nil {
			user.AvatarURL = download.Download
		}
	}
	response.Success(ctx, user)
}

// UpdateProfile handles PUT /users/profile.
func (c *UserController) UpdateProfile(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		ctx.Error(apperrors.NewUnauthorized("未登录", nil))
		return
	}
	var req dto.UserProfileUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if err := c.svc.UpdateProfile(ctx, userID, &req); err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, nil)
}

// GetPermissions handles GET /users/permissions.
func (c *UserController) GetPermissions(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		ctx.Error(apperrors.NewUnauthorized("未登录", nil))
		return
	}
	roles, permissions, err := c.svc.GetPermissions(ctx, userID)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.Success(ctx, gin.H{"roles": roles, "permissions": permissions})
}

// AdminListUsers handles GET /admin/users.
func (c *UserController) AdminListUsers(ctx *gin.Context) {
	var q dto.ListQuery
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	var userType *int8
	if ut := ctx.Query("user_type"); ut != "" {
		v, _ := strconv.ParseInt(ut, 10, 8)
		t := int8(v)
		userType = &t
	}
	users, total, err := c.svc.List(ctx, q.GetPage(), q.GetPageSize(), q.Keyword, userType)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.PaginatedSuccess(ctx, users, total, q.GetPage(), q.GetPageSize())
}

// AdminGetUser handles GET /admin/users/:id.
func (c *UserController) AdminGetUser(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的用户ID", err))
		return
	}
	user, svcErr := c.svc.GetByID(ctx, id)
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, user)
}

// AdminCreateUser handles POST /admin/users.
func (c *UserController) AdminCreateUser(ctx *gin.Context) {
	var req dto.CreateAdminUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	id, err := c.svc.CreateAdminUser(ctx, &req)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.SuccessWithStatus(ctx, http.StatusCreated, gin.H{"id": id})
}

// AdminUpdateUser handles PUT /admin/users/:id.
func (c *UserController) AdminUpdateUser(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的用户ID", err))
		return
	}
	operatorID, _ := middleware.GetCurrentUserID(ctx)
	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if svcErr := c.svc.UpdateUser(ctx, id, &req, operatorID); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, nil)
}

// AdminDeleteUser handles DELETE /admin/users/:id.
func (c *UserController) AdminDeleteUser(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的用户ID", err))
		return
	}
	if svcErr := c.svc.DeleteUser(ctx, id); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// AdminAssignRoles handles PUT /admin/users/:id/roles.
func (c *UserController) AdminAssignRoles(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的用户ID", err))
		return
	}
	var req dto.AssignRolesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if svcErr := c.svc.AssignRoles(ctx, id, &req); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, nil)
}

// AdminAddUserTag handles POST /admin/users/:id/tags.
func (c *UserController) AdminAddUserTag(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的用户ID", err))
		return
	}
	var req dto.AddTagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	tagID, svcErr := c.svc.AddTag(ctx, id, &req)
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.SuccessWithStatus(ctx, http.StatusCreated, gin.H{"id": tagID})
}

// AdminDeleteUserTag handles DELETE /admin/users/:id/tags.
func (c *UserController) AdminDeleteUserTag(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的用户ID", err))
		return
	}
	tagIDStr := ctx.Query("tag_id")
	tagID64, err := strconv.ParseUint(tagIDStr, 10, bits.UintSize)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的标签ID", err))
		return
	}
	tagID := uint(tagID64)
	if svcErr := c.svc.DeleteTag(ctx, userID, tagID); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	ctx.Status(http.StatusNoContent)
}
