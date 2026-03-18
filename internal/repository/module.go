package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type moduleRepository struct {
	db *gorm.DB
}

// NewModuleRepository creates a new ModuleRepository.
func NewModuleRepository(db *gorm.DB) ModuleRepository {
	return &moduleRepository{db: db}
}

func (r *moduleRepository) GetByID(ctx context.Context, id uint) (*entity.Module, error) {
	var m entity.Module
	res := r.db.WithContext(ctx).First(&m, id)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询模块失败", res.Error)
	}
	return &m, nil
}

func (r *moduleRepository) List(ctx context.Context, status *int8) ([]*entity.Module, error) {
	db := r.db.WithContext(ctx)
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	var modules []*entity.Module
	if err := db.Order("sort_order ASC, created_at ASC").Find(&modules).Error; err != nil {
		return nil, errors.NewInternal("查询模块列表失败", err)
	}
	return modules, nil
}

func (r *moduleRepository) Create(ctx context.Context, module *entity.Module) error {
	if err := r.db.WithContext(ctx).Create(module).Error; err != nil {
		return errors.NewInternal("创建模块失败", err)
	}
	return nil
}

func (r *moduleRepository) Update(ctx context.Context, module *entity.Module) error {
	if err := r.db.WithContext(ctx).Save(module).Error; err != nil {
		return errors.NewInternal("更新模块失败", err)
	}
	return nil
}

func (r *moduleRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&entity.Module{}, id).Error; err != nil {
		return errors.NewInternal("删除模块失败", err)
	}
	return nil
}

func (r *moduleRepository) HasAssociations(ctx context.Context, id uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT (
			(SELECT COUNT(1) FROM module_pages WHERE module_id = ?) +
			(SELECT COUNT(1) FROM articles WHERE module_id = ?) +
			(SELECT COUNT(1) FROM courses WHERE module_id = ?)
		) AS cnt
	`, id, id, id).Scan(&count).Error; err != nil {
		return false, errors.NewInternal("查询模块关联失败", err)
	}
	return count > 0, nil
}

// ==================== ModulePage Repository ====================
