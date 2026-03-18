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

var fileColumns = []string{"id", "key", "filename", "usage", "category", "business", "static_url", "created_by", "created_at"}

func TestNewFileRepository(t *testing.T) {
	db, _ := newTestDB(t)
	repo := NewFileRepository(db)
	require.NotNil(t, repo)
}

func TestFileRepository_GetByID_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewFileRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(fileColumns).AddRow(1, "avatar/a.png", "a.png", "avatar", "image", "", "", 1, now),
	)

	f, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, f)
	assert.Equal(t, uint64(1), f.ID)
}

func TestFileRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewFileRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(fileColumns))

	f, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, f)
}

func TestFileRepository_GetByID_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewFileRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestFileRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewFileRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), &entity.File{
		Key:      "avatar/a.png",
		Filename: "a.png",
		Usage:    "avatar",
		Category: "image",
	})
	require.NoError(t, err)
}

func TestFileRepository_Create_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewFileRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), &entity.File{
		Key:      "avatar/a.png",
		Filename: "a.png",
		Usage:    "avatar",
		Category: "image",
	})
	assert.Error(t, err)
}
