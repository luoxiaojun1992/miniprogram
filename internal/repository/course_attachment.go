package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type courseAttachmentRepository struct {
	db *gorm.DB
}

// NewCourseAttachmentRepository creates a new CourseAttachmentRepository.
func NewCourseAttachmentRepository(db *gorm.DB) CourseAttachmentRepository {
	return &courseAttachmentRepository{db: db}
}

func (r *courseAttachmentRepository) ListFileIDs(ctx context.Context, courseID uint64) ([]uint64, error) {
	rows, err := r.ListByCourseID(ctx, courseID)
	if err != nil {
		return nil, errors.NewInternal("查询课程附件失败", err)
	}
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.FileID)
	}
	return ids, nil
}

func (r *courseAttachmentRepository) ListByCourseID(ctx context.Context, courseID uint64) ([]*entity.CourseAttachment, error) {
	var rows []*entity.CourseAttachment
	if err := r.db.WithContext(ctx).Where("course_id = ?", courseID).Order("sort_order ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, errors.NewInternal("查询课程附件失败", err)
	}
	return rows, nil
}

func (r *courseAttachmentRepository) GetByFileID(ctx context.Context, fileID uint64) (*entity.CourseAttachment, error) {
	var row entity.CourseAttachment
	res := r.db.WithContext(ctx).Where("file_id = ?", fileID).Order("id DESC").First(&row)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询课程附件失败", res.Error)
	}
	return &row, nil
}

func (r *courseAttachmentRepository) Replace(ctx context.Context, courseID uint64, fileIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("course_id = ?", courseID).Delete(&entity.CourseAttachment{}).Error; err != nil {
			return errors.NewInternal("清理课程附件失败", err)
		}
		rows := make([]*entity.CourseAttachment, 0, len(fileIDs))
		seen := make(map[uint64]struct{}, len(fileIDs))
		for i, fileID := range fileIDs {
			if fileID == 0 {
				continue
			}
			if _, ok := seen[fileID]; ok {
				continue
			}
			seen[fileID] = struct{}{}
			rows = append(rows, &entity.CourseAttachment{
				CourseID:  courseID,
				FileID:    fileID,
				SortOrder: i,
			})
		}
		if len(rows) == 0 {
			return nil
		}
		if err := tx.Create(&rows).Error; err != nil {
			return errors.NewInternal("写入课程附件失败", err)
		}
		return nil
	})
}
