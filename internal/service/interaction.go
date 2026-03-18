package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

type studyRecordService struct {
	studyRecordRepo repository.StudyRecordRepository
	courseUnitRepo  repository.CourseUnitRepository
	courseRepo      repository.CourseRepository
	log             *logrus.Logger
}

// NewStudyRecordService creates a new StudyRecordService.
func NewStudyRecordService(
	studyRecordRepo repository.StudyRecordRepository,
	courseUnitRepo repository.CourseUnitRepository,
	courseRepo repository.CourseRepository,
	log *logrus.Logger,
) StudyRecordService {
	return &studyRecordService{
		studyRecordRepo: studyRecordRepo,
		courseUnitRepo:  courseUnitRepo,
		courseRepo:      courseRepo,
		log:             log,
	}
}

func (s *studyRecordService) List(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.UserStudyRecord, int64, error) {
	return s.studyRecordRepo.ListByUser(ctx, userID, page, pageSize)
}

func (s *studyRecordService) Update(ctx context.Context, userID uint64, req *dto.UpdateStudyRecordRequest) error {
	unit, err := s.courseUnitRepo.GetByID(ctx, req.UnitID)
	if err != nil {
		return err
	}
	if unit == nil {
		return errors.NewNotFound("课程单元不存在", nil)
	}
	record := &entity.UserStudyRecord{
		UserID:   userID,
		CourseID: unit.CourseID,
		UnitID:   req.UnitID,
		Progress: req.Progress,
		Status:   req.Status,
	}
	existing, err := s.studyRecordRepo.GetByUserAndUnit(ctx, userID, req.UnitID)
	if err != nil {
		return err
	}
	if err := s.studyRecordRepo.Upsert(ctx, record); err != nil {
		return err
	}
	if existing == nil && s.courseRepo != nil {
		if err := s.courseRepo.IncrStudyCount(ctx, unit.CourseID); err != nil {
			s.log.WithError(err).Warn("更新课程学习人数失败")
		}
	}
	return nil
}
