package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type userTagRepository struct {
	db *gorm.DB
}

// NewUserTagRepository creates a new UserTagRepository.
func NewUserTagRepository(db *gorm.DB) UserTagRepository {
	return &userTagRepository{db: db}
}

func (r *userTagRepository) GetByUserID(ctx context.Context, userID uint64) ([]*entity.UserTag, error) {
	var tags []*entity.UserTag
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&tags).Error; err != nil {
		return nil, errors.NewInternal("查询用户标签失败", err)
	}
	return tags, nil
}

func (r *userTagRepository) Create(ctx context.Context, tag *entity.UserTag) error {
	if err := r.db.WithContext(ctx).Create(tag).Error; err != nil {
		return errors.NewInternal("创建标签失败", err)
	}
	return nil
}

func (r *userTagRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&entity.UserTag{}, id).Error; err != nil {
		return errors.NewInternal("删除标签失败", err)
	}
	return nil
}
