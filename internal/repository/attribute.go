package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

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

func (r *attributeRepository) HasUserAssociations(ctx context.Context, id uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT (
			(SELECT COUNT(1) FROM user_attributes WHERE attribute_id = ?) +
			(SELECT COUNT(1) FROM article_attributes WHERE attribute_id = ?) +
			(SELECT COUNT(1) FROM course_attributes WHERE attribute_id = ?)
		) AS cnt
	`, id, id, id).Scan(&count).Error; err != nil {
		return false, errors.NewInternal("查询属性关联失败", err)
	}
	return count > 0, nil
}
