package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

// ==================== WechatConfig Repository ====================

type auditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new AuditLogRepository.
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) GetByID(ctx context.Context, id uint64) (*entity.AuditLog, error) {
	var log entity.AuditLog
	res := r.db.WithContext(ctx).First(&log, id)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询审计日志失败", res.Error)
	}
	return &log, nil
}

func (r *auditLogRepository) List(ctx context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.AuditLog{})
	if module != "" {
		db = db.Where("module = ?", module)
	}
	if action != "" {
		db = db.Where("action = ?", action)
	}
	if startTime != nil {
		db = db.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		db = db.Where("created_at <= ?", *endTime)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternal("查询审计日志失败", err)
	}
	var logs []*entity.AuditLog
	if err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, errors.NewInternal("查询审计日志失败", err)
	}
	return logs, total, nil
}

func (r *auditLogRepository) Create(ctx context.Context, auditLog *entity.AuditLog) error {
	if err := r.db.WithContext(ctx).Create(auditLog).Error; err != nil {
		return errors.NewInternal("创建审计日志失败", err)
	}
	return nil
}

// ==================== LogConfig Repository ====================
