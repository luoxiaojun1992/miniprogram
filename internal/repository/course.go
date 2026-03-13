package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type courseRepository struct {
	db *gorm.DB
}

// NewCourseRepository creates a new CourseRepository.
func NewCourseRepository(db *gorm.DB) CourseRepository {
	return &courseRepository{db: db}
}

func (r *courseRepository) GetByID(ctx context.Context, id uint64) (*entity.Course, error) {
	var c entity.Course
	res := r.db.WithContext(ctx).Preload("Author").Preload("Units").First(&c, id)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询课程失败", res.Error)
	}
	return &c, nil
}

func (r *courseRepository) List(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8, isFree *bool) ([]*entity.Course, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.Course{}).Preload("Author")
	if keyword != "" {
		db = db.Where("title LIKE ?", "%"+keyword+"%")
	}
	if moduleID != nil {
		db = db.Where("module_id = ?", *moduleID)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	if isFree != nil && *isFree {
		db = db.Where("price = 0")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternal("查询课程列表失败", err)
	}
	var courses []*entity.Course
	if err := db.Order("sort_order DESC, created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&courses).Error; err != nil {
		return nil, 0, errors.NewInternal("查询课程列表失败", err)
	}
	return courses, total, nil
}

func (r *courseRepository) Create(ctx context.Context, course *entity.Course) error {
	if err := r.db.WithContext(ctx).Create(course).Error; err != nil {
		return errors.NewInternal("创建课程失败", err)
	}
	return nil
}

func (r *courseRepository) Update(ctx context.Context, course *entity.Course) error {
	if err := r.db.WithContext(ctx).Save(course).Error; err != nil {
		return errors.NewInternal("更新课程失败", err)
	}
	return nil
}

func (r *courseRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&entity.Course{}, id).Error; err != nil {
		return errors.NewInternal("删除课程失败", err)
	}
	return nil
}

func (r *courseRepository) IncrViewCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec("UPDATE courses SET view_count = view_count + 1 WHERE id = ?", id).Error; err != nil {
		return errors.NewInternal("更新浏览量失败", err)
	}
	return nil
}

// ==================== CourseUnit Repository ====================

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
