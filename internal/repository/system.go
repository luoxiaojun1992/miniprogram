package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

// ==================== WechatConfig Repository ====================

type wechatConfigRepository struct {
	db *gorm.DB
}

// NewWechatConfigRepository creates a new WechatConfigRepository.
func NewWechatConfigRepository(db *gorm.DB) WechatConfigRepository {
	return &wechatConfigRepository{db: db}
}

func (r *wechatConfigRepository) Get(ctx context.Context) (*entity.WechatConfig, error) {
	var cfg entity.WechatConfig
	res := r.db.WithContext(ctx).First(&cfg)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询微信配置失败", res.Error)
	}
	return &cfg, nil
}

func (r *wechatConfigRepository) Update(ctx context.Context, cfg *entity.WechatConfig) error {
	if err := r.db.WithContext(ctx).Save(cfg).Error; err != nil {
		return errors.NewInternal("更新微信配置失败", err)
	}
	return nil
}

// ==================== AuditLog Repository ====================

type auditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new AuditLogRepository.
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
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

type logConfigRepository struct {
	db *gorm.DB
}

// NewLogConfigRepository creates a new LogConfigRepository.
func NewLogConfigRepository(db *gorm.DB) LogConfigRepository {
	return &logConfigRepository{db: db}
}

func (r *logConfigRepository) Get(ctx context.Context) (*entity.LogConfig, error) {
	var cfg entity.LogConfig
	res := r.db.WithContext(ctx).First(&cfg)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询日志配置失败", res.Error)
	}
	return &cfg, nil
}

func (r *logConfigRepository) Update(ctx context.Context, cfg *entity.LogConfig) error {
	if err := r.db.WithContext(ctx).Save(cfg).Error; err != nil {
		return errors.NewInternal("更新日志配置失败", err)
	}
	return nil
}
