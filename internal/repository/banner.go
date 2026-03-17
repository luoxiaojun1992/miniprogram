package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type bannerRepository struct {
	db *gorm.DB
}

// NewBannerRepository creates a new BannerRepository.
func NewBannerRepository(db *gorm.DB) BannerRepository {
	return &bannerRepository{db: db}
}

func (r *bannerRepository) GetByID(ctx context.Context, id uint64) (*entity.Banner, error) {
	var b entity.Banner
	res := r.db.WithContext(ctx).First(&b, id)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询轮播图失败", res.Error)
	}
	return &b, nil
}

func (r *bannerRepository) List(ctx context.Context, status *int8) ([]*entity.Banner, error) {
	db := r.db.WithContext(ctx).Model(&entity.Banner{})
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	var banners []*entity.Banner
	if err := db.Order("sort_order DESC, id DESC").Find(&banners).Error; err != nil {
		return nil, errors.NewInternal("查询轮播图列表失败", err)
	}
	return banners, nil
}

func (r *bannerRepository) Create(ctx context.Context, banner *entity.Banner) error {
	if err := r.db.WithContext(ctx).Create(banner).Error; err != nil {
		return errors.NewInternal("创建轮播图失败", err)
	}
	return nil
}

func (r *bannerRepository) Update(ctx context.Context, banner *entity.Banner) error {
	if err := r.db.WithContext(ctx).Save(banner).Error; err != nil {
		return errors.NewInternal("更新轮播图失败", err)
	}
	return nil
}

func (r *bannerRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&entity.Banner{}, id).Error; err != nil {
		return errors.NewInternal("删除轮播图失败", err)
	}
	return nil
}
