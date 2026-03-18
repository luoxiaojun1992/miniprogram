package controller

import (
	"context"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

type accessChecker struct {
	contentPermRepo repository.ContentPermissionRepository
	roleRepo        repository.RoleRepository
}

func newAccessChecker(contentPermRepo repository.ContentPermissionRepository, roleRepo repository.RoleRepository) *accessChecker {
	return &accessChecker{
		contentPermRepo: contentPermRepo,
		roleRepo:        roleRepo,
	}
}

func (c *accessChecker) canAccess(ctx context.Context, contentType int8, contentID uint64, userID *uint64, ownerID *uint64) (bool, error) {
	if c == nil || c.contentPermRepo == nil {
		return true, nil
	}
	perms, err := c.contentPermRepo.GetByContent(ctx, contentType, contentID)
	if err != nil {
		return false, err
	}
	if len(perms) == 0 {
		return true, nil
	}
	allowedRoles := map[uint]struct{}{}
	for _, perm := range perms {
		if perm.RoleID == nil {
			return true, nil
		}
		allowedRoles[*perm.RoleID] = struct{}{}
	}
	if ownerID != nil && userID != nil && *ownerID > 0 && *ownerID == *userID {
		return true, nil
	}
	if len(allowedRoles) == 0 {
		return true, nil
	}
	if c.roleRepo == nil || userID == nil || *userID == 0 {
		return false, nil
	}
	roles, err := c.roleRepo.GetUserRoles(ctx, *userID)
	if err != nil {
		return false, err
	}
	userRoleIDs := map[uint]struct{}{}
	for _, role := range roles {
		userRoleIDs[role.ID] = struct{}{}
	}
	for roleID := range userRoleIDs {
		if _, ok := allowedRoles[roleID]; ok {
			return true, nil
		}
	}
	allRoles, err := c.roleRepo.List(ctx)
	if err != nil {
		return false, err
	}
	return hasRoleHierarchyAccess(userRoleIDs, allowedRoles, allRoles), nil
}

func hasRoleHierarchyAccess(userRoleIDs, allowedRoles map[uint]struct{}, allRoles []*entity.Role) bool {
	parentByRole := make(map[uint]uint, len(allRoles))
	childrenByRole := make(map[uint][]uint, len(allRoles))
	for _, role := range allRoles {
		parentByRole[role.ID] = role.ParentID
		if role.ParentID > 0 {
			childrenByRole[role.ParentID] = append(childrenByRole[role.ParentID], role.ID)
		}
	}
	visited := map[uint]struct{}{}
	stack := make([]uint, 0, len(userRoleIDs))
	for roleID := range userRoleIDs {
		stack = append(stack, roleID)
	}
	for len(stack) > 0 {
		n := len(stack) - 1
		roleID := stack[n]
		stack = stack[:n]
		if _, ok := visited[roleID]; ok {
			continue
		}
		visited[roleID] = struct{}{}
		if _, ok := allowedRoles[roleID]; ok {
			return true
		}
		if parentID, ok := parentByRole[roleID]; ok && parentID > 0 {
			stack = append(stack, parentID)
		}
		stack = append(stack, childrenByRole[roleID]...)
	}
	return false
}
