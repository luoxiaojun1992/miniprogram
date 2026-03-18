package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

// ==================== WechatConfig Repository ====================

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
