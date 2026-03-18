package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type userAttributeRepository struct {
	db *gorm.DB
}

// NewUserAttributeRepository creates a new UserAttributeRepository.
func NewUserAttributeRepository(db *gorm.DB) UserAttributeRepository {
	return &userAttributeRepository{db: db}
}

func (r *userAttributeRepository) ListByUserID(ctx context.Context, userID uint64) ([]*entity.UserAttribute, error) {
	var uas []*entity.UserAttribute
	if err := r.db.WithContext(ctx).Preload("Attribute").Where("user_id = ?", userID).Find(&uas).Error; err != nil {
		return nil, errors.NewInternal("查询用户属性失败", err)
	}
	return uas, nil
}

func (r *userAttributeRepository) Upsert(ctx context.Context, ua *entity.UserAttribute) error {
	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "attribute_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"value_string", "value_bigint", "updated_at"}),
	}).Create(ua)
	if res.Error != nil {
		return errors.NewInternal("设置用户属性失败", res.Error)
	}
	return nil
}

func (r *userAttributeRepository) Delete(ctx context.Context, userID uint64, attributeID uint) error {
	res := r.db.WithContext(ctx).Where("user_id = ? AND attribute_id = ?", userID, attributeID).Delete(&entity.UserAttribute{})
	if res.Error != nil {
		return errors.NewInternal("删除用户属性失败", res.Error)
	}
	return nil
}
