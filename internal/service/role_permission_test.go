package service

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/testutil"
)

// ==================== RoleService ====================

func TestRoleService_List(t *testing.T) {
	roles := []*entity.Role{{ID: 1, Name: "admin"}}
	repo := &testutil.MockRoleRepository{
		ListFn: func(_ context.Context) ([]*entity.Role, error) {
			return roles, nil
		},
	}
	svc := NewRoleService(repo, logrus.New())
	got, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Equal(t, roles, got)
}

func TestRoleService_GetByID_Found(t *testing.T) {
	role := &entity.Role{ID: 1, Name: "admin"}
	repo := &testutil.MockRoleRepository{
		GetWithPermsFn: func(_ context.Context, id uint) (*entity.Role, error) {
			return role, nil
		},
	}
	svc := NewRoleService(repo, logrus.New())
	got, err := svc.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, role, got)
}

func TestRoleService_GetByID_NotFound(t *testing.T) {
	repo := &testutil.MockRoleRepository{
		GetWithPermsFn: func(_ context.Context, id uint) (*entity.Role, error) {
			return nil, nil
		},
	}
	svc := NewRoleService(repo, logrus.New())
	_, err := svc.GetByID(context.Background(), 1)
	require.Error(t, err)
}

func TestRoleService_Create_NoParent(t *testing.T) {
	repo := &testutil.MockRoleRepository{
		CreateFn: func(_ context.Context, r *entity.Role) error {
			r.ID = 1
			return nil
		},
	}
	svc := NewRoleService(repo, logrus.New())
	id, err := svc.Create(context.Background(), &dto.CreateRoleRequest{Name: "editor"})
	require.NoError(t, err)
	assert.Equal(t, uint(1), id)
}

func TestRoleService_Create_WithParent(t *testing.T) {
	parent := &entity.Role{ID: 1, Level: 1}
	repo := &testutil.MockRoleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Role, error) {
			return parent, nil
		},
		CreateFn: func(_ context.Context, r *entity.Role) error {
			r.ID = 2
			return nil
		},
	}
	svc := NewRoleService(repo, logrus.New())
	id, err := svc.Create(context.Background(), &dto.CreateRoleRequest{Name: "child", ParentID: 1})
	require.NoError(t, err)
	assert.Equal(t, uint(2), id)
}

func TestRoleService_Create_WithPermissions(t *testing.T) {
	repo := &testutil.MockRoleRepository{
		CreateFn: func(_ context.Context, r *entity.Role) error {
			r.ID = 1
			return nil
		},
		AssignPermsFn: func(_ context.Context, roleID uint, permissionIDs []uint) error {
			return nil
		},
	}
	svc := NewRoleService(repo, logrus.New())
	id, err := svc.Create(context.Background(), &dto.CreateRoleRequest{Name: "editor", PermissionIDs: []uint{1, 2}})
	require.NoError(t, err)
	assert.Equal(t, uint(1), id)
}

func TestRoleService_Create_ParentNotFound(t *testing.T) {
	repo := &testutil.MockRoleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Role, error) {
			return nil, nil
		},
	}
	svc := NewRoleService(repo, logrus.New())
	_, err := svc.Create(context.Background(), &dto.CreateRoleRequest{Name: "child", ParentID: 99})
	require.Error(t, err)
}

func TestRoleService_Create_ParentDBError(t *testing.T) {
	repo := &testutil.MockRoleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Role, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := NewRoleService(repo, logrus.New())
	_, err := svc.Create(context.Background(), &dto.CreateRoleRequest{Name: "child", ParentID: 1})
	require.Error(t, err)
}

func TestRoleService_Create_MaxLevel(t *testing.T) {
	parent := &entity.Role{ID: 1, Level: 5}
	repo := &testutil.MockRoleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Role, error) {
			return parent, nil
		},
	}
	svc := NewRoleService(repo, logrus.New())
	_, err := svc.Create(context.Background(), &dto.CreateRoleRequest{Name: "child", ParentID: 1})
	require.Error(t, err)
	appErr := err.(*apperrors.AppError)
	assert.Equal(t, 400001, appErr.Code)
}

func TestRoleService_Create_AssignPermsError(t *testing.T) {
	repo := &testutil.MockRoleRepository{
		CreateFn: func(_ context.Context, r *entity.Role) error {
			r.ID = 1
			return nil
		},
		AssignPermsFn: func(_ context.Context, roleID uint, permissionIDs []uint) error {
			return apperrors.NewInternal("db error", nil)
		},
	}
	svc := NewRoleService(repo, logrus.New())
	_, err := svc.Create(context.Background(), &dto.CreateRoleRequest{Name: "editor", PermissionIDs: []uint{1}})
	require.Error(t, err)
}

