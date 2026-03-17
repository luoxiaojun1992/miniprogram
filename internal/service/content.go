package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

// ==================== Module Service ====================

type moduleService struct {
	moduleRepo     repository.ModuleRepository
	modulePageRepo repository.ModulePageRepository
	log            *logrus.Logger
}

// NewModuleService creates a new ModuleService.
func NewModuleService(
	moduleRepo repository.ModuleRepository,
	modulePageRepo repository.ModulePageRepository,
	log *logrus.Logger,
) ModuleService {
	return &moduleService{
		moduleRepo:     moduleRepo,
		modulePageRepo: modulePageRepo,
		log:            log,
	}
}

func (s *moduleService) List(ctx context.Context, status *int8) ([]*entity.Module, error) {
	return s.moduleRepo.List(ctx, status)
}

func (s *moduleService) Create(ctx context.Context, req *dto.CreateModuleRequest) (uint, error) {
	m := &entity.Module{
		Title:       req.Title,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		Status:      1,
	}
	if err := s.moduleRepo.Create(ctx, m); err != nil {
		return 0, err
	}
	return m.ID, nil
}

func (s *moduleService) Update(ctx context.Context, id uint, req *dto.CreateModuleRequest) error {
	m, err := s.moduleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return errors.NewNotFound("模块不存在", nil)
	}
	m.Title = req.Title
	m.Description = req.Description
	m.SortOrder = req.SortOrder
	return s.moduleRepo.Update(ctx, m)
}

func (s *moduleService) Delete(ctx context.Context, id uint) error {
	m, err := s.moduleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return errors.NewNotFound("模块不存在", nil)
	}
	return s.moduleRepo.Delete(ctx, id)
}

func (s *moduleService) GetPages(ctx context.Context, moduleID uint) ([]*entity.ModulePage, error) {
	return s.modulePageRepo.ListByModuleID(ctx, moduleID)
}

func (s *moduleService) CreatePage(ctx context.Context, moduleID uint, req *dto.CreateModulePageRequest) (uint, error) {
	p := &entity.ModulePage{
		ModuleID:    moduleID,
		Title:       req.Title,
		Content:     req.Content,
		ContentType: req.ContentType,
		SortOrder:   req.SortOrder,
		Status:      1,
	}
	if p.ContentType == 0 {
		p.ContentType = 1
	}
	if err := s.modulePageRepo.Create(ctx, p); err != nil {
		return 0, err
	}
	return p.ID, nil
}

func (s *moduleService) UpdatePage(ctx context.Context, moduleID, pageID uint, req *dto.CreateModulePageRequest) error {
	p, err := s.modulePageRepo.GetByID(ctx, pageID)
	if err != nil {
		return err
	}
	if p == nil || p.ModuleID != moduleID {
		return errors.NewNotFound("页面不存在", nil)
	}
	p.Title = req.Title
	p.Content = req.Content
	p.ContentType = req.ContentType
	p.SortOrder = req.SortOrder
	return s.modulePageRepo.Update(ctx, p)
}

func (s *moduleService) DeletePage(ctx context.Context, moduleID, pageID uint) error {
	p, err := s.modulePageRepo.GetByID(ctx, pageID)
	if err != nil {
		return err
	}
	if p == nil || p.ModuleID != moduleID {
		return errors.NewNotFound("页面不存在", nil)
	}
	return s.modulePageRepo.Delete(ctx, pageID)
}

// ==================== Article Service ====================

type articleService struct {
	articleRepo       repository.ArticleRepository
	contentPermRepo   repository.ContentPermissionRepository
	sensitiveWordRepo repository.SensitiveWordRepository
	log               *logrus.Logger
}

// NewArticleService creates a new ArticleService.
func NewArticleService(
	articleRepo repository.ArticleRepository,
	contentPermRepo repository.ContentPermissionRepository,
	log *logrus.Logger,
	sensitiveWordRepo ...repository.SensitiveWordRepository,
) ArticleService {
	var swRepo repository.SensitiveWordRepository
	if len(sensitiveWordRepo) > 0 {
		swRepo = sensitiveWordRepo[0]
	}
	return &articleService{
		articleRepo:       articleRepo,
		contentPermRepo:   contentPermRepo,
		sensitiveWordRepo: swRepo,
		log:               log,
	}
}

