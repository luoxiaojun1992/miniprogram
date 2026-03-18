package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type adminUserRepository struct {
	db *gorm.DB
}

// NewAdminUserRepository creates a new AdminUserRepository.
func NewAdminUserRepository(db *gorm.DB) AdminUserRepository {
	return &adminUserRepository{db: db}
}

func (r *adminUserRepository) GetByEmail(ctx context.Context, email string) (*entity.AdminUser, error) {
	var a entity.AdminUser
	res := r.db.WithContext(ctx).Where("email = ?", email).First(&a)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询管理员失败", res.Error)
	}
	return &a, nil
}

func (r *adminUserRepository) GetByUserID(ctx context.Context, userID uint64) (*entity.AdminUser, error) {
	var a entity.AdminUser
	res := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&a)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询管理员失败", res.Error)
	}
	return &a, nil
}

func (r *adminUserRepository) Create(ctx context.Context, admin *entity.AdminUser) error {
	if err := r.db.WithContext(ctx).Create(admin).Error; err != nil {
		return errors.NewInternal("创建管理员失败", err)
	}
	return nil
}

func (r *adminUserRepository) UpdateLastLogin(ctx context.Context, id uint64) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&entity.AdminUser{}).Where("id = ?", id).Update("last_login_at", now).Error; err != nil {
		return errors.NewInternal("更新登录时间失败", err)
	}
	return nil
}

// ==================== UserTag Repository ====================
