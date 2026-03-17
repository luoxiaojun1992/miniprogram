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

var notificationColumns = []string{"id", "user_id", "title", "content", "is_read", "created_at"}

// ==================== NotificationRepository ====================

func TestNotificationRepository_GetByID_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewNotificationRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(notificationColumns).AddRow(1, 10, "Title", "Body", 0, now),
	)

	n, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, n)
}

func TestNotificationRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewNotificationRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(notificationColumns))

	n, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, n)
}

func TestNotificationRepository_GetByID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewNotificationRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestNotificationRepository_List_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewNotificationRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(notificationColumns).AddRow(1, 10, "Title", "Body", 0, now),
	)

	notifs, total, err := repo.List(context.Background(), 10, 1, 10, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, notifs, 1)
}

func TestNotificationRepository_List_WithIsRead(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewNotificationRepository(db)

	isRead := true
	now := time.Now()
	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(notificationColumns).AddRow(1, 10, "Title", "Body", 1, now),
	)

	_, _, err := repo.List(context.Background(), 10, 1, 10, &isRead)
	require.NoError(t, err)
}

func TestNotificationRepository_List_IsReadFalse(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewNotificationRepository(db)

	isRead := false
	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(notificationColumns))

	_, _, err := repo.List(context.Background(), 10, 1, 10, &isRead)
	require.NoError(t, err)
}

func TestNotificationRepository_List_CountError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewNotificationRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnError(fmt.Errorf("count error"))

	_, _, err := repo.List(context.Background(), 1, 1, 10, nil)
	assert.Error(t, err)
}

func TestNotificationRepository_List_FindError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewNotificationRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("find error"))

	_, _, err := repo.List(context.Background(), 1, 1, 10, nil)
	assert.Error(t, err)
}

func TestNotificationRepository_UnreadCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewNotificationRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := repo.UnreadCount(context.Background(), 10)
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestNotificationRepository_UnreadCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewNotificationRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnError(fmt.Errorf("count error"))

	_, err := repo.UnreadCount(context.Background(), 1)
	assert.Error(t, err)
}

func TestNotificationRepository_MarkRead_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewNotificationRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.MarkRead(context.Background(), 1)
	require.NoError(t, err)
}

func TestNotificationRepository_MarkRead_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewNotificationRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("update error"))
	mock.ExpectRollback()

	err := repo.MarkRead(context.Background(), 1)
	assert.Error(t, err)
}

func TestNotificationRepository_MarkAllRead_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewNotificationRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(5, 5))
	mock.ExpectCommit()

	err := repo.MarkAllRead(context.Background(), 10)
	require.NoError(t, err)
}

func TestNotificationRepository_MarkAllRead_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewNotificationRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("update error"))
	mock.ExpectRollback()

	err := repo.MarkAllRead(context.Background(), 10)
	assert.Error(t, err)
}

func TestNotificationRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewNotificationRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), &entity.Notification{
		Type:    2,
		Title:   "title",
		Content: "content",
	})
	require.NoError(t, err)
}
