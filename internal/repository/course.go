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
	} else {
		db = db.Where("(module_id IS NULL OR module_id = 0)")
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
	if err := r.db.WithContext(ctx).Exec(`
		INSERT INTO course_attributes (course_id, attribute_id, value_bigint)
		SELECT ?, id, 1 FROM attributes WHERE name = 'view_count' LIMIT 1
		ON DUPLICATE KEY UPDATE value_bigint = COALESCE(value_bigint, 0) + 1
	`, id).Error; err != nil {
		return errors.NewInternal("更新浏览量失败", err)
	}
	return nil
}

func (r *courseRepository) IncrLikeCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec(`
		INSERT INTO course_attributes (course_id, attribute_id, value_bigint)
		SELECT ?, id, 1 FROM attributes WHERE name = 'like_count' LIMIT 1
		ON DUPLICATE KEY UPDATE value_bigint = COALESCE(value_bigint, 0) + 1
	`, id).Error; err != nil {
		return errors.NewInternal("更新点赞数失败", err)
	}
	return nil
}

func (r *courseRepository) DecrLikeCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec(`
		INSERT INTO course_attributes (course_id, attribute_id, value_bigint)
		SELECT ?, id, 0 FROM attributes WHERE name = 'like_count' LIMIT 1
		ON DUPLICATE KEY UPDATE value_bigint = GREATEST(0, COALESCE(value_bigint, 0) - 1)
	`, id).Error; err != nil {
		return errors.NewInternal("更新点赞数失败", err)
	}
	return nil
}

func (r *courseRepository) IncrCollectCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec(`
		INSERT INTO course_attributes (course_id, attribute_id, value_bigint)
		SELECT ?, id, 1 FROM attributes WHERE name = 'collect_count' LIMIT 1
		ON DUPLICATE KEY UPDATE value_bigint = COALESCE(value_bigint, 0) + 1
	`, id).Error; err != nil {
		return errors.NewInternal("更新收藏数失败", err)
	}
	return nil
}

func (r *courseRepository) DecrCollectCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec(`
		INSERT INTO course_attributes (course_id, attribute_id, value_bigint)
		SELECT ?, id, 0 FROM attributes WHERE name = 'collect_count' LIMIT 1
		ON DUPLICATE KEY UPDATE value_bigint = GREATEST(0, COALESCE(value_bigint, 0) - 1)
	`, id).Error; err != nil {
		return errors.NewInternal("更新收藏数失败", err)
	}
	return nil
}

func (r *courseRepository) IncrCommentCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec(`
		INSERT INTO course_attributes (course_id, attribute_id, value_bigint)
		SELECT ?, id, 1 FROM attributes WHERE name = 'comment_count' LIMIT 1
		ON DUPLICATE KEY UPDATE value_bigint = COALESCE(value_bigint, 0) + 1
	`, id).Error; err != nil {
		return errors.NewInternal("更新评论数失败", err)
	}
	return nil
}

func (r *courseRepository) DecrCommentCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec(`
		INSERT INTO course_attributes (course_id, attribute_id, value_bigint)
		SELECT ?, id, 0 FROM attributes WHERE name = 'comment_count' LIMIT 1
		ON DUPLICATE KEY UPDATE value_bigint = GREATEST(0, COALESCE(value_bigint, 0) - 1)
	`, id).Error; err != nil {
		return errors.NewInternal("更新评论数失败", err)
	}
	return nil
}

func (r *courseRepository) IncrShareCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec(`
		INSERT INTO course_attributes (course_id, attribute_id, value_bigint)
		SELECT ?, id, 1 FROM attributes WHERE name = 'share_count' LIMIT 1
		ON DUPLICATE KEY UPDATE value_bigint = COALESCE(value_bigint, 0) + 1
	`, id).Error; err != nil {
		return errors.NewInternal("更新分享数失败", err)
	}
	return nil
}

func (r *courseRepository) IncrStudyCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec(`
		INSERT INTO course_attributes (course_id, attribute_id, value_bigint)
		SELECT ?, id, 1 FROM attributes WHERE name = 'study_count' LIMIT 1
		ON DUPLICATE KEY UPDATE value_bigint = COALESCE(value_bigint, 0) + 1
	`, id).Error; err != nil {
		return errors.NewInternal("更新学习人数失败", err)
	}
	return nil
}

