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

var bannerColumns = []string{"id", "title", "image_file_id", "link_url", "sort_order", "status", "created_at", "updated_at"}

func TestNewBannerRepository(t *testing.T) {
	db, _ := newTestDB(t)
	repo := NewBannerRepository(db)
	require.NotNil(t, repo)
}

func TestBannerRepository_GetByID_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewBannerRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(bannerColumns).AddRow(1, "banner", nil, "https://example.com", 10, 1, now, now),
	)

	b, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, b)
	assert.Equal(t, uint64(1), b.ID)
}

func TestBannerRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewBannerRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(bannerColumns))

	b, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, b)
}

func TestBannerRepository_GetByID_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewBannerRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestBannerRepository_List_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewBannerRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(bannerColumns).AddRow(1, "banner", nil, "https://example.com", 10, 1, now, now),
	)

	list, err := repo.List(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestBannerRepository_List_WithStatus_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewBannerRepository(db)

	status := int8(1)
	mock.ExpectQuery("SELECT").WithArgs(status).WillReturnRows(sqlmock.NewRows(bannerColumns))

	_, err := repo.List(context.Background(), &status)
	require.NoError(t, err)
}

func TestBannerRepository_List_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewBannerRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("find error"))

	_, err := repo.List(context.Background(), nil)
	assert.Error(t, err)
}

func TestBannerRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewBannerRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), &entity.Banner{Title: "new"})
	require.NoError(t, err)
}

func TestBannerRepository_Create_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewBannerRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), &entity.Banner{Title: "new"})
	assert.Error(t, err)
}

func TestBannerRepository_Update_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewBannerRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(context.Background(), &entity.Banner{ID: 1, Title: "updated"})
	require.NoError(t, err)
}

func TestBannerRepository_Update_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewBannerRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("update error"))
	mock.ExpectRollback()

	err := repo.Update(context.Background(), &entity.Banner{ID: 1, Title: "updated"})
	assert.Error(t, err)
}

func TestBannerRepository_Delete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewBannerRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestBannerRepository_Delete_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewBannerRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

