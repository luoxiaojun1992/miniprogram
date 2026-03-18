package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var permissionByRoleColumns = []string{"id", "name", "code", "type", "parent_id", "level", "is_builtin", "created_at"}

func TestPermissionRepository_GetPermissionsByRoleIDs_Empty(t *testing.T) {
	db, _ := newTestDB(t)
	repo := NewPermissionRepository(db)

	perms, err := repo.GetPermissionsByRoleIDs(context.Background(), []uint{})
	require.NoError(t, err)
	assert.Empty(t, perms)
}

func TestPermissionRepository_GetPermissionsByRoleIDs_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewPermissionRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT DISTINCT p\\.\\* FROM permissions p").
		WillReturnRows(sqlmock.NewRows(permissionByRoleColumns).AddRow(1, "文章管理", "article:manage", 1, 0, 1, 0, now))

	perms, err := repo.GetPermissionsByRoleIDs(context.Background(), []uint{1, 2})
	require.NoError(t, err)
	require.Len(t, perms, 1)
	assert.Equal(t, uint(1), perms[0].ID)
}

func TestPermissionRepository_GetPermissionsByRoleIDs_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewPermissionRepository(db)

	mock.ExpectQuery("SELECT DISTINCT p\\.\\* FROM permissions p").
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetPermissionsByRoleIDs(context.Background(), []uint{1})
	assert.Error(t, err)
}
