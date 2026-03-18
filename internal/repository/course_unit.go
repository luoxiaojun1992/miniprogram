package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type courseUnitRepository struct {
	db *gorm.DB
}

// NewCourseUnitRepository creates a new CourseUnitRepository.
func NewCourseUnitRepository(db *gorm.DB) CourseUnitRepository {
	return &courseUnitRepository{db: db}
}

func (r *courseUnitRepository) GetByID(ctx context.Context, id uint64) (*entity.CourseUnit, error) {
	var u entity.CourseUnit
	res := r.db.WithContext(ctx).First(&u, id)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询课程单元失败", res.Error)
	}
	return &u, nil
}

func (r *courseUnitRepository) ListByCourseID(ctx context.Context, courseID uint64) ([]*entity.CourseUnit, error) {
	var units []*entity.CourseUnit
	if err := r.db.WithContext(ctx).Where("course_id = ?", courseID).Order("sort_order ASC").Find(&units).Error; err != nil {
		return nil, errors.NewInternal("查询课程单元失败", err)
	}
	return units, nil
}

func (r *courseUnitRepository) Create(ctx context.Context, unit *entity.CourseUnit) error {
	if err := r.db.WithContext(ctx).Create(unit).Error; err != nil {
		return errors.NewInternal("创建课程单元失败", err)
	}
	return nil
}

func (r *courseUnitRepository) Update(ctx context.Context, unit *entity.CourseUnit) error {
	if err := r.db.WithContext(ctx).Save(unit).Error; err != nil {
		return errors.NewInternal("更新课程单元失败", err)
	}
	return nil
}

func (r *courseUnitRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&entity.CourseUnit{}, id).Error; err != nil {
		return errors.NewInternal("删除课程单元失败", err)
	}
	return nil
}

func (r *courseUnitRepository) HasStudyRecords(ctx context.Context, id uint64) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.UserStudyRecord{}).Where("unit_id = ?", id).Count(&count).Error; err != nil {
		return false, errors.NewInternal("查询课程单元关联失败", err)
	}
	return count > 0, nil
}
