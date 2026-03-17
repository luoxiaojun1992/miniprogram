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

// ArticleController handles article requests.
type ArticleController struct {
	svc service.ArticleService
	log *logrus.Logger
}

// NewArticleController creates a new ArticleController.
func NewArticleController(svc service.ArticleService, log *logrus.Logger) *ArticleController {
	return &ArticleController{svc: svc, log: log}
}

// List handles GET /articles.
func (c *ArticleController) List(ctx *gin.Context) {
	var q dto.ListQuery
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	var moduleID *uint
	if m := ctx.Query("module_id"); m != "" {
		v, _ := strconv.ParseUint(m, 10, 32)
		u := uint(v)
		moduleID = &u
	}
	userID, _ := middleware.GetCurrentUserID(ctx)
	var uid *uint64
	if userID > 0 {
		uid = &userID
	}
	articles, total, err := c.svc.List(ctx, q.GetPage(), q.GetPageSize(), q.Keyword, moduleID, q.Sort, uid)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.PaginatedSuccess(ctx, articles, total, q.GetPage(), q.GetPageSize())
}

// GetByID handles GET /articles/:id.
func (c *ArticleController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的文章ID", err))
		return
	}
	userID, _ := middleware.GetCurrentUserID(ctx)
	var uid *uint64
	if userID > 0 {
		uid = &userID
	}
	article, svcErr := c.svc.GetByID(ctx, id, uid)
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, article)
}

// AdminList handles GET /admin/articles.
func (c *ArticleController) AdminList(ctx *gin.Context) {
	var q dto.ListQuery
	if err := ctx.ShouldBindQuery(&q); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	var moduleID *uint
	if m := ctx.Query("module_id"); m != "" {
		v, _ := strconv.ParseUint(m, 10, 32)
		u := uint(v)
		moduleID = &u
	}
	var status *int8
	if s := ctx.Query("status"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 8)
		t := int8(v)
		status = &t
	}
	articles, total, err := c.svc.AdminList(ctx, q.GetPage(), q.GetPageSize(), q.Keyword, moduleID, status)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.PaginatedSuccess(ctx, articles, total, q.GetPage(), q.GetPageSize())
}

// AdminGetByID handles GET /admin/articles/:id.
func (c *ArticleController) AdminGetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的文章ID", err))
		return
	}
	article, svcErr := c.svc.AdminGetByID(ctx, id)
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, article)
}

// AdminCreate handles POST /admin/articles.
func (c *ArticleController) AdminCreate(ctx *gin.Context) {
	var req dto.CreateArticleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	authorID, _ := middleware.GetCurrentUserID(ctx)
	id, err := c.svc.Create(ctx, &req, authorID)
	if err != nil {
		ctx.Error(err)
		return
	}
	response.SuccessWithStatus(ctx, http.StatusCreated, gin.H{"id": id})
}

// AdminUpdate handles PUT /admin/articles/:id.
func (c *ArticleController) AdminUpdate(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的文章ID", err))
		return
	}
	var req dto.UpdateArticleRequest
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

// AdminDelete handles DELETE /admin/articles/:id.
func (c *ArticleController) AdminDelete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的文章ID", err))
		return
	}
	if svcErr := c.svc.Delete(ctx, id); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// AdminPublish handles POST /admin/articles/:id/publish.
func (c *ArticleController) AdminPublish(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的文章ID", err))
		return
	}
	var req dto.PublishArticleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err := req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if svcErr := c.svc.Publish(ctx, id, &req); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, nil)
}

// AdminPin handles POST /admin/articles/:id/pin.
func (c *ArticleController) AdminPin(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的文章ID", err))
		return
	}
	var req dto.PinArticleRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperrors.NewBadRequest("参数绑定失败", err))
		return
	}
	if err = req.Validate(); err != nil {
		ctx.Error(apperrors.NewValidation("参数校验失败", err))
		return
	}
	if svcErr := c.svc.Pin(ctx, id, &req); svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.Success(ctx, nil)
}

// AdminCopy handles POST /admin/articles/:id/copy.
func (c *ArticleController) AdminCopy(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的文章ID", err))
		return
	}
	authorID, _ := middleware.GetCurrentUserID(ctx)
	newID, svcErr := c.svc.Copy(ctx, id, authorID)
	if svcErr != nil {
		ctx.Error(svcErr)
		return
	}
	response.SuccessWithStatus(ctx, http.StatusCreated, gin.H{"id": newID})
}
