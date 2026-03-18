package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

type articleService struct {
	articleRepo       repository.ArticleRepository
	attachmentRepo    repository.ArticleAttachmentRepository
	contentPermRepo   repository.ContentPermissionRepository
	roleRepo          repository.RoleRepository
	sensitiveWordRepo repository.SensitiveWordRepository
	log               *logrus.Logger
}

// NewArticleService creates a new ArticleService.
func NewArticleService(
	articleRepo repository.ArticleRepository,
	contentPermRepo repository.ContentPermissionRepository,
	log *logrus.Logger,
	deps ...interface{},
) ArticleService {
	var swRepo repository.SensitiveWordRepository
	var attachmentRepo repository.ArticleAttachmentRepository
	var roleRepo repository.RoleRepository
	for _, dep := range deps {
		switch v := dep.(type) {
		case repository.SensitiveWordRepository:
			swRepo = v
		case repository.ArticleAttachmentRepository:
			attachmentRepo = v
		case repository.RoleRepository:
			roleRepo = v
		}
	}
	return &articleService{
		articleRepo:       articleRepo,
		attachmentRepo:    attachmentRepo,
		contentPermRepo:   normalizeContentPermRepo(contentPermRepo),
		roleRepo:          roleRepo,
		sensitiveWordRepo: swRepo,
		log:               log,
	}
}

func (s *articleService) List(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, sort string, userID *uint64) ([]*entity.Article, int64, error) {
	status := int8(1)
	return s.articleRepo.List(ctx, page, pageSize, keyword, moduleID, &status, sort)
}

func (s *articleService) GetByID(ctx context.Context, id uint64, userID *uint64) (*entity.Article, error) {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if article == nil {
		return nil, errors.NewNotFound("文章不存在", nil)
	}
	if article.Status != 1 {
		return nil, errors.NewNotFound("文章不存在", nil)
	}
	allowed, accessErr := s.canAccessContent(ctx, 1, id, userID)
	if accessErr != nil {
		return nil, accessErr
	}
	if !allowed {
		return nil, errors.NewForbidden("无权限访问该内容", nil)
	}
	go func() {
		_ = s.articleRepo.IncrViewCount(context.Background(), id)
	}()
	s.bindArticleAttachmentIDs(ctx, article)
	return article, nil
}

func (s *articleService) AdminList(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8) ([]*entity.Article, int64, error) {
	return s.articleRepo.List(ctx, page, pageSize, keyword, moduleID, status, "-created_at")
}

func (s *articleService) AdminGetByID(ctx context.Context, id uint64) (*entity.Article, error) {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if article == nil {
		return nil, errors.NewNotFound("文章不存在", nil)
	}
	s.bindArticleAttachmentIDs(ctx, article)
	return article, nil
}

func (s *articleService) Create(ctx context.Context, req *dto.CreateArticleRequest, authorID uint64) (uint64, error) {
	words := loadSensitiveWords(ctx, s.sensitiveWordRepo, s.log)
	article := &entity.Article{
		Title:       maskText(req.Title, words),
		Summary:     maskText(req.Summary, words),
		Content:     maskText(req.Content, words),
		ContentType: req.ContentType,
		CoverImage:  req.CoverImage,
		AuthorID:    authorID,
		ModuleID:    req.ModuleID,
		Status:      req.Status,
		PublishTime: req.PublishTime,
	}
	if article.ContentType == 0 {
		article.ContentType = 1
	}
	if article.Status == 1 && article.PublishTime == nil {
		now := time.Now()
		article.PublishTime = &now
	}
	if err := s.articleRepo.Create(ctx, article); err != nil {
		return 0, err
	}
	if s.attachmentRepo != nil {
		if err := s.attachmentRepo.Replace(ctx, article.ID, req.AttachmentFileIDs); err != nil {
			return 0, err
		}
	}
	if len(req.RolePermissions) > 0 {
		if err := s.contentPermRepo.SetContentPermissions(ctx, 1, article.ID, req.RolePermissions); err != nil {
			s.log.WithError(err).Warn("设置文章权限失败")
		}
	}
	return article.ID, nil
}

