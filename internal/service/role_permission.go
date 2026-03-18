package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

type roleService struct {
	roleRepo repository.RoleRepository
	log      *logrus.Logger
}

// NewRoleService creates a new RoleService.
func NewRoleService(roleRepo repository.RoleRepository, log *logrus.Logger) RoleService {
	return &roleService{roleRepo: roleRepo, log: log}
}

func (s *roleService) List(ctx context.Context) ([]*entity.Role, error) {
	return s.roleRepo.List(ctx)
}

func (s *roleService) GetByID(ctx context.Context, id uint) (*entity.Role, error) {
	role, err := s.roleRepo.GetWithPermissions(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.NewNotFound("角色不存在", nil)
	}
	return role, nil
}

func (s *roleService) Create(ctx context.Context, req *dto.CreateRoleRequest) (uint, error) {
	level := int8(1)
	if req.ParentID > 0 {
		parent, err := s.roleRepo.GetByID(ctx, req.ParentID)
		if err != nil {
			return 0, err
		}
		if parent == nil {
			return 0, errors.NewBadRequest("父角色不存在", nil)
		}
		level = parent.Level + 1
		if level > 5 {
			return 0, errors.NewBadRequest("角色层级不能超过5层", nil)
		}
	}

	role := &entity.Role{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		Level:       level,
	}
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return 0, err
	}

	if len(req.PermissionIDs) > 0 {
		if err := s.roleRepo.AssignPermissions(ctx, role.ID, req.PermissionIDs); err != nil {
			return 0, err
		}
	}

	return role.ID, nil
}

func (s *roleService) Update(ctx context.Context, id uint, req *dto.UpdateRoleRequest) error {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.NewNotFound("角色不存在", nil)
	}

	role.Name = req.Name
	role.Description = req.Description

	if err = s.roleRepo.Update(ctx, role); err != nil {
		return err
	}

	return s.roleRepo.AssignPermissions(ctx, id, req.PermissionIDs)
}

func (s *roleService) Delete(ctx context.Context, id uint) error {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.NewNotFound("角色不存在", nil)
	}
	if role.IsBuiltin == 1 {
		return errors.NewForbidden("内置角色不可删除", nil)
	}
	hasUsers, err := s.roleRepo.HasUsers(ctx, id)
	if err != nil {
		return err
	}
	if hasUsers {
		return errors.NewBadRequest("该角色已分配给用户，请先解除关联", nil)
	}
	return s.roleRepo.Delete(ctx, id)
}

// ==================== Permission Service ====================
