package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

type attributeService struct {
	attrRepo   repository.AttributeRepository
	uaRepo     repository.UserAttributeRepository
	userRepo   repository.UserRepository
	log        *logrus.Logger
}

// NewAttributeService creates a new AttributeService.
func NewAttributeService(
	attrRepo repository.AttributeRepository,
	uaRepo repository.UserAttributeRepository,
	userRepo repository.UserRepository,
	log *logrus.Logger,
) AttributeService {
	return &attributeService{
		attrRepo: attrRepo,
		uaRepo:   uaRepo,
		userRepo: userRepo,
		log:      log,
	}
}

func (s *attributeService) List(ctx context.Context) ([]*entity.Attribute, error) {
	return s.attrRepo.List(ctx)
}

func (s *attributeService) Create(ctx context.Context, req *dto.CreateAttributeRequest) (uint, error) {
	attr := &entity.Attribute{
		Name: req.Name,
	}
	if err := s.attrRepo.Create(ctx, attr); err != nil {
		return 0, err
	}
	return attr.ID, nil
}

func (s *attributeService) Update(ctx context.Context, id uint, req *dto.UpdateAttributeRequest) error {
	attr, err := s.attrRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if attr == nil {
		return errors.NewNotFound("属性不存在", nil)
	}
	attr.Name = req.Name
	return s.attrRepo.Update(ctx, attr)
}

func (s *attributeService) Delete(ctx context.Context, id uint) error {
	attr, err := s.attrRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if attr == nil {
		return errors.NewNotFound("属性不存在", nil)
	}
	return s.attrRepo.Delete(ctx, id)
}

func (s *attributeService) ListUserAttributes(ctx context.Context, userID uint64) ([]*entity.UserAttribute, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NewNotFound("用户不存在", nil)
	}
	return s.uaRepo.ListByUserID(ctx, userID)
}

func (s *attributeService) SetUserAttribute(ctx context.Context, userID uint64, req *dto.SetUserAttributeRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.NewNotFound("用户不存在", nil)
	}
	attr, err := s.attrRepo.GetByID(ctx, req.AttributeID)
	if err != nil {
		return err
	}
	if attr == nil {
		return errors.NewNotFound("属性不存在", nil)
	}
	ua := &entity.UserAttribute{
		UserID:      userID,
		AttributeID: req.AttributeID,
		Value:       req.Value,
	}
	return s.uaRepo.Upsert(ctx, ua)
}

func (s *attributeService) DeleteUserAttribute(ctx context.Context, userID uint64, attributeID uint) error {
	return s.uaRepo.Delete(ctx, userID, attributeID)
}
