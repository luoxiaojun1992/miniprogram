package controller

import (
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

// CollectionController handles collection requests.
type CollectionController struct {
	svc service.CollectionService
	log *logrus.Logger
}

// NewCollectionController creates a new CollectionController.
func NewCollectionController(svc service.CollectionService, log *logrus.Logger) *CollectionController {
	return &CollectionController{svc: svc, log: log}
}

// List handles GET /collections.
func (c *CollectionController) List(ctx *gin.Context) {
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
	var contentType *int8
	if ct := ctx.Query("content_type"); ct != "" {
		v, _ := strconv.ParseInt(ct, 10, 8)
		t := int8(v)
		contentType = &t
	}
	cols, total, err := c.svc.List(ctx, userID, q.GetPage(), q.GetPageSize(), contentType)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.PaginatedSuccess(ctx, cols, total, q.GetPage(), q.GetPageSize())
}

// Add handles POST /collections/:content_type/:content_id.
func (c *CollectionController) Add(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		ctx.Error(apperrors.NewUnauthorized("未登录", nil))
		return
	}
	ct, err := strconv.ParseInt(ctx.Param("content_type"), 10, 8)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的内容类型", err))
		return
	}
	cid, err := strconv.ParseUint(ctx.Param("content_id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的内容ID", err))
		return
	}
	if svcErr := c.svc.Add(ctx, userID, int8(ct), cid); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.SuccessWithStatus(ctx, http.StatusCreated, nil)
}

// Remove handles DELETE /collections/:content_type/:content_id.
func (c *CollectionController) Remove(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		ctx.Error(apperrors.NewUnauthorized("未登录", nil))
		return
	}
	ct, err := strconv.ParseInt(ctx.Param("content_type"), 10, 8)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的内容类型", err))
		return
	}
	cid, err := strconv.ParseUint(ctx.Param("content_id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的内容ID", err))
		return
	}
	if svcErr := c.svc.Remove(ctx, userID, int8(ct), cid); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// LikeController handles like requests.
type LikeController struct {
	svc service.LikeService
	log *logrus.Logger
}

// NewLikeController creates a new LikeController.
func NewLikeController(svc service.LikeService, log *logrus.Logger) *LikeController {
	return &LikeController{svc: svc, log: log}
}

// Add handles POST /likes/:content_type/:content_id.
func (c *LikeController) Add(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		ctx.Error(apperrors.NewUnauthorized("未登录", nil))
		return
	}
	ct, err := strconv.ParseInt(ctx.Param("content_type"), 10, 8)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的内容类型", err))
		return
	}
	cid, err := strconv.ParseUint(ctx.Param("content_id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的内容ID", err))
		return
	}
	if svcErr := c.svc.Add(ctx, userID, int8(ct), cid); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.SuccessWithStatus(ctx, http.StatusCreated, gin.H{"content_type": ct, "content_id": cid, "is_liked": true})
}

// Remove handles DELETE /likes/:content_type/:content_id.
func (c *LikeController) Remove(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		ctx.Error(apperrors.NewUnauthorized("未登录", nil))
		return
	}
	ct, err := strconv.ParseInt(ctx.Param("content_type"), 10, 8)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的内容类型", err))
		return
	}
	cid, err := strconv.ParseUint(ctx.Param("content_id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的内容ID", err))
		return
	}
	if svcErr := c.svc.Remove(ctx, userID, int8(ct), cid); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, gin.H{"content_type": ct, "content_id": cid, "is_liked": false})
}

// CommentController handles comment requests.
type CommentController struct {
	svc service.CommentService
	log *logrus.Logger
}

// NewCommentController creates a new CommentController.
func NewCommentController(svc service.CommentService, log *logrus.Logger) *CommentController {
	return &CommentController{svc: svc, log: log}
}

// List handles GET /comments/:content_type/:content_id.
func (c *CommentController) List(ctx *gin.Context) {
	ct, err := strconv.ParseInt(ctx.Param("content_type"), 10, 8)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的内容类型", err))
		return
	}
	cid, err := strconv.ParseUint(ctx.Param("content_id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的内容ID", err))
		return
	}
	var q dto.ListQuery
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	comments, total, svcErr := c.svc.List(ctx, int8(ct), cid, q.GetPage(), q.GetPageSize())
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.PaginatedSuccess(ctx, comments, total, q.GetPage(), q.GetPageSize())
}

// Create handles POST /comments/:content_type/:content_id.
func (c *CommentController) Create(ctx *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(ctx)
	if !ok {
		ctx.Error(apperrors.NewUnauthorized("未登录", nil))
		return
	}
	ct, err := strconv.ParseInt(ctx.Param("content_type"), 10, 8)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的内容类型", err))
		return
	}
	cid, err := strconv.ParseUint(ctx.Param("content_id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的内容ID", err))
		return
	}
	var req dto.CreateCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	comment, svcErr := c.svc.Create(ctx, userID, int8(ct), cid, &req)
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.SuccessWithStatus(ctx, http.StatusCreated, comment)
}

// AdminList handles GET /admin/comments.
func (c *CommentController) AdminList(ctx *gin.Context) {
	var q dto.ListQuery
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	var status *int8
	if s := ctx.Query("status"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 8)
		t := int8(v)
		status = &t
	}
	comments, total, err := c.svc.AdminList(ctx, q.GetPage(), q.GetPageSize(), status)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.PaginatedSuccess(ctx, comments, total, q.GetPage(), q.GetPageSize())
}

// AdminAudit handles PUT /admin/comments/:id/audit.
func (c *CommentController) AdminAudit(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的评论ID", err))
		return
	}
	var req dto.AuditCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if svcErr := c.svc.Audit(ctx, id, &req); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, nil)
}

// AdminDelete handles DELETE /admin/comments/:id.
func (c *CommentController) AdminDelete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的评论ID", err))
		return
	}
	if svcErr := c.svc.Delete(ctx, id); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	ctx.Status(http.StatusNoContent)
}
