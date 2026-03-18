package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

type permissionService struct {
	permRepo repository.PermissionRepository
	log      *logrus.Logger
}

// NewPermissionService creates a new PermissionService.
func NewPermissionService(permRepo repository.PermissionRepository, log *logrus.Logger) PermissionService {
	return &permissionService{permRepo: permRepo, log: log}
}

func (s *permissionService) GetTree(ctx context.Context) ([]*entity.Permission, error) {
	all, err := s.permRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	return buildPermissionTree(all, 0), nil
}

func buildPermissionTree(all []*entity.Permission, parentID uint) []*entity.Permission {
	var result []*entity.Permission
	for _, p := range all {
		if p.ParentID == parentID {
			p.Children = buildPermissionTree(all, p.ID)
			result = append(result, p)
		}
	}
	return result
}
