package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

const (
	contentTypeArticle           int8 = 1
	contentTypeArticleAttachment int8 = 4
)

type articleService struct {
	articleRepo       repository.ArticleRepository
	attachmentRepo    repository.ArticleAttachmentRepository
	contentPermRepo   repository.ContentPermissionRepository
	fileRepo          repository.FileRepository
	fileObjectRemover fileObjectRemover
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
	var fileRepo repository.FileRepository
	var remover fileObjectRemover
	for _, dep := range deps {
		switch v := dep.(type) {
		case repository.SensitiveWordRepository:
			swRepo = v
		case repository.ArticleAttachmentRepository:
			attachmentRepo = v
		case repository.RoleRepository:
			roleRepo = v
		case repository.FileRepository:
			fileRepo = v
		case fileObjectRemover:
			remover = v
		}
	}
	return &articleService{
		articleRepo:       articleRepo,
		attachmentRepo:    attachmentRepo,
		contentPermRepo:   normalizeContentPermRepo(contentPermRepo),
		fileRepo:          fileRepo,
		fileObjectRemover: remover,
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
	allowed, accessErr := s.canAccessContent(ctx, 1, id, userID, &article.AuthorID)
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
		ContentType: 1,
		CoverImage:  req.CoverImage,
		CoverFileID: toOptionalUint64(req.CoverFileID),
		AuthorID:    authorID,
		ModuleID:    req.ModuleID,
		Status:      req.Status,
		PublishTime: req.PublishTime,
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
		if err := s.contentPermRepo.SetContentPermissions(ctx, contentTypeArticle, article.ID, req.RolePermissions); err != nil {
			s.log.WithError(err).Warn("设置文章权限失败")
		}
	}
	s.bindAttachmentPermissions(ctx, article.ID, req.AttachmentPermissions)
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
	article.ContentType = 1
	article.CoverImage = req.CoverImage
	article.CoverFileID = toOptionalUint64(req.CoverFileID)
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
	if err = s.contentPermRepo.SetContentPermissions(ctx, contentTypeArticle, id, req.RolePermissions); err != nil {
		return err
	}
	s.bindAttachmentPermissions(ctx, id, req.AttachmentPermissions)
	return nil
}

func (s *articleService) Delete(ctx context.Context, id uint64) error {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if article == nil {
		return errors.NewNotFound("文章不存在", nil)
	}
	ids, keys, err := s.collectDeleteFileIDsAndKeys(ctx, article.CoverFileID, id)
	if err != nil {
		return err
	}
	if err := s.articleRepo.DeleteCascade(ctx, id, ids); err != nil {
		return err
	}
	s.cleanupCOSObjects(ctx, keys)
	return nil
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

func (s *articleService) bindAttachmentPermissions(ctx context.Context, articleID uint64, reqs []dto.AttachmentPermissionRequest) {
	if s.attachmentRepo == nil || s.contentPermRepo == nil {
		return
	}
	rows, err := s.attachmentRepo.ListByArticleID(ctx, articleID)
	if err != nil {
		return
	}
	roleMap := make(map[uint64][]uint, len(reqs))
	for _, req := range reqs {
		roleMap[req.FileID] = req.RolePermissions
	}
	for _, row := range rows {
		_ = s.contentPermRepo.SetContentPermissions(ctx, contentTypeArticleAttachment, row.ID, roleMap[row.FileID])
	}
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

func (s *articleService) canAccessContent(ctx context.Context, contentType int8, contentID uint64, userID *uint64, ownerID *uint64) (bool, error) {
	return canAccessContentByRole(ctx, s.contentPermRepo, s.roleRepo, contentType, contentID, userID, ownerID)
}

type fileObjectRemover interface {
	DeleteObject(ctx context.Context, key string) error
}

func (s *articleService) collectDeleteFileIDsAndKeys(ctx context.Context, coverFileID *uint64, articleID uint64) ([]uint64, []string, error) {
	ids := make([]uint64, 0, 8)
	if coverFileID != nil && *coverFileID > 0 {
		ids = append(ids, *coverFileID)
	}
	if s.attachmentRepo != nil {
		attachmentIDs, err := s.attachmentRepo.ListFileIDs(ctx, articleID)
		if err != nil {
			return nil, nil, err
		}
		ids = append(ids, attachmentIDs...)
	}
	return s.resolveFileDeletionTargets(ctx, ids)
}

func (s *articleService) resolveFileDeletionTargets(ctx context.Context, ids []uint64) ([]uint64, []string, error) {
	uniq := make(map[uint64]struct{}, len(ids))
	resolvedIDs := make([]uint64, 0, len(ids))
	keys := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := uniq[id]; ok {
			continue
		}
		uniq[id] = struct{}{}
		resolvedIDs = append(resolvedIDs, id)
		if s.fileRepo == nil {
			continue
		}
		file, err := s.fileRepo.GetByID(ctx, id)
		if err != nil {
			return nil, nil, err
		}
		if file == nil {
			continue
		}
		key := strings.TrimSpace(file.Key)
		if key != "" {
			keys = append(keys, key)
		}
	}
	return resolvedIDs, keys, nil
}

func (s *articleService) cleanupCOSObjects(ctx context.Context, keys []string) {
	if s.fileObjectRemover == nil || len(keys) == 0 {
		return
	}
	uniq := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		k := strings.TrimSpace(key)
		if k == "" {
			continue
		}
		if _, ok := uniq[k]; ok {
			continue
		}
		uniq[k] = struct{}{}
		if err := s.fileObjectRemover.DeleteObject(ctx, k); err != nil {
			s.log.WithError(err).Warn("删除COS文件失败")
		}
	}
}
