package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type followRepository struct {
	db *gorm.DB
}

// NewFollowRepository creates a new FollowRepository.
func NewFollowRepository(db *gorm.DB) FollowRepository {
	return &followRepository{db: db}
}

func (r *followRepository) Get(ctx context.Context, followerID, followedID uint64) (*entity.Follow, error) {
	var f entity.Follow
	res := r.db.WithContext(ctx).Where("follower_id = ? AND followed_id = ?", followerID, followedID).First(&f)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询关注关系失败", res.Error)
	}
	return &f, nil
}

func (r *followRepository) Create(ctx context.Context, follow *entity.Follow) error {
	if err := r.db.WithContext(ctx).Create(follow).Error; err != nil {
		return errors.NewInternal("创建关注失败", err)
	}
	return nil
}

func (r *followRepository) Delete(ctx context.Context, followerID, followedID uint64) error {
	if err := r.db.WithContext(ctx).
		Where("follower_id = ? AND followed_id = ?", followerID, followedID).
		Delete(&entity.Follow{}).Error; err != nil {
		return errors.NewInternal("取消关注失败", err)
	}
	return nil
}
