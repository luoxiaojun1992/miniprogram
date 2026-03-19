package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

type courseService struct {
	courseRepo        repository.CourseRepository
	courseUnitRepo    repository.CourseUnitRepository
	unitAttachRepo    repository.CourseUnitAttachmentRepository
	attachmentRepo    repository.CourseAttachmentRepository
	fileRepo          repository.FileRepository
	fileObjectRemover fileObjectRemover
	contentPermRepo   repository.ContentPermissionRepository
	roleRepo          repository.RoleRepository
	sensitiveWordRepo repository.SensitiveWordRepository
	log               *logrus.Logger
}

// NewCourseService creates a new CourseService.
func NewCourseService(
	courseRepo repository.CourseRepository,
	courseUnitRepo repository.CourseUnitRepository,
	contentPermRepo repository.ContentPermissionRepository,
	log *logrus.Logger,
	deps ...interface{},
) CourseService {
	var swRepo repository.SensitiveWordRepository
	var attachmentRepo repository.CourseAttachmentRepository
	var unitAttachRepo repository.CourseUnitAttachmentRepository
	var roleRepo repository.RoleRepository
	var fileRepo repository.FileRepository
	var remover fileObjectRemover
	for _, dep := range deps {
		switch v := dep.(type) {
		case repository.SensitiveWordRepository:
			swRepo = v
		case repository.CourseAttachmentRepository:
			attachmentRepo = v
		case repository.CourseUnitAttachmentRepository:
			unitAttachRepo = v
		case repository.RoleRepository:
			roleRepo = v
		case repository.FileRepository:
			fileRepo = v
		case fileObjectRemover:
			remover = v
		}
	}
	return &courseService{
		courseRepo:        courseRepo,
		courseUnitRepo:    courseUnitRepo,
		unitAttachRepo:    unitAttachRepo,
		attachmentRepo:    attachmentRepo,
		fileRepo:          fileRepo,
		fileObjectRemover: remover,
		contentPermRepo:   normalizeContentPermRepo(contentPermRepo),
		roleRepo:          roleRepo,
		sensitiveWordRepo: swRepo,
		log:               log,
	}
}

func (s *courseService) List(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, isFree *bool, userID *uint64) ([]*entity.Course, int64, error) {
	status := int8(1)
	return s.courseRepo.List(ctx, page, pageSize, keyword, moduleID, &status, isFree)
}

func (s *courseService) GetByID(ctx context.Context, id uint64, userID *uint64) (*entity.Course, error) {
	course, err := s.courseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.NewNotFound("课程不存在", nil)
	}
	if course.Status != 1 {
		return nil, errors.NewNotFound("课程不存在", nil)
	}
	allowed, accessErr := s.canAccessContent(ctx, 2, id, userID, &course.AuthorID)
	if accessErr != nil {
		return nil, accessErr
	}
	if !allowed {
		return nil, errors.NewForbidden("无权限访问该内容", nil)
	}
	go func() {
		_ = s.courseRepo.IncrViewCount(context.Background(), id)
	}()
	s.bindCourseAttachmentIDs(ctx, course)
	return course, nil
}

func (s *courseService) AdminList(ctx context.Context, page, pageSize int, keyword string, status *int8) ([]*entity.Course, int64, error) {
	return s.courseRepo.List(ctx, page, pageSize, keyword, nil, status, nil)
}

func (s *courseService) AdminGetByID(ctx context.Context, id uint64) (*entity.Course, error) {
	course, err := s.courseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.NewNotFound("课程不存在", nil)
	}
	s.bindCourseAttachmentIDs(ctx, course)
	return course, nil
}

func (s *courseService) Create(ctx context.Context, req *dto.CreateCourseRequest, authorID uint64) (uint64, error) {
	words := loadSensitiveWords(ctx, s.sensitiveWordRepo, s.log)
	course := &entity.Course{
		Title:       maskText(req.Title, words),
		Description: maskText(req.Description, words),
		CoverImage:  req.CoverImage,
		CoverFileID: toOptionalUint64(req.CoverFileID),
		Price:       req.Price,
		AuthorID:    authorID,
		ModuleID:    req.ModuleID,
		Status:      req.Status,
		PublishTime: req.PublishTime,
	}
	if course.Status == 1 && course.PublishTime == nil {
		now := time.Now()
		course.PublishTime = &now
	}
	if err := s.courseRepo.Create(ctx, course); err != nil {
		return 0, err
	}
	if s.attachmentRepo != nil {
		if err := s.attachmentRepo.Replace(ctx, course.ID, req.AttachmentFileIDs); err != nil {
			return 0, err
		}
	}
	if len(req.RolePermissions) > 0 {
		if err := s.contentPermRepo.SetContentPermissions(ctx, 2, course.ID, req.RolePermissions); err != nil {
			s.log.WithError(err).Warn("设置课程权限失败")
		}
	}
	s.bindCourseAttachmentPermissions(ctx, course.ID, req.AttachmentPermissions)
	return course.ID, nil
}

