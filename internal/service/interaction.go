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
	log             *logrus.Logger
}

// NewStudyRecordService creates a new StudyRecordService.
func NewStudyRecordService(studyRecordRepo repository.StudyRecordRepository, courseUnitRepo repository.CourseUnitRepository, log *logrus.Logger) StudyRecordService {
	return &studyRecordService{studyRecordRepo: studyRecordRepo, courseUnitRepo: courseUnitRepo, log: log}
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
	return s.studyRecordRepo.Upsert(ctx, record)
}

// ==================== Collection Service ====================

type collectionService struct {
	collectionRepo repository.CollectionRepository
	log            *logrus.Logger
}

// NewCollectionService creates a new CollectionService.
func NewCollectionService(collectionRepo repository.CollectionRepository, log *logrus.Logger) CollectionService {
	return &collectionService{collectionRepo: collectionRepo, log: log}
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
	return s.collectionRepo.Create(ctx, c)
}

func (s *collectionService) Remove(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	return s.collectionRepo.Delete(ctx, userID, contentType, contentID)
}

// ==================== Like Service ====================

type likeService struct {
	likeRepo    repository.LikeRepository
	articleRepo repository.ArticleRepository
	courseRepo  repository.CourseRepository
	log         *logrus.Logger
}

// NewLikeService creates a new LikeService.
func NewLikeService(
	likeRepo repository.LikeRepository,
	articleRepo repository.ArticleRepository,
	courseRepo repository.CourseRepository,
	log *logrus.Logger,
) LikeService {
	return &likeService{
		likeRepo:    likeRepo,
		articleRepo: articleRepo,
		courseRepo:  courseRepo,
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
	return s.likeRepo.Create(ctx, l)
}

func (s *likeService) Remove(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	return s.likeRepo.Delete(ctx, userID, contentType, contentID)
}

// ==================== Comment Service ====================

type commentService struct {
	commentRepo repository.CommentRepository
	log         *logrus.Logger
}

// NewCommentService creates a new CommentService.
func NewCommentService(commentRepo repository.CommentRepository, log *logrus.Logger) CommentService {
	return &commentService{commentRepo: commentRepo, log: log}
}

func (s *commentService) List(ctx context.Context, contentType int8, contentID uint64, page, pageSize int) ([]*entity.Comment, int64, error) {
	return s.commentRepo.List(ctx, contentType, contentID, page, pageSize)
}

func (s *commentService) Create(ctx context.Context, userID uint64, contentType int8, contentID uint64, req *dto.CreateCommentRequest) (*entity.Comment, error) {
	c := &entity.Comment{
		UserID:      userID,
		ContentType: contentType,
		ContentID:   contentID,
		ParentID:    req.ParentID,
		Content:     req.Content,
		Status:      1,
	}
	if err := s.commentRepo.Create(ctx, c); err != nil {
		return nil, err
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
	return s.commentRepo.Delete(ctx, id)
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
