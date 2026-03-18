package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type modulePageRepository struct {
	db *gorm.DB
}

// NewModulePageRepository creates a new ModulePageRepository.
func NewModulePageRepository(db *gorm.DB) ModulePageRepository {
	return &modulePageRepository{db: db}
}

func (r *modulePageRepository) GetByID(ctx context.Context, id uint) (*entity.ModulePage, error) {
	var p entity.ModulePage
	res := r.db.WithContext(ctx).First(&p, id)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询模块页面失败", res.Error)
	}
	return &p, nil
}

func (r *modulePageRepository) ListByModuleID(ctx context.Context, moduleID uint) ([]*entity.ModulePage, error) {
	var pages []*entity.ModulePage
	if err := r.db.WithContext(ctx).Where("module_id = ?", moduleID).Order("sort_order ASC").Find(&pages).Error; err != nil {
		return nil, errors.NewInternal("查询模块页面失败", err)
	}
	return pages, nil
}

func (r *modulePageRepository) Create(ctx context.Context, page *entity.ModulePage) error {
	if err := r.db.WithContext(ctx).Create(page).Error; err != nil {
		return errors.NewInternal("创建模块页面失败", err)
	}
	return nil
}

func (r *modulePageRepository) Update(ctx context.Context, page *entity.ModulePage) error {
	if err := r.db.WithContext(ctx).Save(page).Error; err != nil {
		return errors.NewInternal("更新模块页面失败", err)
	}
	return nil
}

func (r *modulePageRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&entity.ModulePage{}, id).Error; err != nil {
		return errors.NewInternal("删除模块页面失败", err)
	}
	return nil
}
