package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

// ==================== Attribute Repository ====================

type attributeRepository struct {
	db *gorm.DB
}

// NewAttributeRepository creates a new AttributeRepository.
func NewAttributeRepository(db *gorm.DB) AttributeRepository {
	return &attributeRepository{db: db}
}

func (r *attributeRepository) GetByID(ctx context.Context, id uint) (*entity.Attribute, error) {
	var a entity.Attribute
	res := r.db.WithContext(ctx).First(&a, id)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询属性失败", res.Error)
	}
	return &a, nil
}

func (r *attributeRepository) List(ctx context.Context) ([]*entity.Attribute, error) {
	var attrs []*entity.Attribute
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&attrs).Error; err != nil {
		return nil, errors.NewInternal("查询属性列表失败", err)
	}
	return attrs, nil
}

func (r *attributeRepository) Create(ctx context.Context, attr *entity.Attribute) error {
	if err := r.db.WithContext(ctx).Create(attr).Error; err != nil {
		return errors.NewInternal("创建属性失败", err)
	}
	return nil
}

func (r *attributeRepository) Update(ctx context.Context, attr *entity.Attribute) error {
	if err := r.db.WithContext(ctx).Save(attr).Error; err != nil {
		return errors.NewInternal("更新属性失败", err)
	}
	return nil
}

func (r *attributeRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&entity.Attribute{}, id).Error; err != nil {
		return errors.NewInternal("删除属性失败", err)
	}
	return nil
}

// ==================== UserAttribute Repository ====================

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
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
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
