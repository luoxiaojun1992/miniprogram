package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

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
