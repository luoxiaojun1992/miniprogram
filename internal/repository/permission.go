package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new PermissionRepository.
func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) List(ctx context.Context) ([]*entity.Permission, error) {
	var perms []*entity.Permission
	if err := r.db.WithContext(ctx).Find(&perms).Error; err != nil {
		return nil, errors.NewInternal("查询权限列表失败", err)
	}
	return perms, nil
}

func (r *permissionRepository) GetByID(ctx context.Context, id uint) (*entity.Permission, error) {
	var perm entity.Permission
	res := r.db.WithContext(ctx).First(&perm, id)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询权限失败", res.Error)
	}
	return &perm, nil
}

func (r *permissionRepository) GetUserPermissions(ctx context.Context, userID uint64) ([]*entity.Permission, error) {
	var perms []*entity.Permission
	err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT p.* FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = ?
	`, userID).Scan(&perms).Error
	if err != nil {
		return nil, errors.NewInternal("查询用户权限失败", err)
	}
	return perms, nil
}

func (r *permissionRepository) GetPermissionsByRoleIDs(ctx context.Context, roleIDs []uint) ([]*entity.Permission, error) {
	if len(roleIDs) == 0 {
		return []*entity.Permission{}, nil
	}
	var perms []*entity.Permission
	err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT p.* FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id IN ?
	`, roleIDs).Scan(&perms).Error
	if err != nil {
		return nil, errors.NewInternal("按角色查询权限失败", err)
	}
	return perms, nil
}