func (s *articleService) List(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, sort string, userID *uint64) ([]*entity.Article, int64, error) {
	status := int8(1)
	return s.articleRepo.List(ctx, page, pageSize, keyword, moduleID, &status, sort)
}

func (s *articleService) GetByID(ctx context.Context, id uint64, userID *uint64) (*entity.Article, error) {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if article == nil {
		return nil, errors.NewNotFound("文章不存在", nil)
	}
	if article.Status != 1 {
		return nil, errors.NewNotFound("文章不存在", nil)
	}
	go func() {
		_ = s.articleRepo.IncrViewCount(context.Background(), id)
	}()
	return article, nil
}

func (s *articleService) AdminList(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8) ([]*entity.Article, int64, error) {
	return s.articleRepo.List(ctx, page, pageSize, keyword, moduleID, status, "-created_at")
}

func (s *articleService) AdminGetByID(ctx context.Context, id uint64) (*entity.Article, error) {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if article == nil {
		return nil, errors.NewNotFound("文章不存在", nil)
	}
	return article, nil
}

func (s *articleService) Create(ctx context.Context, req *dto.CreateArticleRequest, authorID uint64) (uint64, error) {
	words := loadSensitiveWords(ctx, s.sensitiveWordRepo, s.log)
	article := &entity.Article{
		Title:       maskText(req.Title, words),
		Summary:     maskText(req.Summary, words),
		Content:     maskText(req.Content, words),
		ContentType: req.ContentType,
		CoverImage:  req.CoverImage,
		AuthorID:    authorID,
		ModuleID:    req.ModuleID,
		Status:      req.Status,
		PublishTime: req.PublishTime,
	}
	if article.ContentType == 0 {
		article.ContentType = 1
	}
	if article.Status == 1 && article.PublishTime == nil {
		now := time.Now()
		article.PublishTime = &now
	}
	if err := s.articleRepo.Create(ctx, article); err != nil {
		return 0, err
	}
	if len(req.RolePermissions) > 0 {
		if err := s.contentPermRepo.SetContentPermissions(ctx, 1, article.ID, req.RolePermissions); err != nil {
			s.log.WithError(err).Warn("设置文章权限失败")
		}
	}
	return article.ID, nil
}

func (s *articleService) Update(ctx context.Context, id uint64, req *dto.UpdateArticleRequest) error {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if article == nil {
		return errors.NewNotFound("文章不存在", nil)
	}
	words := loadSensitiveWords(ctx, s.sensitiveWordRepo, s.log)
	article.Title = maskText(req.Title, words)
	article.Summary = maskText(req.Summary, words)
	article.Content = maskText(req.Content, words)
	article.ContentType = req.ContentType
	article.CoverImage = req.CoverImage
	article.ModuleID = req.ModuleID
	article.Status = req.Status
	article.PublishTime = req.PublishTime
	if err = s.articleRepo.Update(ctx, article); err != nil {
		return err
	}
	return s.contentPermRepo.SetContentPermissions(ctx, 1, id, req.RolePermissions)
}

func (s *articleService) Delete(ctx context.Context, id uint64) error {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if article == nil {
		return errors.NewNotFound("文章不存在", nil)
	}
	return s.articleRepo.Delete(ctx, id)
}

func (s *articleService) Publish(ctx context.Context, id uint64, req *dto.PublishArticleRequest) error {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if article == nil {
		return errors.NewNotFound("文章不存在", nil)
	}
	article.Status = req.Status
	if req.Status == 1 {
		now := time.Now()
		article.PublishTime = &now
	}
	return s.articleRepo.Update(ctx, article)
}

func (s *articleService) Pin(ctx context.Context, id uint64, req *dto.PinArticleRequest) error {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if article == nil {
		return errors.NewNotFound("文章不存在", nil)
	}
	article.SortOrder = req.SortOrder
	return s.articleRepo.Update(ctx, article)
}

