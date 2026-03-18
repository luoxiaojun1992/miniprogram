package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

// ==================== ContentPermission Repository ====================

type studyRecordRepository struct {
	db *gorm.DB
}

// NewStudyRecordRepository creates a new StudyRecordRepository.
func NewStudyRecordRepository(db *gorm.DB) StudyRecordRepository {
	return &studyRecordRepository{db: db}
}

func (r *studyRecordRepository) GetByUserAndUnit(ctx context.Context, userID, unitID uint64) (*entity.UserStudyRecord, error) {
	var rec entity.UserStudyRecord
	res := r.db.WithContext(ctx).Where("user_id = ? AND unit_id = ?", userID, unitID).First(&rec)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询学习记录失败", res.Error)
	}
	return &rec, nil
}

func (r *studyRecordRepository) ListByUser(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.UserStudyRecord, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.UserStudyRecord{}).Where("user_id = ?", userID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternal("查询学习记录失败", err)
	}
	var records []*entity.UserStudyRecord
	if err := db.Order("updated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, errors.NewInternal("查询学习记录失败", err)
	}
	return records, total, nil
}

func (r *studyRecordRepository) Upsert(ctx context.Context, record *entity.UserStudyRecord) error {
	now := time.Now()
	record.LastStudyAt = &now
	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "unit_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"progress", "status", "last_study_at", "updated_at"}),
		}).Create(record)
	if res.Error != nil {
		return errors.NewInternal("更新学习记录失败", res.Error)
	}
	return nil
}

// ==================== Collection Repository ====================
