package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new NotificationRepository.
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) GetByID(ctx context.Context, id uint64) (*entity.Notification, error) {
	var n entity.Notification
	res := r.db.WithContext(ctx).First(&n, id)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询通知失败", res.Error)
	}
	return &n, nil
}

func (r *notificationRepository) List(ctx context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.Notification{}).Where("user_id = ? OR user_id IS NULL", userID)
	if isRead != nil {
		v := int8(0)
		if *isRead {
			v = 1
		}
		db = db.Where("is_read = ?", v)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternal("查询通知失败", err)
	}
	var notifications []*entity.Notification
	if err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&notifications).Error; err != nil {
		return nil, 0, errors.NewInternal("查询通知失败", err)
	}
	return notifications, total, nil
}

func (r *notificationRepository) UnreadCount(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.Notification{}).
		Where("(user_id = ? OR user_id IS NULL) AND is_read = 0", userID).
		Count(&count).Error; err != nil {
		return 0, errors.NewInternal("查询未读通知数失败", err)
	}
	return count, nil
}

func (r *notificationRepository) MarkRead(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Model(&entity.Notification{}).Where("id = ?", id).Update("is_read", 1).Error; err != nil {
		return errors.NewInternal("标记已读失败", err)
	}
	return nil
}

func (r *notificationRepository) MarkAllRead(ctx context.Context, userID uint64) error {
	if err := r.db.WithContext(ctx).Model(&entity.Notification{}).
		Where("user_id = ? AND is_read = 0", userID).
		Update("is_read", 1).Error; err != nil {
		return errors.NewInternal("标记全部已读失败", err)
	}
	return nil
}
