package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

// ==================== Study Record Service ====================

type commentService struct {
	commentRepo       repository.CommentRepository
	articleRepo       repository.ArticleRepository
	courseRepo        repository.CourseRepository
	notifRepo         repository.NotificationRepository
	sensitiveWordRepo repository.SensitiveWordRepository
	log               *logrus.Logger
}

// NewCommentService creates a new CommentService.
func NewCommentService(
	commentRepo repository.CommentRepository,
	articleRepo repository.ArticleRepository,
	courseRepo repository.CourseRepository,
	notifRepo repository.NotificationRepository,
	log *logrus.Logger,
	sensitiveWordRepo ...repository.SensitiveWordRepository,
) CommentService {
	var swRepo repository.SensitiveWordRepository
	if len(sensitiveWordRepo) > 0 {
		swRepo = sensitiveWordRepo[0]
	}
	return &commentService{
		commentRepo:       commentRepo,
		articleRepo:       articleRepo,
		courseRepo:        courseRepo,
		notifRepo:         notifRepo,
		sensitiveWordRepo: swRepo,
		log:               log,
	}
}

func (s *commentService) List(ctx context.Context, contentType int8, contentID uint64, page, pageSize int) ([]*entity.Comment, int64, error) {
	return s.commentRepo.List(ctx, contentType, contentID, page, pageSize)
}

func (s *commentService) Create(ctx context.Context, userID uint64, contentType int8, contentID uint64, req *dto.CreateCommentRequest) (*entity.Comment, error) {
	if hasHTMLTag(req.Content) {
		return nil, errors.NewValidation("评论内容不允许包含HTML标签", nil)
	}
	words := loadSensitiveWords(ctx, s.sensitiveWordRepo, s.log)
	c := &entity.Comment{
		UserID:      userID,
		ContentType: contentType,
		ContentID:   contentID,
		ParentID:    req.ParentID,
		Content:     maskText(req.Content, words),
		Status:      1,
	}
	if err := s.commentRepo.Create(ctx, c); err != nil {
		return nil, err
	}
	switch contentType {
	case 1:
		if s.articleRepo != nil {
			if err := s.articleRepo.IncrCommentCount(ctx, contentID); err != nil {
				s.log.WithError(err).Warn("更新文章评论数失败")
			}
		}
	case 2:
		if s.courseRepo != nil {
			if err := s.courseRepo.IncrCommentCount(ctx, contentID); err != nil {
				s.log.WithError(err).Warn("更新课程评论数失败")
			}
		}
	}
	if s.notifRepo != nil {
		var targetUserID uint64
		if req.ParentID > 0 {
			parent, err := s.commentRepo.GetByID(ctx, req.ParentID)
			if err != nil {
				s.log.WithError(err).Warn("查询父评论失败")
			} else if parent != nil {
				targetUserID = parent.UserID
			}
		}
		if targetUserID == 0 {
			switch contentType {
			case 1:
				article, err := s.articleRepo.GetByID(ctx, contentID)
				if err != nil {
					s.log.WithError(err).Warn("查询文章失败")
				} else if article != nil {
					targetUserID = article.AuthorID
				}
			case 2:
				course, err := s.courseRepo.GetByID(ctx, contentID)
				if err != nil {
					s.log.WithError(err).Warn("查询课程失败")
				} else if course != nil {
					targetUserID = course.AuthorID
				}
			}
		}
		if targetUserID > 0 && targetUserID != userID {
			target := targetUserID
			if err := s.notifRepo.Create(ctx, &entity.Notification{
				UserID:  &target,
				Type:    2,
				Title:   "收到新的评论",
				Content: "你的内容收到新的评论或回复",
				IsRead:  0,
			}); err != nil {
				s.log.WithError(err).Warn("发送评论通知失败")
			}
		}
	}
	return c, nil
}

func (s *commentService) AdminList(ctx context.Context, page, pageSize int, status *int8) ([]*entity.Comment, int64, error) {
	return s.commentRepo.ListAdmin(ctx, page, pageSize, status)
}

func (s *commentService) Audit(ctx context.Context, id uint64, req *dto.AuditCommentRequest) error {
	c, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if c == nil {
		return errors.NewNotFound("评论不存在", nil)
	}
	return s.commentRepo.UpdateStatus(ctx, id, req.Status)
}

func (s *commentService) Delete(ctx context.Context, id uint64) error {
	c, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if c == nil {
		return errors.NewNotFound("评论不存在", nil)
	}
	hasReplies, err := s.commentRepo.HasReplies(ctx, id)
	if err != nil {
		return err
	}
	if hasReplies {
		return errors.NewBadRequest("评论存在回复，禁止删除", nil)
	}
	if err := s.commentRepo.Delete(ctx, id); err != nil {
		return err
	}
	switch c.ContentType {
	case 1:
		if s.articleRepo != nil {
			if err := s.articleRepo.DecrCommentCount(ctx, c.ContentID); err != nil {
				s.log.WithError(err).Warn("更新文章评论数失败")
			}
		}
	case 2:
		if s.courseRepo != nil {
			if err := s.courseRepo.DecrCommentCount(ctx, c.ContentID); err != nil {
				s.log.WithError(err).Warn("更新课程评论数失败")
			}
		}
	}
	return nil
}

// ==================== Notification Service ====================
