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

// ArticleController handles article requests.
type ArticleController struct {
	svc           service.ArticleService
	log           *logrus.Logger
	accessChecker *accessChecker
	articleRepo   repository.ArticleRepository
	attachRepo    repository.ArticleAttachmentRepository
}

// NewArticleController creates a new ArticleController.
func NewArticleController(
	svc service.ArticleService,
	log *logrus.Logger,
	deps ...interface{},
) *ArticleController {
	var contentPermRepo repository.ContentPermissionRepository
	var roleRepo repository.RoleRepository
	var articleRepo repository.ArticleRepository
	var attachRepo repository.ArticleAttachmentRepository
	for _, dep := range deps {
		switch v := dep.(type) {
		case repository.ContentPermissionRepository:
			contentPermRepo = v
		case repository.RoleRepository:
			roleRepo = v
		case repository.ArticleRepository:
			articleRepo = v
		case repository.ArticleAttachmentRepository:
			attachRepo = v
		}
	}
	return &ArticleController{
		svc:           svc,
		log:           log,
		accessChecker: newAccessChecker(contentPermRepo, roleRepo),
		articleRepo:   articleRepo,
		attachRepo:    attachRepo,
	}
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
	filtered := make([]*entity.Article, 0, len(articles))
	for _, item := range articles {
		if moduleID == nil {
			filtered = append(filtered, item)
			continue
		}
		allowed, accessErr := c.accessChecker.canAccess(ctx, 3, uint64(item.ModuleID), uid, nil)
		if accessErr != nil {
			ctx.Error(accessErr)
			return
		}
		if allowed {
			filtered = append(filtered, item)
		}
	}
	response.PaginatedSuccess(ctx, filtered, total, q.GetPage(), q.GetPageSize())
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

// GetAttachments handles GET /articles/:id/attachments.
func (c *ArticleController) GetAttachments(ctx *gin.Context) {
	articleID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Error(apperrors.NewBadRequest("无效的文章ID", err))
		return
	}
	if c.articleRepo == nil || c.attachRepo == nil {
		ctx.Error(apperrors.NewBadRequest("文章附件仓储未初始化", nil))
		return
	}
	article, repoErr := c.articleRepo.GetByID(ctx, articleID)
	if repoErr != nil {
		ctx.Error(repoErr)
		return
	}
	if article == nil {
		ctx.Error(apperrors.NewNotFound("文章不存在", nil))
		return
	}
	userID, _ := middleware.GetCurrentUserID(ctx)
	var uid *uint64
	if userID > 0 {
		uid = &userID
	}
	allowed, accessErr := c.accessChecker.canAccess(ctx, 1, articleID, uid, &article.AuthorID)
	if accessErr != nil {
		ctx.Error(accessErr)
		return
	}
	if !allowed {
		ctx.Error(apperrors.NewForbidden("无权限访问该文章附件", nil))
		return
	}
	fileIDs, listErr := c.attachRepo.ListFileIDs(ctx, articleID)
	if listErr != nil {
		ctx.Error(listErr)
		return
	}
	response.Success(ctx, fileIDs)
}