func (s *courseService) Update(ctx context.Context, id uint64, req *dto.UpdateCourseRequest) error {
	course, err := s.courseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if course == nil {
		return errors.NewNotFound("课程不存在", nil)
	}
	words := loadSensitiveWords(ctx, s.sensitiveWordRepo, s.log)
	course.Title = maskText(req.Title, words)
	course.Description = maskText(req.Description, words)
	course.CoverImage = req.CoverImage
	course.CoverFileID = toOptionalUint64(req.CoverFileID)
	course.Price = req.Price
	course.ModuleID = req.ModuleID
	course.Status = req.Status
	course.PublishTime = req.PublishTime
	if err = s.courseRepo.Update(ctx, course); err != nil {
		return err
	}
	if s.attachmentRepo != nil {
		if err := s.attachmentRepo.Replace(ctx, course.ID, req.AttachmentFileIDs); err != nil {
			return err
		}
	}
	if err = s.contentPermRepo.SetContentPermissions(ctx, 2, id, req.RolePermissions); err != nil {
		return err
	}
	s.bindCourseAttachmentPermissions(ctx, id, req.AttachmentPermissions)
	return nil
}

func (s *courseService) Delete(ctx context.Context, id uint64) error {
	course, err := s.courseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if course == nil {
		return errors.NewNotFound("课程不存在", nil)
	}
	ids, keys, err := s.collectCourseDeleteFileIDsAndKeys(ctx, course.CoverFileID, id)
	if err != nil {
		return err
	}
	if err := s.courseRepo.DeleteCascade(ctx, id, ids); err != nil {
		return err
	}
	s.cleanupCOSObjects(ctx, keys)
	return nil
}

func (s *courseService) Publish(ctx context.Context, id uint64, req *dto.PublishCourseRequest) error {
	course, err := s.courseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if course == nil {
		return errors.NewNotFound("课程不存在", nil)
	}
	course.Status = req.Status
	if req.Status == 1 {
		now := time.Now()
		course.PublishTime = &now
	}
	return s.courseRepo.Update(ctx, course)
}

func (s *courseService) Pin(ctx context.Context, id uint64, req *dto.PinCourseRequest) error {
	course, err := s.courseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if course == nil {
		return errors.NewNotFound("课程不存在", nil)
	}
	course.SortOrder = req.SortOrder
	return s.courseRepo.Update(ctx, course)
}

func (s *courseService) Copy(ctx context.Context, id uint64, authorID uint64) (uint64, error) {
	course, err := s.courseRepo.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}
	if course == nil {
		return 0, errors.NewNotFound("课程不存在", nil)
	}
	dup := &entity.Course{
		Title:       fmt.Sprintf("%s-副本", course.Title),
		Description: course.Description,
		CoverImage:  course.CoverImage,
		Duration:    course.Duration,
		AuthorID:    authorID,
		ModuleID:    course.ModuleID,
		Status:      0,
		PublishTime: nil,
		Price:       course.Price,
		SortOrder:   course.SortOrder,
	}
	if err = s.courseRepo.Create(ctx, dup); err != nil {
		return 0, err
	}
	units, unitErr := s.courseUnitRepo.ListByCourseID(ctx, id)
	if unitErr == nil {
		for _, unit := range units {
			_ = s.courseUnitRepo.Create(ctx, &entity.CourseUnit{
				CourseID:    dup.ID,
				Title:       unit.Title,
				VideoFileID: unit.VideoFileID,
				Duration:    unit.Duration,
				SortOrder:   unit.SortOrder,
				Status:      unit.Status,
			})
		}
	}
	if s.attachmentRepo != nil {
		attachmentIDs, listErr := s.attachmentRepo.ListFileIDs(ctx, id)
		if listErr == nil {
			_ = s.attachmentRepo.Replace(ctx, dup.ID, attachmentIDs)
		}
	}
	roles, permErr := s.contentPermRepo.GetByContent(ctx, 2, id)
	if permErr == nil && len(roles) > 0 {
		roleIDs := make([]uint, 0, len(roles))
		for _, r := range roles {
			if r.RoleID != nil {
				roleIDs = append(roleIDs, *r.RoleID)
			}
		}
		if len(roleIDs) > 0 {
			_ = s.contentPermRepo.SetContentPermissions(ctx, 2, dup.ID, roleIDs)
		}
	}
	return dup.ID, nil
}

