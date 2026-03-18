package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

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
	hasAssociations, err := s.moduleRepo.HasAssociations(ctx, id)
	if err != nil {
		return err
	}
	if hasAssociations {
		return errors.NewBadRequest("模块存在关联内容，禁止删除", nil)
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