func TestRoleService_Update_Success(t *testing.T) {
	role := &entity.Role{ID: 1, Name: "old"}
	repo := &testutil.MockRoleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Role, error) {
			return role, nil
		},
		UpdateFn: func(_ context.Context, r *entity.Role) error {
			return nil
		},
		AssignPermsFn: func(_ context.Context, roleID uint, permissionIDs []uint) error {
			return nil
		},
	}
	svc := NewRoleService(repo, logrus.New())
	err := svc.Update(context.Background(), 1, &dto.UpdateRoleRequest{Name: "new", PermissionIDs: []uint{1}})
	require.NoError(t, err)
}

func TestRoleService_Update_NotFound(t *testing.T) {
	repo := &testutil.MockRoleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Role, error) {
			return nil, nil
		},
	}
	svc := NewRoleService(repo, logrus.New())
	err := svc.Update(context.Background(), 1, &dto.UpdateRoleRequest{Name: "new"})
	require.Error(t, err)
}

func TestRoleService_Update_DBError(t *testing.T) {
	repo := &testutil.MockRoleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Role, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := NewRoleService(repo, logrus.New())
	err := svc.Update(context.Background(), 1, &dto.UpdateRoleRequest{})
	require.Error(t, err)
}

func TestRoleService_Delete_Success(t *testing.T) {
	role := &entity.Role{ID: 1, IsBuiltin: 0}
	repo := &testutil.MockRoleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Role, error) {
			return role, nil
		},
		HasUsersFn: func(_ context.Context, roleID uint) (bool, error) {
			return false, nil
		},
		DeleteFn: func(_ context.Context, id uint) error {
			return nil
		},
	}
	svc := NewRoleService(repo, logrus.New())
	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestRoleService_Delete_NotFound(t *testing.T) {
	repo := &testutil.MockRoleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Role, error) {
			return nil, nil
		},
	}
	svc := NewRoleService(repo, logrus.New())
	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
}

func TestRoleService_Delete_Builtin(t *testing.T) {
	role := &entity.Role{ID: 1, IsBuiltin: 1}
	repo := &testutil.MockRoleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Role, error) {
			return role, nil
		},
	}
	svc := NewRoleService(repo, logrus.New())
	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
	appErr := err.(*apperrors.AppError)
	assert.Equal(t, 403001, appErr.Code)
}

func TestRoleService_Delete_HasUsers(t *testing.T) {
	role := &entity.Role{ID: 1, IsBuiltin: 0}
	repo := &testutil.MockRoleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Role, error) {
			return role, nil
		},
		HasUsersFn: func(_ context.Context, roleID uint) (bool, error) {
			return true, nil
		},
	}
	svc := NewRoleService(repo, logrus.New())
	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
	appErr := err.(*apperrors.AppError)
	assert.Equal(t, 400001, appErr.Code)
}

func TestRoleService_Delete_HasUsersError(t *testing.T) {
	role := &entity.Role{ID: 1, IsBuiltin: 0}
	repo := &testutil.MockRoleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Role, error) {
			return role, nil
		},
		HasUsersFn: func(_ context.Context, roleID uint) (bool, error) {
			return false, apperrors.NewInternal("db error", nil)
		},
	}
	svc := NewRoleService(repo, logrus.New())
	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
}

// ==================== PermissionService ====================

func TestPermissionService_GetTree(t *testing.T) {
	perms := []*entity.Permission{
		{ID: 1, ParentID: 0, Code: "user", Name: "User"},
		{ID: 2, ParentID: 1, Code: "user:list", Name: "List Users"},
	}
	repo := &testutil.MockPermissionRepository{
		ListFn: func(_ context.Context) ([]*entity.Permission, error) {
			return perms, nil
		},
	}
	svc := NewPermissionService(repo, logrus.New())
	tree, err := svc.GetTree(context.Background())
	require.NoError(t, err)
	assert.Len(t, tree, 1)
	assert.Len(t, tree[0].Children, 1)
}

func TestPermissionService_GetTree_Error(t *testing.T) {
	repo := &testutil.MockPermissionRepository{
		ListFn: func(_ context.Context) ([]*entity.Permission, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := NewPermissionService(repo, logrus.New())
	_, err := svc.GetTree(context.Background())
	require.Error(t, err)
}

func TestPermissionService_GetTree_Empty(t *testing.T) {
	repo := &testutil.MockPermissionRepository{
		ListFn: func(_ context.Context) ([]*entity.Permission, error) {
			return []*entity.Permission{}, nil
		},
	}
	svc := NewPermissionService(repo, logrus.New())
	tree, err := svc.GetTree(context.Background())
	require.NoError(t, err)
	assert.Empty(t, tree)
}