func (s *articleService) Update(ctx context.Context, id uint64, req *dto.UpdateArticleRequest) error {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if article == nil {
		return errors.NewNotFound("文章不存在", nil)
	}
	words := loadSensitiveWords(ctx, s.sensitiveWordRepo, s.log)
	article.Title = maskText(req.Title, words)
	article.Summary = maskText(req.Summary, words)
	article.Content = maskText(req.Content, words)
	article.ContentType = req.ContentType
	article.CoverImage = req.CoverImage
	article.ModuleID = req.ModuleID
	article.Status = req.Status
	article.PublishTime = req.PublishTime
	if err = s.articleRepo.Update(ctx, article); err != nil {
		return err
	}
	if s.attachmentRepo != nil {
		if err := s.attachmentRepo.Replace(ctx, article.ID, req.AttachmentFileIDs); err != nil {
			return err
		}
	}
	return s.contentPermRepo.SetContentPermissions(ctx, 1, id, req.RolePermissions)
}

func (s *articleService) Delete(ctx context.Context, id uint64) error {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if article == nil {
		return errors.NewNotFound("文章不存在", nil)
	}
	hasAssociations, err := s.articleRepo.HasAssociations(ctx, id)
	if err != nil {
		return err
	}
	if hasAssociations {
		return errors.NewBadRequest("文章存在关联互动数据，禁止删除", nil)
	}
	return s.articleRepo.Delete(ctx, id)
}

func (s *articleService) Publish(ctx context.Context, id uint64, req *dto.PublishArticleRequest) error {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if article == nil {
		return errors.NewNotFound("文章不存在", nil)
	}
	article.Status = req.Status
	if req.Status == 1 {
		now := time.Now()
		article.PublishTime = &now
	}
	return s.articleRepo.Update(ctx, article)
}

func (s *articleService) Pin(ctx context.Context, id uint64, req *dto.PinArticleRequest) error {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if article == nil {
		return errors.NewNotFound("文章不存在", nil)
	}
	article.SortOrder = req.SortOrder
	return s.articleRepo.Update(ctx, article)
}

func (s *articleService) Copy(ctx context.Context, id uint64, authorID uint64) (uint64, error) {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}
	if article == nil {
		return 0, errors.NewNotFound("文章不存在", nil)
	}
	now := time.Now()
	dup := &entity.Article{
		Title:       fmt.Sprintf("%s-副本", article.Title),
		Summary:     article.Summary,
		Content:     article.Content,
		ContentType: article.ContentType,
		CoverImage:  article.CoverImage,
		AuthorID:    authorID,
		ModuleID:    article.ModuleID,
		Status:      0,
		PublishTime: nil,
		SortOrder:   article.SortOrder,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err = s.articleRepo.Create(ctx, dup); err != nil {
		return 0, err
	}
	if s.attachmentRepo != nil {
		attachmentIDs, listErr := s.attachmentRepo.ListFileIDs(ctx, id)
		if listErr == nil {
			_ = s.attachmentRepo.Replace(ctx, dup.ID, attachmentIDs)
		}
	}
	roles, permErr := s.contentPermRepo.GetByContent(ctx, 1, id)
	if permErr == nil && len(roles) > 0 {
		roleIDs := make([]uint, 0, len(roles))
		for _, r := range roles {
			if r.RoleID != nil {
				roleIDs = append(roleIDs, *r.RoleID)
			}
		}
		if len(roleIDs) > 0 {
			_ = s.contentPermRepo.SetContentPermissions(ctx, 1, dup.ID, roleIDs)
		}
	}
	return dup.ID, nil
}

func (s *articleService) bindArticleAttachmentIDs(ctx context.Context, article *entity.Article) {
	if article == nil || s.attachmentRepo == nil {
		return
	}
	ids, err := s.attachmentRepo.ListFileIDs(ctx, article.ID)
	if err == nil {
		article.AttachmentFileIDs = ids
	}
}

func (s *articleService) canAccessContent(ctx context.Context, contentType int8, contentID uint64, userID *uint64) (bool, error) {
	return canAccessContentByRole(ctx, s.contentPermRepo, s.roleRepo, contentType, contentID, userID)
}
