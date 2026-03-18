package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

// ==================== Study Record Service ====================

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
