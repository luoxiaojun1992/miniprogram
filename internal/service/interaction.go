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

type studyRecordService struct {
	studyRecordRepo repository.StudyRecordRepository
	courseUnitRepo  repository.CourseUnitRepository
	courseRepo      repository.CourseRepository
	log             *logrus.Logger
}

// NewStudyRecordService creates a new StudyRecordService.
func NewStudyRecordService(
	studyRecordRepo repository.StudyRecordRepository,
	courseUnitRepo repository.CourseUnitRepository,
	courseRepo repository.CourseRepository,
	log *logrus.Logger,
) StudyRecordService {
	return &studyRecordService{
		studyRecordRepo: studyRecordRepo,
		courseUnitRepo:  courseUnitRepo,
		courseRepo:      courseRepo,
		log:             log,
	}
}

func (s *studyRecordService) List(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.UserStudyRecord, int64, error) {
	return s.studyRecordRepo.ListByUser(ctx, userID, page, pageSize)
}

func (s *studyRecordService) Update(ctx context.Context, userID uint64, req *dto.UpdateStudyRecordRequest) error {
	unit, err := s.courseUnitRepo.GetByID(ctx, req.UnitID)
	if err != nil {
		return err
	}
	if unit == nil {
		return errors.NewNotFound("课程单元不存在", nil)
	}
	record := &entity.UserStudyRecord{
		UserID:   userID,
		CourseID: unit.CourseID,
		UnitID:   req.UnitID,
		Progress: req.Progress,
		Status:   req.Status,
	}
	existing, err := s.studyRecordRepo.GetByUserAndUnit(ctx, userID, req.UnitID)
	if err != nil {
		return err
	}
	if err := s.studyRecordRepo.Upsert(ctx, record); err != nil {
		return err
	}
	if existing == nil && s.courseRepo != nil {
		if err := s.courseRepo.IncrStudyCount(ctx, unit.CourseID); err != nil {
			s.log.WithError(err).Warn("更新课程学习人数失败")
		}
	}
	return nil
}

// ==================== Collection Service ====================

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

type likeService struct {
	likeRepo    repository.LikeRepository
	articleRepo repository.ArticleRepository
	courseRepo  repository.CourseRepository
	notifRepo   repository.NotificationRepository
	commentRepo repository.CommentRepository
	log         *logrus.Logger
}

// NewLikeService creates a new LikeService.
func NewLikeService(
	likeRepo repository.LikeRepository,
	articleRepo repository.ArticleRepository,
	courseRepo repository.CourseRepository,
	notifRepo repository.NotificationRepository,
	log *logrus.Logger,
) LikeService {
	return &likeService{
		likeRepo:    likeRepo,
		articleRepo: articleRepo,
		courseRepo:  courseRepo,
		notifRepo:   notifRepo,
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
	return nil
}

func (s *likeService) Remove(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	if err := s.likeRepo.Delete(ctx, userID, contentType, contentID); err != nil {
		return err
	}
	switch contentType {
	case 1:
		if s.articleRepo != nil {
			if err := s.articleRepo.DecrLikeCount(ctx, contentID); err != nil {
				s.log.WithError(err).Warn("更新文章点赞数失败")
			}
		}
	case 2:
		if s.courseRepo != nil {
			if err := s.courseRepo.DecrLikeCount(ctx, contentID); err != nil {
				s.log.WithError(err).Warn("更新课程点赞数失败")
			}
		}
	}
	return nil
}

// ==================== Comment Service ====================

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

type notificationService struct {
	notifRepo repository.NotificationRepository
	log       *logrus.Logger
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(notifRepo repository.NotificationRepository, log *logrus.Logger) NotificationService {
	return &notificationService{notifRepo: notifRepo, log: log}
}

func (s *notificationService) List(ctx context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, int64, error) {
	notifs, total, err := s.notifRepo.List(ctx, userID, page, pageSize, isRead)
	if err != nil {
		return nil, 0, 0, err
	}
	unread, err := s.notifRepo.UnreadCount(ctx, userID)
	if err != nil {
		return nil, 0, 0, err
	}
	return notifs, total, unread, nil
}

func (s *notificationService) MarkRead(ctx context.Context, id uint64) error {
	return s.notifRepo.MarkRead(ctx, id)
}

func (s *notificationService) MarkAllRead(ctx context.Context, userID uint64) error {
	return s.notifRepo.MarkAllRead(ctx, userID)
}

func (s *notificationService) Send(ctx context.Context, notification *entity.Notification) error {
	return s.notifRepo.Create(ctx, notification)
}

// ==================== WechatConfig Service ====================

type wechatConfigService struct {
	wechatConfigRepo repository.WechatConfigRepository
	log              *logrus.Logger
}

// NewWechatConfigService creates a new WechatConfigService.
func NewWechatConfigService(wechatConfigRepo repository.WechatConfigRepository, log *logrus.Logger) WechatConfigService {
	return &wechatConfigService{wechatConfigRepo: wechatConfigRepo, log: log}
}

func (s *wechatConfigService) Get(ctx context.Context) (*entity.WechatConfig, error) {
	cfg, err := s.wechatConfigRepo.Get(ctx)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return &entity.WechatConfig{}, nil
	}
	return cfg, nil
}

func (s *wechatConfigService) Update(ctx context.Context, req *dto.UpdateWechatConfigRequest) error {
	cfg, err := s.wechatConfigRepo.Get(ctx)
	if err != nil {
		return err
	}
	if cfg == nil {
		cfg = &entity.WechatConfig{}
	}
	cfg.AppID = req.AppID
	cfg.AppSecret = req.AppSecret
	if req.APIToken != "" {
		cfg.APIToken = req.APIToken
	}
	return s.wechatConfigRepo.Update(ctx, cfg)
}

// ==================== AuditLog Service ====================

type auditLogService struct {
	auditLogRepo repository.AuditLogRepository
	log          *logrus.Logger
}

// NewAuditLogService creates a new AuditLogService.
func NewAuditLogService(auditLogRepo repository.AuditLogRepository, log *logrus.Logger) AuditLogService {
	return &auditLogService{auditLogRepo: auditLogRepo, log: log}
}

func (s *auditLogService) List(ctx context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error) {
	return s.auditLogRepo.List(ctx, page, pageSize, module, action, startTime, endTime)
}

func (s *auditLogService) Log(ctx context.Context, log *entity.AuditLog) {
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		s.log.WithError(err).Warn("记录审计日志失败")
	}
}

// ==================== LogConfig Service ====================

type logConfigService struct {
	logConfigRepo repository.LogConfigRepository
	log           *logrus.Logger
}

// NewLogConfigService creates a new LogConfigService.
func NewLogConfigService(logConfigRepo repository.LogConfigRepository, log *logrus.Logger) LogConfigService {
	return &logConfigService{logConfigRepo: logConfigRepo, log: log}
}

func (s *logConfigService) Get(ctx context.Context) (*entity.LogConfig, error) {
	cfg, err := s.logConfigRepo.Get(ctx)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return &entity.LogConfig{RetentionDays: 90}, nil
	}
	return cfg, nil
}

func (s *logConfigService) Update(ctx context.Context, req *dto.UpdateLogConfigRequest) error {
	cfg, err := s.logConfigRepo.Get(ctx)
	if err != nil {
		return err
	}
	if cfg == nil {
		cfg = &entity.LogConfig{}
	}
	cfg.RetentionDays = req.RetentionDays
	return s.logConfigRepo.Update(ctx, cfg)
}
