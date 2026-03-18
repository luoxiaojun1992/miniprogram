package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

// ==================== Study Record Service ====================

type collectionService struct {
	collectionRepo repository.CollectionRepository
	articleRepo    repository.ArticleRepository
	courseRepo     repository.CourseRepository
	log            *logrus.Logger
}

// NewCollectionService creates a new CollectionService.
func NewCollectionService(
	collectionRepo repository.CollectionRepository,
	articleRepo repository.ArticleRepository,
	courseRepo repository.CourseRepository,
	log *logrus.Logger,
) CollectionService {
	return &collectionService{
		collectionRepo: collectionRepo,
		articleRepo:    articleRepo,
		courseRepo:     courseRepo,
		log:            log,
	}
}

func (s *collectionService) List(ctx context.Context, userID uint64, page, pageSize int, contentType *int8) ([]*entity.Collection, int64, error) {
	return s.collectionRepo.List(ctx, userID, page, pageSize, contentType)
}

func (s *collectionService) Add(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	existing, err := s.collectionRepo.Get(ctx, userID, contentType, contentID)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.NewConflict("已收藏", nil)
	}
	c := &entity.Collection{
		UserID:      userID,
		ContentType: contentType,
		ContentID:   contentID,
	}
	if err := s.collectionRepo.Create(ctx, c); err != nil {
		return err
	}
	switch contentType {
	case 1:
		if s.articleRepo != nil {
			if err := s.articleRepo.IncrCollectCount(ctx, contentID); err != nil {
				s.log.WithError(err).Warn("更新文章收藏数失败")
			}
		}
	case 2:
		if s.courseRepo != nil {
			if err := s.courseRepo.IncrCollectCount(ctx, contentID); err != nil {
				s.log.WithError(err).Warn("更新课程收藏数失败")
			}
		}
	}
	return nil
}

func (s *collectionService) Remove(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	if err := s.collectionRepo.Delete(ctx, userID, contentType, contentID); err != nil {
		return err
	}
	switch contentType {
	case 1:
		if s.articleRepo != nil {
			if err := s.articleRepo.DecrCollectCount(ctx, contentID); err != nil {
				s.log.WithError(err).Warn("更新文章收藏数失败")
			}
		}
	case 2:
		if s.courseRepo != nil {
			if err := s.courseRepo.DecrCollectCount(ctx, contentID); err != nil {
				s.log.WithError(err).Warn("更新课程收藏数失败")
			}
		}
	}
	return nil
}

// ==================== Like Service ====================