func (s *courseService) GetUnits(ctx context.Context, courseID uint64) ([]*entity.CourseUnit, error) {
	return s.courseUnitRepo.ListByCourseID(ctx, courseID)
}

func (s *courseService) CreateUnit(ctx context.Context, courseID uint64, req *dto.CreateCourseUnitRequest) (uint64, error) {
	unit := &entity.CourseUnit{
		CourseID:    courseID,
		Title:       req.Title,
		VideoFileID: toOptionalUint64(req.VideoFileID),
		Duration:    req.Duration,
		SortOrder:   req.SortOrder,
		Status:      1,
	}
	if err := s.courseUnitRepo.Create(ctx, unit); err != nil {
		return 0, err
	}
	if s.unitAttachRepo != nil {
		if err := s.unitAttachRepo.Replace(ctx, unit.ID, req.AttachmentFileIDs); err != nil {
			return 0, err
		}
	}
	if s.contentPermRepo != nil {
		if err := s.contentPermRepo.SetContentPermissions(ctx, 6, unit.ID, req.RolePermissions); err != nil {
			return 0, err
		}
	}
	s.bindUnitAttachmentPermissions(ctx, unit.ID, req.AttachmentPermissions)
	return unit.ID, nil
}

func (s *courseService) UpdateUnit(ctx context.Context, courseID, unitID uint64, req *dto.CreateCourseUnitRequest) error {
	unit, err := s.courseUnitRepo.GetByID(ctx, unitID)
	if err != nil {
		return err
	}
	if unit == nil || unit.CourseID != courseID {
		return errors.NewNotFound("课程单元不存在", nil)
	}
	unit.Title = req.Title
	unit.VideoFileID = toOptionalUint64(req.VideoFileID)
	unit.Duration = req.Duration
	unit.SortOrder = req.SortOrder
	if err = s.courseUnitRepo.Update(ctx, unit); err != nil {
		return err
	}
	if s.unitAttachRepo != nil {
		if err := s.unitAttachRepo.Replace(ctx, unit.ID, req.AttachmentFileIDs); err != nil {
			return err
		}
	}
	if s.contentPermRepo != nil {
		if err = s.contentPermRepo.SetContentPermissions(ctx, 6, unit.ID, req.RolePermissions); err != nil {
			return err
		}
	}
	s.bindUnitAttachmentPermissions(ctx, unit.ID, req.AttachmentPermissions)
	return nil
}

func (s *courseService) bindCourseAttachmentIDs(ctx context.Context, course *entity.Course) {
	if course == nil || s.attachmentRepo == nil {
		return
	}
	ids, err := s.attachmentRepo.ListFileIDs(ctx, course.ID)
	if err == nil {
		course.AttachmentFileIDs = ids
	}
}

func toOptionalUint64(v uint64) *uint64 {
	if v == 0 {
		return nil
	}
	return &v
}

func (s *courseService) canAccessContent(ctx context.Context, contentType int8, contentID uint64, userID *uint64, ownerID *uint64) (bool, error) {
	return canAccessContentByRole(ctx, s.contentPermRepo, s.roleRepo, contentType, contentID, userID, ownerID)
}

func (s *courseService) DeleteUnit(ctx context.Context, courseID, unitID uint64) error {
	unit, err := s.courseUnitRepo.GetByID(ctx, unitID)
	if err != nil {
		return err
	}
	if unit == nil || unit.CourseID != courseID {
		return errors.NewNotFound("课程单元不存在", nil)
	}
	ids, keys, err := s.collectUnitDeleteFileIDsAndKeys(ctx, unit.VideoFileID, unitID)
	if err != nil {
		return err
	}
	if err := s.courseUnitRepo.DeleteCascade(ctx, unitID, ids); err != nil {
		return err
	}
	s.cleanupCOSObjects(ctx, keys)
	return nil
}

