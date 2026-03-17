package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new RoleRepository.
func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) GetByID(ctx context.Context, id uint) (*entity.Role, error) {
	var role entity.Role
	res := r.db.WithContext(ctx).First(&role, id)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询角色失败", res.Error)
	}
	return &role, nil
}

func (r *roleRepository) GetWithPermissions(ctx context.Context, id uint) (*entity.Role, error) {
	var role entity.Role
	res := r.db.WithContext(ctx).Preload("Permissions").First(&role, id)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询角色失败", res.Error)
	}
	return &role, nil
}

func (r *roleRepository) List(ctx context.Context) ([]*entity.Role, error) {
	var roles []*entity.Role
	if err := r.db.WithContext(ctx).Find(&roles).Error; err != nil {
		return nil, errors.NewInternal("查询角色列表失败", err)
	}
	return roles, nil
}

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	if err := r.db.WithContext(ctx).Create(role).Error; err != nil {
		return errors.NewInternal("创建角色失败", err)
	}
	return nil
}

func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	if err := r.db.WithContext(ctx).Save(role).Error; err != nil {
		return errors.NewInternal("更新角色失败", err)
	}
	return nil
}

func (r *roleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", id).Error; err != nil {
			return errors.NewInternal("删除角色权限失败", err)
		}
		if err := tx.Exec("DELETE FROM user_roles WHERE role_id = ?", id).Error; err != nil {
			return errors.NewInternal("删除用户角色失败", err)
		}
		if err := tx.Delete(&entity.Role{}, id).Error; err != nil {
			return errors.NewInternal("删除角色失败", err)
		}
		return nil
	})
}

func (r *roleRepository) AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID).Error; err != nil {
			return errors.NewInternal("清除角色权限失败", err)
		}
		for _, pid := range permissionIDs {
			if err := tx.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, pid).Error; err != nil {
				return errors.NewInternal("分配权限失败", err)
			}
		}
		return nil
	})
}

func (r *roleRepository) GetUserRoles(ctx context.Context, userID uint64) ([]*entity.Role, error) {
	var roles []*entity.Role
	err := r.db.WithContext(ctx).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	if err != nil {
		return nil, errors.NewInternal("查询用户角色失败", err)
	}
	return roles, nil
}

func (r *roleRepository) AssignUserRoles(ctx context.Context, userID uint64, roleIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM user_roles WHERE user_id = ?", userID).Error; err != nil {
			return errors.NewInternal("清除用户角色失败", err)
		}
		for _, rid := range roleIDs {
			if err := tx.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, rid).Error; err != nil {
				return errors.NewInternal("分配角色失败", err)
			}
		}
		return nil
	})
}

func (r *roleRepository) HasUsers(ctx context.Context, roleID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM user_roles WHERE role_id = ?", roleID).Scan(&count).Error; err != nil {
		return false, errors.NewInternal("查询角色用户数失败", err)
	}
	return count > 0, nil
}

// ==================== Permission Repository ====================

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