func (s *articleService) Copy(ctx context.Context, id uint64, authorID uint64) (uint64, error) {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}
	if article == nil {
		return 0, errors.NewNotFound("文章不存在", nil)
	}
	now := time.Now()
	dup := &entity.Article{
		Title:       fmt.Sprintf("%s-副本", article.Title),
		Summary:     article.Summary,
		Content:     article.Content,
		ContentType: article.ContentType,
		CoverImage:  article.CoverImage,
		AuthorID:    authorID,
		ModuleID:    article.ModuleID,
		Status:      0,
		PublishTime: nil,
		SortOrder:   article.SortOrder,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err = s.articleRepo.Create(ctx, dup); err != nil {
		return 0, err
	}
	roles, permErr := s.contentPermRepo.GetByContent(ctx, 1, id)
	if permErr == nil && len(roles) > 0 {
		roleIDs := make([]uint, 0, len(roles))
		for _, r := range roles {
			if r.RoleID != nil {
				roleIDs = append(roleIDs, *r.RoleID)
			}
		}
		if len(roleIDs) > 0 {
			_ = s.contentPermRepo.SetContentPermissions(ctx, 1, dup.ID, roleIDs)
		}
	}
	return dup.ID, nil
}

// ==================== Course Service ====================

type courseService struct {
	courseRepo        repository.CourseRepository
	courseUnitRepo    repository.CourseUnitRepository
	contentPermRepo   repository.ContentPermissionRepository
	sensitiveWordRepo repository.SensitiveWordRepository
	log               *logrus.Logger
}

// NewCourseService creates a new CourseService.
func NewCourseService(
	courseRepo repository.CourseRepository,
	courseUnitRepo repository.CourseUnitRepository,
	contentPermRepo repository.ContentPermissionRepository,
	log *logrus.Logger,
	sensitiveWordRepo ...repository.SensitiveWordRepository,
) CourseService {
	var swRepo repository.SensitiveWordRepository
	if len(sensitiveWordRepo) > 0 {
		swRepo = sensitiveWordRepo[0]
	}
	return &courseService{
		courseRepo:        courseRepo,
		courseUnitRepo:    courseUnitRepo,
		contentPermRepo:   contentPermRepo,
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
	go func() {
		_ = s.courseRepo.IncrViewCount(context.Background(), id)
	}()
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
	return course, nil
}

func (s *courseService) Create(ctx context.Context, req *dto.CreateCourseRequest, authorID uint64) (uint64, error) {
	words := loadSensitiveWords(ctx, s.sensitiveWordRepo, s.log)
	course := &entity.Course{
		Title:       maskText(req.Title, words),
		Description: maskText(req.Description, words),
		CoverImage:  req.CoverImage,
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
	if len(req.RolePermissions) > 0 {
		if err := s.contentPermRepo.SetContentPermissions(ctx, 2, course.ID, req.RolePermissions); err != nil {
			s.log.WithError(err).Warn("设置课程权限失败")
		}
	}
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
	course.Price = req.Price
	course.ModuleID = req.ModuleID
	course.Status = req.Status
	course.PublishTime = req.PublishTime
	if err = s.courseRepo.Update(ctx, course); err != nil {
		return err
	}
	return s.contentPermRepo.SetContentPermissions(ctx, 2, id, req.RolePermissions)
}

func (s *courseService) Delete(ctx context.Context, id uint64) error {
	course, err := s.courseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if course == nil {
		return errors.NewNotFound("课程不存在", nil)
	}
	return s.courseRepo.Delete(ctx, id)
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
		VideoURL:    course.VideoURL,
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
				CourseID:  dup.ID,
				Title:     unit.Title,
				VideoURL:  unit.VideoURL,
				Duration:  unit.Duration,
				SortOrder: unit.SortOrder,
				Status:    unit.Status,
			})
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
		CourseID:  courseID,
		Title:     req.Title,
		VideoURL:  req.VideoURL,
		Duration:  req.Duration,
		SortOrder: req.SortOrder,
		Status:    1,
	}
	if err := s.courseUnitRepo.Create(ctx, unit); err != nil {
		return 0, err
	}
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
	unit.VideoURL = req.VideoURL
	unit.Duration = req.Duration
	unit.SortOrder = req.SortOrder
	return s.courseUnitRepo.Update(ctx, unit)
}

func (s *courseService) DeleteUnit(ctx context.Context, courseID, unitID uint64) error {
	unit, err := s.courseUnitRepo.GetByID(ctx, unitID)
	if err != nil {
		return err
	}
	if unit == nil || unit.CourseID != courseID {
		return errors.NewNotFound("课程单元不存在", nil)
	}
	return s.courseUnitRepo.Delete(ctx, unitID)
}
