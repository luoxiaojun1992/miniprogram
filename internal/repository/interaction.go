package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type contentPermissionRepository struct {
	db *gorm.DB
}

// NewContentPermissionRepository creates a new ContentPermissionRepository.
func NewContentPermissionRepository(db *gorm.DB) ContentPermissionRepository {
	return &contentPermissionRepository{db: db}
}

func (r *contentPermissionRepository) GetByContent(ctx context.Context, contentType int8, contentID uint64) ([]*entity.ContentPermission, error) {
	var perms []*entity.ContentPermission
	if err := r.db.WithContext(ctx).Where("content_type = ? AND content_id = ?", contentType, contentID).Find(&perms).Error; err != nil {
		return nil, errors.NewInternal("查询内容权限失败", err)
	}
	return perms, nil
}

func (r *contentPermissionRepository) SetContentPermissions(ctx context.Context, contentType int8, contentID uint64, roleIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM content_permissions WHERE content_type = ? AND content_id = ?", contentType, contentID).Error; err != nil {
			return errors.NewInternal("清除内容权限失败", err)
		}
		for _, roleID := range roleIDs {
			rid := roleID
			cp := &entity.ContentPermission{
				ContentType: contentType,
				ContentID:   contentID,
				RoleID:      &rid,
			}
			if err := tx.Create(cp).Error; err != nil {
				return errors.NewInternal("设置内容权限失败", err)
			}
		}
		return nil
	})
}
