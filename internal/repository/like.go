package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

// ==================== ContentPermission Repository ====================

type likeRepository struct {
	db *gorm.DB
}

// NewLikeRepository creates a new LikeRepository.
func NewLikeRepository(db *gorm.DB) LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) Get(ctx context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Like, error) {
	var l entity.Like
	res := r.db.WithContext(ctx).Where("user_id = ? AND content_type = ? AND content_id = ?", userID, contentType, contentID).First(&l)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询点赞失败", res.Error)
	}
	return &l, nil
}

func (r *likeRepository) Create(ctx context.Context, like *entity.Like) error {
	if err := r.db.WithContext(ctx).Create(like).Error; err != nil {
		return errors.NewInternal("创建点赞失败", err)
	}
	return nil
}

func (r *likeRepository) Delete(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	if err := r.db.WithContext(ctx).Where("user_id = ? AND content_type = ? AND content_id = ?", userID, contentType, contentID).Delete(&entity.Like{}).Error; err != nil {
		return errors.NewInternal("删除点赞失败", err)
	}
	return nil
}

// ==================== Comment Repository ====================
