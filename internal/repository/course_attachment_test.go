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

var courseAttachmentColumns = []string{"id", "course_id", "file_id", "sort_order", "created_at"}

func TestNewCourseAttachmentRepository(t *testing.T) {
	db, _ := newTestDB(t)
	repo := NewCourseAttachmentRepository(db)
	require.NotNil(t, repo)
}

func TestCourseAttachmentRepository_ListFileIDs_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseAttachmentRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(courseAttachmentColumns).
			AddRow(1, 10, 101, 1, now).
			AddRow(2, 10, 102, 2, now),
	)

	ids, err := repo.ListFileIDs(context.Background(), 10)
	require.NoError(t, err)
	assert.Equal(t, []uint64{101, 102}, ids)
}

func TestCourseAttachmentRepository_ListFileIDs_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseAttachmentRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("query error"))

	_, err := repo.ListFileIDs(context.Background(), 10)
	assert.Error(t, err)
}

func TestCourseAttachmentRepository_Replace_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseAttachmentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 2))
	mock.ExpectCommit()

	err := repo.Replace(context.Background(), 10, []uint64{101, 0, 101, 102})
	require.NoError(t, err)
}

func TestCourseAttachmentRepository_Replace_DeleteError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseAttachmentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.Replace(context.Background(), 10, []uint64{101})
	assert.Error(t, err)
}

func TestCourseAttachmentRepository_Replace_CreateError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseAttachmentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.Replace(context.Background(), 10, []uint64{101, 102})
	assert.Error(t, err)
}

func TestCourseAttachmentRepository_Replace_OnlyZeroIDs(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseAttachmentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Replace(context.Background(), 10, []uint64{0, 0})
	require.NoError(t, err)
}
