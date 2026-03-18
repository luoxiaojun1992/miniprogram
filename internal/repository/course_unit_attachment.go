package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type courseUnitAttachmentRepository struct {
	db *gorm.DB
}

// NewCourseUnitAttachmentRepository creates a new CourseUnitAttachmentRepository.
func NewCourseUnitAttachmentRepository(db *gorm.DB) CourseUnitAttachmentRepository {
	return &courseUnitAttachmentRepository{db: db}
}

func (r *courseUnitAttachmentRepository) ListFileIDs(ctx context.Context, unitID uint64) ([]uint64, error) {
	rows, err := r.ListByUnitID(ctx, unitID)
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.FileID)
	}
	return ids, nil
}

func (r *courseUnitAttachmentRepository) ListByUnitID(ctx context.Context, unitID uint64) ([]*entity.CourseUnitAttachment, error) {
	var rows []*entity.CourseUnitAttachment
	if err := r.db.WithContext(ctx).Where("unit_id = ?", unitID).Order("sort_order ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, errors.NewInternal("查询课程单元附件失败", err)
	}
	return rows, nil
}

func (r *courseUnitAttachmentRepository) GetByFileID(ctx context.Context, fileID uint64) (*entity.CourseUnitAttachment, error) {
	var row entity.CourseUnitAttachment
	res := r.db.WithContext(ctx).Where("file_id = ?", fileID).Order("id DESC").First(&row)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询课程单元附件失败", res.Error)
	}
	return &row, nil
}

func (r *courseUnitAttachmentRepository) Replace(ctx context.Context, unitID uint64, fileIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("unit_id = ?", unitID).Delete(&entity.CourseUnitAttachment{}).Error; err != nil {
			return errors.NewInternal("清理课程单元附件失败", err)
		}
		rows := make([]*entity.CourseUnitAttachment, 0, len(fileIDs))
		seen := make(map[uint64]struct{}, len(fileIDs))
		for i, fileID := range fileIDs {
			if fileID == 0 {
				continue
			}
			if _, ok := seen[fileID]; ok {
				continue
			}
			seen[fileID] = struct{}{}
			rows = append(rows, &entity.CourseUnitAttachment{
				UnitID:    unitID,
				FileID:    fileID,
				SortOrder: i,
			})
		}
		if len(rows) == 0 {
			return nil
		}
		if err := tx.Create(&rows).Error; err != nil {
			return errors.NewInternal("写入课程单元附件失败", err)
		}
		return nil
	})
}
