package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
)

var roleColumns = []string{"id", "name", "description", "created_at", "updated_at"}
var permissionColumns = []string{"id", "name", "code", "parent_id", "created_at", "updated_at"}

// ==================== RoleRepository ====================

func TestRoleRepository_GetByID_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(roleColumns).AddRow(1, "admin", "Admin Role", nil, nil),
	)

	r, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, r)
}

func TestRoleRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(roleColumns))

	r, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, r)
}

func TestRoleRepository_GetByID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestRoleRepository_GetWithPermissions_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	// Preload("Permissions") → 2 queries
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(roleColumns).AddRow(1, "admin", "Admin Role", nil, nil),
	)
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(permissionColumns))

	r, err := repo.GetWithPermissions(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, r)
}

func TestRoleRepository_GetWithPermissions_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(roleColumns))

	r, err := repo.GetWithPermissions(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, r)
}

func TestRoleRepository_GetWithPermissions_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetWithPermissions(context.Background(), 1)
	assert.Error(t, err)
}

func TestRoleRepository_List_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(roleColumns).
			AddRow(1, "admin", "Admin", nil, nil).
			AddRow(2, "editor", "Editor", nil, nil),
	)

	roles, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, roles, 2)
}

func TestRoleRepository_List_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.List(context.Background())
	assert.Error(t, err)
}

func TestRoleRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), &entity.Role{Name: "tester"})
	require.NoError(t, err)
}

func TestRoleRepository_Create_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), &entity.Role{Name: "fail"})
	assert.Error(t, err)
}

func TestRoleRepository_Update_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(context.Background(), &entity.Role{ID: 1, Name: "updated"})
	require.NoError(t, err)
}

func TestRoleRepository_Update_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("update error"))
	mock.ExpectRollback()

	err := repo.Update(context.Background(), &entity.Role{ID: 1})
	assert.Error(t, err)
}

func TestRoleRepository_Delete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM role_permissions").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM user_roles").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestRoleRepository_Delete_RolePermError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM role_permissions").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestRoleRepository_Delete_UserRoleError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM role_permissions").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM user_roles").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestRoleRepository_Delete_RoleError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM role_permissions").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM user_roles").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE").WillReturnError(fmt.Errorf("delete role error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestRoleRepository_AssignPermissions_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM role_permissions").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO role_permissions").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO role_permissions").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	err := repo.AssignPermissions(context.Background(), 1, []uint{10, 20})
	require.NoError(t, err)
}

func TestRoleRepository_AssignPermissions_DeleteError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM role_permissions").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.AssignPermissions(context.Background(), 1, []uint{1})
	assert.Error(t, err)
}

func TestRoleRepository_AssignPermissions_InsertError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM role_permissions").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO role_permissions").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.AssignPermissions(context.Background(), 1, []uint{1})
	assert.Error(t, err)
}

func TestRoleRepository_GetUserRoles_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(roleColumns).AddRow(1, "editor", "Editor", nil, nil),
	)

	roles, err := repo.GetUserRoles(context.Background(), 10)
	require.NoError(t, err)
	assert.Len(t, roles, 1)
}

func TestRoleRepository_GetUserRoles_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetUserRoles(context.Background(), 1)
	assert.Error(t, err)
}

func TestRoleRepository_AssignUserRoles_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM user_roles").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO user_roles").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.AssignUserRoles(context.Background(), 10, []uint{1})
	require.NoError(t, err)
}

func TestRoleRepository_AssignUserRoles_DeleteError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM user_roles").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.AssignUserRoles(context.Background(), 10, []uint{1})
	assert.Error(t, err)
}

func TestRoleRepository_AssignUserRoles_InsertError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM user_roles").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO user_roles").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.AssignUserRoles(context.Background(), 10, []uint{1})
	assert.Error(t, err)
}

func TestRoleRepository_HasUsers_True(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	has, err := repo.HasUsers(context.Background(), 1)
	require.NoError(t, err)
	assert.True(t, has)
}

func TestRoleRepository_HasUsers_False(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	has, err := repo.HasUsers(context.Background(), 1)
	require.NoError(t, err)
	assert.False(t, has)
}

func TestRoleRepository_HasUsers_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewRoleRepository(db)

	mock.ExpectQuery("SELECT COUNT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.HasUsers(context.Background(), 1)
	assert.Error(t, err)
}

// ==================== PermissionRepository ====================

func TestPermissionRepository_List_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewPermissionRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(permissionColumns).
			AddRow(1, "Read Articles", "article:read", 0, nil, nil).
			AddRow(2, "Write Articles", "article:write", 0, nil, nil),
	)

	perms, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, perms, 2)
}

func TestPermissionRepository_List_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewPermissionRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.List(context.Background())
	assert.Error(t, err)
}

func TestPermissionRepository_GetByID_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewPermissionRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(permissionColumns).AddRow(1, "Read", "article:read", 0, nil, nil),
	)

	p, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, p)
}

func TestPermissionRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewPermissionRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(permissionColumns))

	p, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, p)
}

func TestPermissionRepository_GetByID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewPermissionRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestPermissionRepository_GetUserPermissions_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewPermissionRepository(db)

	mock.ExpectQuery("SELECT DISTINCT").WillReturnRows(
		sqlmock.NewRows(permissionColumns).AddRow(1, "Read", "article:read", 0, nil, nil),
	)

	perms, err := repo.GetUserPermissions(context.Background(), 10)
	require.NoError(t, err)
	assert.Len(t, perms, 1)
}

func TestPermissionRepository_GetUserPermissions_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewPermissionRepository(db)

	mock.ExpectQuery("SELECT DISTINCT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetUserPermissions(context.Background(), 1)
	assert.Error(t, err)
}
