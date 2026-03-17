package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(ctx context.Context, id uint64) (*entity.User, error) {
	var u entity.User
	res := r.db.WithContext(ctx).First(&u, id)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询用户失败", res.Error)
	}
	return &u, nil
}

func (r *userRepository) GetByOpenID(ctx context.Context, openID string) (*entity.User, error) {
	var u entity.User
	res := r.db.WithContext(ctx).Where("open_id = ?", openID).First(&u)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询用户失败", res.Error)
	}
	return &u, nil
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return errors.NewInternal("创建用户失败", err)
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return errors.NewInternal("更新用户失败", err)
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&entity.User{}, id).Error; err != nil {
		return errors.NewInternal("删除用户失败", err)
	}
	return nil
}

func (r *userRepository) List(ctx context.Context, page, pageSize int, keyword string, userType, status *int8) ([]*entity.User, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.User{})
	if keyword != "" {
		db = db.Where("nickname LIKE ?", "%"+keyword+"%")
	}
	if userType != nil {
		db = db.Where("user_type = ?", *userType)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternal("查询用户列表失败", err)
	}
	var users []*entity.User
	if err := db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, errors.NewInternal("查询用户列表失败", err)
	}
	return users, total, nil
}

func (r *userRepository) GetWithTags(ctx context.Context, id uint64) (*entity.User, error) {
	var u entity.User
	res := r.db.WithContext(ctx).Preload("Tags").First(&u, id)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询用户失败", res.Error)
	}
	return &u, nil
}

func (r *userRepository) HasAssociations(ctx context.Context, id uint64) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT (
			(SELECT COUNT(1) FROM admin_users WHERE user_id = ?) +
			(SELECT COUNT(1) FROM user_tags WHERE user_id = ?) +
			(SELECT COUNT(1) FROM user_roles WHERE user_id = ?) +
			(SELECT COUNT(1) FROM user_attributes WHERE user_id = ?) +
			(SELECT COUNT(1) FROM articles WHERE author_id = ?) +
			(SELECT COUNT(1) FROM courses WHERE author_id = ?) +
			(SELECT COUNT(1) FROM comments WHERE user_id = ?) +
			(SELECT COUNT(1) FROM likes WHERE user_id = ?) +
			(SELECT COUNT(1) FROM collections WHERE user_id = ?) +
			(SELECT COUNT(1) FROM user_study_records WHERE user_id = ?)
		) AS cnt
	`, id, id, id, id, id, id, id, id, id, id).Scan(&count).Error; err != nil {
		return false, errors.NewInternal("查询用户关联失败", err)
	}
	return count > 0, nil
}

// ==================== AdminUser Repository ====================

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

func (r *userTagRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&entity.UserTag{}, id).Error; err != nil {
		return errors.NewInternal("删除标签失败", err)
	}
	return nil
}
