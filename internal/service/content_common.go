package service

import (
	"context"
	"reflect"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

func normalizeContentPermRepo(repo repository.ContentPermissionRepository) repository.ContentPermissionRepository {
	if repo == nil {
		return nil
	}
	rv := reflect.ValueOf(repo)
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return nil
	}
	return repo
}

func canAccessContentByRole(
	ctx context.Context,
	contentPermRepo repository.ContentPermissionRepository,
	roleRepo repository.RoleRepository,
	contentType int8,
	contentID uint64,
	userID *uint64,
) (bool, error) {
	if contentPermRepo == nil {
		return true, nil
	}
	perms, err := contentPermRepo.GetByContent(ctx, contentType, contentID)
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
	if len(allowedRoles) == 0 {
		return true, nil
	}
	if userID == nil || *userID == 0 || roleRepo == nil {
		return false, nil
	}
	roles, roleErr := roleRepo.GetUserRoles(ctx, *userID)
	if roleErr != nil {
		return false, roleErr
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
	allRoles, listErr := roleRepo.List(ctx)
	if listErr != nil {
		return false, listErr
	}
	if hasRoleHierarchyMatch(userRoleIDs, allowedRoles, allRoles) {
		return true, nil
	}
	return false, nil
}

func hasRoleHierarchyMatch(userRoleIDs, allowedRoles map[uint]struct{}, allRoles []*entity.Role) bool {
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
		if _, seen := visited[roleID]; seen {
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