func (s *courseService) bindCourseAttachmentPermissions(ctx context.Context, courseID uint64, reqs []dto.AttachmentPermissionRequest) {
	if s.attachmentRepo == nil || s.contentPermRepo == nil {
		return
	}
	rows, err := s.attachmentRepo.ListByCourseID(ctx, courseID)
	if err != nil {
		return
	}
	roleMap := make(map[uint64][]uint, len(reqs))
	for _, req := range reqs {
		roleMap[req.FileID] = req.RolePermissions
	}
	for _, row := range rows {
		_ = s.contentPermRepo.SetContentPermissions(ctx, 5, row.ID, roleMap[row.FileID])
	}
}

func (s *courseService) bindUnitAttachmentPermissions(ctx context.Context, unitID uint64, reqs []dto.AttachmentPermissionRequest) {
	if s.unitAttachRepo == nil || s.contentPermRepo == nil {
		return
	}
	rows, err := s.unitAttachRepo.ListByUnitID(ctx, unitID)
	if err != nil {
		return
	}
	roleMap := make(map[uint64][]uint, len(reqs))
	for _, req := range reqs {
		roleMap[req.FileID] = req.RolePermissions
	}
	for _, row := range rows {
		_ = s.contentPermRepo.SetContentPermissions(ctx, 7, row.ID, roleMap[row.FileID])
	}
}

func (s *courseService) collectCourseDeleteFileIDsAndKeys(ctx context.Context, coverFileID *uint64, courseID uint64) ([]uint64, []string, error) {
	ids := make([]uint64, 0, 16)
	if coverFileID != nil && *coverFileID > 0 {
		ids = append(ids, *coverFileID)
	}
	if s.attachmentRepo != nil {
		attachmentIDs, err := s.attachmentRepo.ListFileIDs(ctx, courseID)
		if err != nil {
			return nil, nil, err
		}
		ids = append(ids, attachmentIDs...)
	}
	if s.courseUnitRepo != nil {
		units, err := s.courseUnitRepo.ListByCourseID(ctx, courseID)
		if err != nil {
			return nil, nil, err
		}
		for _, unit := range units {
			if unit == nil {
				continue
			}
			if unit.VideoFileID != nil && *unit.VideoFileID > 0 {
				ids = append(ids, *unit.VideoFileID)
			}
			if s.unitAttachRepo == nil {
				continue
			}
			unitAttachmentIDs, attachErr := s.unitAttachRepo.ListFileIDs(ctx, unit.ID)
			if attachErr != nil {
				return nil, nil, attachErr
			}
			ids = append(ids, unitAttachmentIDs...)
		}
	}
	return s.resolveFileDeletionTargets(ctx, ids)
}

func (s *courseService) collectUnitDeleteFileIDsAndKeys(ctx context.Context, videoFileID *uint64, unitID uint64) ([]uint64, []string, error) {
	ids := make([]uint64, 0, 8)
	if videoFileID != nil && *videoFileID > 0 {
		ids = append(ids, *videoFileID)
	}
	if s.unitAttachRepo != nil {
		attachmentIDs, err := s.unitAttachRepo.ListFileIDs(ctx, unitID)
		if err != nil {
			return nil, nil, err
		}
		ids = append(ids, attachmentIDs...)
	}
	return s.resolveFileDeletionTargets(ctx, ids)
}

func (s *courseService) resolveFileDeletionTargets(ctx context.Context, ids []uint64) ([]uint64, []string, error) {
	uniq := make(map[uint64]struct{}, len(ids))
	resolvedIDs := make([]uint64, 0, len(ids))
	keys := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := uniq[id]; ok {
			continue
		}
		uniq[id] = struct{}{}
		resolvedIDs = append(resolvedIDs, id)
		if s.fileRepo == nil {
			continue
		}
		file, err := s.fileRepo.GetByID(ctx, id)
		if err != nil {
			return nil, nil, err
		}
		if file == nil {
			continue
		}
		key := strings.TrimSpace(file.Key)
		if key != "" {
			keys = append(keys, key)
		}
	}
	return resolvedIDs, keys, nil
}

func (s *courseService) cleanupCOSObjects(ctx context.Context, keys []string) {
	if s.fileObjectRemover == nil || len(keys) == 0 {
		return
	}
	uniq := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		k := strings.TrimSpace(key)
		if k == "" {
			continue
		}
		if _, ok := uniq[k]; ok {
			continue
		}
		uniq[k] = struct{}{}
		if err := s.fileObjectRemover.DeleteObject(ctx, k); err != nil {
			s.log.WithError(err).Warn("删除COS文件失败")
		}
	}
}
