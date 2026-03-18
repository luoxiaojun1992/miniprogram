package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

// ==================== ContentPermission Repository ====================

type collectionRepository struct {
	db *gorm.DB
}

// NewCollectionRepository creates a new CollectionRepository.
func NewCollectionRepository(db *gorm.DB) CollectionRepository {
	return &collectionRepository{db: db}
}

func (r *collectionRepository) Get(ctx context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Collection, error) {
	var c entity.Collection
	res := r.db.WithContext(ctx).Where("user_id = ? AND content_type = ? AND content_id = ?", userID, contentType, contentID).First(&c)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询收藏失败", res.Error)
	}
	return &c, nil
}

func (r *collectionRepository) List(ctx context.Context, userID uint64, page, pageSize int, contentType *int8) ([]*entity.Collection, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.Collection{}).Where("user_id = ?", userID)
	if contentType != nil {
		db = db.Where("content_type = ?", *contentType)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternal("查询收藏列表失败", err)
	}
	var collections []*entity.Collection
	if err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&collections).Error; err != nil {
		return nil, 0, errors.NewInternal("查询收藏列表失败", err)
	}
	return collections, total, nil
}

func (r *collectionRepository) Create(ctx context.Context, collection *entity.Collection) error {
	if err := r.db.WithContext(ctx).Create(collection).Error; err != nil {
		return errors.NewInternal("创建收藏失败", err)
	}
	return nil
}

func (r *collectionRepository) Delete(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	if err := r.db.WithContext(ctx).Where("user_id = ? AND content_type = ? AND content_id = ?", userID, contentType, contentID).Delete(&entity.Collection{}).Error; err != nil {
		return errors.NewInternal("删除收藏失败", err)
	}
	return nil
}

// ==================== Like Repository ====================
