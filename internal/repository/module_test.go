package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
)

var moduleColumns = []string{"id", "name", "icon", "status", "sort_order", "created_at", "updated_at"}
var modulePageColumns = []string{"id", "module_id", "page_name", "page_path", "sort_order", "created_at", "updated_at"}

// ==================== ModuleRepository ====================

func TestModuleRepository_GetByID_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModuleRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(moduleColumns).AddRow(1, "Module1", "icon.png", 1, 0, now, now),
	)

	m, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, m)
}

func TestModuleRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModuleRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(moduleColumns))

	m, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, m)
}

func TestModuleRepository_GetByID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModuleRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestModuleRepository_List_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModuleRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(moduleColumns).
			AddRow(1, "Module1", "icon.png", 1, 0, now, now).
			AddRow(2, "Module2", "icon2.png", 1, 1, now, now),
	)

	modules, err := repo.List(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, modules, 2)
}

func TestModuleRepository_List_WithStatus(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModuleRepository(db)

	status := int8(1)
	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(moduleColumns).AddRow(1, "Module1", "icon.png", 1, 0, now, now),
	)

	modules, err := repo.List(context.Background(), &status)
	require.NoError(t, err)
	assert.Len(t, modules, 1)
}

func TestModuleRepository_List_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModuleRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.List(context.Background(), nil)
	assert.Error(t, err)
}

func TestModuleRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModuleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), &entity.Module{Title: "New Module"})
	require.NoError(t, err)
}

func TestModuleRepository_Create_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModuleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), &entity.Module{Title: "Fail"})
	assert.Error(t, err)
}

func TestModuleRepository_Update_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModuleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(context.Background(), &entity.Module{ID: 1, Title: "Updated"})
	require.NoError(t, err)
}

func TestModuleRepository_Update_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModuleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("update error"))
	mock.ExpectRollback()

	err := repo.Update(context.Background(), &entity.Module{ID: 1})
	assert.Error(t, err)
}

func TestModuleRepository_Delete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModuleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestModuleRepository_Delete_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModuleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

// ==================== ModulePageRepository ====================

func TestModulePageRepository_GetByID_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModulePageRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(modulePageColumns).AddRow(1, 10, "Page1", "/path1", 0, now, now),
	)

	p, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, p)
}

func TestModulePageRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModulePageRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(modulePageColumns))

	p, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, p)
}

func TestModulePageRepository_GetByID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModulePageRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestModulePageRepository_ListByModuleID_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModulePageRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(modulePageColumns).
			AddRow(1, 10, "Page1", "/path1", 0, now, now).
			AddRow(2, 10, "Page2", "/path2", 1, now, now),
	)

	pages, err := repo.ListByModuleID(context.Background(), 10)
	require.NoError(t, err)
	assert.Len(t, pages, 2)
}

func TestModulePageRepository_ListByModuleID_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModulePageRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.ListByModuleID(context.Background(), 1)
	assert.Error(t, err)
}

func TestModulePageRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModulePageRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), &entity.ModulePage{ModuleID: 1, Title: "Page1"})
	require.NoError(t, err)
}

func TestModulePageRepository_Create_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModulePageRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), &entity.ModulePage{Title: "Fail"})
	assert.Error(t, err)
}

func TestModulePageRepository_Update_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModulePageRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(context.Background(), &entity.ModulePage{ID: 1, Title: "Updated"})
	require.NoError(t, err)
}

func TestModulePageRepository_Update_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModulePageRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("update error"))
	mock.ExpectRollback()

	err := repo.Update(context.Background(), &entity.ModulePage{ID: 1})
	assert.Error(t, err)
}

func TestModulePageRepository_Delete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModulePageRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestModulePageRepository_Delete_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewModulePageRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}
