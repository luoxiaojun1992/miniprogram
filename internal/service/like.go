package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

type likeService struct {
	likeRepo    repository.LikeRepository
	articleRepo repository.ArticleRepository
	courseRepo  repository.CourseRepository
	notifRepo   repository.NotificationRepository
	commentRepo repository.CommentRepository
	attrRepo    repository.AttributeRepository
	uaRepo      repository.UserAttributeRepository
	log         *logrus.Logger
}

// NewLikeService creates a new LikeService.
func NewLikeService(
	likeRepo repository.LikeRepository,
	articleRepo repository.ArticleRepository,
	courseRepo repository.CourseRepository,
	notifRepo repository.NotificationRepository,
	log *logrus.Logger,
	deps ...interface{},
) LikeService {
	var attrRepo repository.AttributeRepository
	var uaRepo repository.UserAttributeRepository
	for _, dep := range deps {
		switch v := dep.(type) {
		case repository.AttributeRepository:
			attrRepo = v
		case repository.UserAttributeRepository:
			uaRepo = v
		}
	}
	return &likeService{
		likeRepo:    likeRepo,
		articleRepo: articleRepo,
		courseRepo:  courseRepo,
		notifRepo:   notifRepo,
		attrRepo:    attrRepo,
		uaRepo:      uaRepo,
		log:         log,
	}
}

func (s *likeService) Add(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	existing, err := s.likeRepo.Get(ctx, userID, contentType, contentID)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.NewConflict("已点赞", nil)
	}
	l := &entity.Like{
		UserID:      userID,
		ContentType: contentType,
		ContentID:   contentID,
	}
	if err := s.likeRepo.Create(ctx, l); err != nil {
		return err
	}
	var targetUserID uint64
	switch contentType {
	case 1:
		if s.articleRepo != nil {
			if err := s.articleRepo.IncrLikeCount(ctx, contentID); err != nil {
				s.log.WithError(err).Warn("更新文章点赞数失败")
			}
			article, err := s.articleRepo.GetByID(ctx, contentID)
			if err != nil {
				s.log.WithError(err).Warn("查询文章失败")
			} else if article != nil {
				targetUserID = article.AuthorID
			}
		}
	case 2:
		if s.courseRepo != nil {
			if err := s.courseRepo.IncrLikeCount(ctx, contentID); err != nil {
				s.log.WithError(err).Warn("更新课程点赞数失败")
			}
			course, err := s.courseRepo.GetByID(ctx, contentID)
			if err != nil {
				s.log.WithError(err).Warn("查询课程失败")
			} else if course != nil {
				targetUserID = course.AuthorID
			}
		}
	}
	if s.notifRepo != nil && targetUserID > 0 && targetUserID != userID {
		target := targetUserID
		if err := s.notifRepo.Create(ctx, &entity.Notification{
			UserID:  &target,
			Type:    4,
			Title:   "收到新的点赞",
			Content: "你的内容收到一个新的点赞",
			IsRead:  0,
		}); err != nil {
			s.log.WithError(err).Warn("发送点赞通知失败")
		}
	}
	if targetUserID > 0 && targetUserID != userID {
		if err := s.incrUserLikedCount(ctx, targetUserID); err != nil {
			s.log.WithError(err).Warn("更新用户被点赞数失败")
		}
	}
	return nil
}

func (s *likeService) Remove(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	if err := s.likeRepo.Delete(ctx, userID, contentType, contentID); err != nil {
		return err
	}
	var targetUserID uint64
	switch contentType {
	case 1:
		if s.articleRepo != nil {
			if err := s.articleRepo.DecrLikeCount(ctx, contentID); err != nil {
				s.log.WithError(err).Warn("更新文章点赞数失败")
			}
			article, err := s.articleRepo.GetByID(ctx, contentID)
			if err != nil {
				s.log.WithError(err).Warn("查询文章失败")
			} else if article != nil {
				targetUserID = article.AuthorID
			}
		}
	case 2:
		if s.courseRepo != nil {
			if err := s.courseRepo.DecrLikeCount(ctx, contentID); err != nil {
				s.log.WithError(err).Warn("更新课程点赞数失败")
			}
			course, err := s.courseRepo.GetByID(ctx, contentID)
			if err != nil {
				s.log.WithError(err).Warn("查询课程失败")
			} else if course != nil {
				targetUserID = course.AuthorID
			}
		}
	}
	if targetUserID > 0 && targetUserID != userID {
		if err := s.decrUserLikedCount(ctx, targetUserID); err != nil {
			s.log.WithError(err).Warn("更新用户被点赞数失败")
		}
	}
	return nil
}

func (s *likeService) incrUserLikedCount(ctx context.Context, userID uint64) error {
	return s.adjustUserLikedCount(ctx, userID, 1)
}

func (s *likeService) decrUserLikedCount(ctx context.Context, userID uint64) error {
	return s.adjustUserLikedCount(ctx, userID, -1)
}

func (s *likeService) adjustUserLikedCount(ctx context.Context, userID uint64, delta int64) error {
	if s.attrRepo == nil || s.uaRepo == nil {
		return nil
	}
	attr, err := s.attrRepo.GetByName(ctx, "like_count")
	if err != nil {
		return err
	}
	if attr == nil {
		attr = &entity.Attribute{Name: "like_count", Type: entity.AttributeTypeBigInt}
		if err = s.attrRepo.Create(ctx, attr); err != nil {
			return err
		}
	}
	var current int64
	uas, err := s.uaRepo.ListByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, ua := range uas {
		if ua == nil || ua.AttributeID != attr.ID || ua.ValueBigint == nil {
			continue
		}
		current = *ua.ValueBigint
		break
	}
	next := current + delta
	if next < 0 {
		next = 0
	}
	return s.uaRepo.Upsert(ctx, &entity.UserAttribute{
		UserID:      userID,
		AttributeID: attr.ID,
		ValueBigint: &next,
	})
}