func (r *courseRepository) HasAssociations(ctx context.Context, id uint64) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT (
			(SELECT COUNT(1) FROM course_units WHERE course_id = ?) +
			(SELECT COUNT(1) FROM likes WHERE content_type = 2 AND content_id = ?) +
			(SELECT COUNT(1) FROM collections WHERE content_type = 2 AND content_id = ?) +
			(SELECT COUNT(1) FROM comments WHERE content_type = 2 AND content_id = ?) +
			(SELECT COUNT(1) FROM user_study_records WHERE course_id = ?)
		) AS cnt
	`, id, id, id, id, id).Scan(&count).Error; err != nil {
		return false, errors.NewInternal("查询课程关联失败", err)
	}
	return count > 0, nil
}

func (r *courseRepository) DeleteCascade(ctx context.Context, id uint64, fileIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			DELETE FROM content_permissions
			WHERE content_type = 7 AND content_id IN (
				SELECT cua.id
				FROM course_unit_attachments cua
				JOIN course_units cu ON cu.id = cua.unit_id
				WHERE cu.course_id = ?
			)
		`, id).Error; err != nil {
			return errors.NewInternal("删除课程单元附件权限失败", err)
		}
		if err := tx.Exec(`
			DELETE cua FROM course_unit_attachments cua
			JOIN course_units cu ON cu.id = cua.unit_id
			WHERE cu.course_id = ?
		`, id).Error; err != nil {
			return errors.NewInternal("删除课程单元附件关联失败", err)
		}
		if err := tx.Exec("DELETE FROM content_permissions WHERE content_type = 6 AND content_id IN (SELECT id FROM course_units WHERE course_id = ?)", id).Error; err != nil {
			return errors.NewInternal("删除课程单元权限失败", err)
		}
		if err := tx.Exec("DELETE FROM user_study_records WHERE course_id = ?", id).Error; err != nil {
			return errors.NewInternal("删除课程学习记录失败", err)
		}
		if err := tx.Exec("DELETE FROM course_units WHERE course_id = ?", id).Error; err != nil {
			return errors.NewInternal("删除课程单元失败", err)
		}
		if err := tx.Exec(`
			DELETE FROM content_permissions
			WHERE content_type = 5 AND content_id IN (
				SELECT id FROM course_attachments WHERE course_id = ?
			)
		`, id).Error; err != nil {
			return errors.NewInternal("删除课程附件权限失败", err)
		}
		if err := tx.Exec("DELETE FROM course_attachments WHERE course_id = ?", id).Error; err != nil {
			return errors.NewInternal("删除课程附件关联失败", err)
		}
		if err := tx.Exec("DELETE FROM content_permissions WHERE content_type = 2 AND content_id = ?", id).Error; err != nil {
			return errors.NewInternal("删除课程权限失败", err)
		}
		if err := tx.Exec("DELETE FROM course_attributes WHERE course_id = ?", id).Error; err != nil {
			return errors.NewInternal("删除课程属性失败", err)
		}
		if err := tx.Exec("DELETE FROM likes WHERE content_type = 2 AND content_id = ?", id).Error; err != nil {
			return errors.NewInternal("删除课程点赞记录失败", err)
		}
		if err := tx.Exec("DELETE FROM collections WHERE content_type = 2 AND content_id = ?", id).Error; err != nil {
			return errors.NewInternal("删除课程收藏记录失败", err)
		}
		if err := tx.Exec("DELETE FROM comments WHERE content_type = 2 AND content_id = ?", id).Error; err != nil {
			return errors.NewInternal("删除课程评论失败", err)
		}
		uniq := make(map[uint64]struct{}, len(fileIDs))
		for _, fileID := range fileIDs {
			if fileID == 0 {
				continue
			}
			if _, ok := uniq[fileID]; ok {
				continue
			}
			uniq[fileID] = struct{}{}
			if err := tx.Delete(&entity.File{}, fileID).Error; err != nil {
				return errors.NewInternal("删除文件记录失败", err)
			}
		}
		if err := tx.Delete(&entity.Course{}, id).Error; err != nil {
			return errors.NewInternal("删除课程失败", err)
		}
		return nil
	})
}
