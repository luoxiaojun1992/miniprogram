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

var wechatConfigColumns = []string{"id", "app_id", "app_secret", "created_at", "updated_at"}
var auditLogColumns = []string{"id", "user_id", "module", "action", "detail", "created_at"}
var logConfigColumns = []string{"id", "log_level", "log_path", "created_at", "updated_at"}

// ==================== WechatConfigRepository ====================

func TestWechatConfigRepository_Get_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewWechatConfigRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(wechatConfigColumns).AddRow(1, "wx_app_id", "wx_secret", now, now),
	)

	cfg, err := repo.Get(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestWechatConfigRepository_Get_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewWechatConfigRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(wechatConfigColumns))

	cfg, err := repo.Get(context.Background())
	require.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestWechatConfigRepository_Get_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewWechatConfigRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.Get(context.Background())
	assert.Error(t, err)
}

func TestWechatConfigRepository_Update_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewWechatConfigRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	now := time.Now()
	err := repo.Update(context.Background(), &entity.WechatConfig{ID: 1, AppID: "wx_app_id", AppSecret: "secret", UpdatedAt: now})
	require.NoError(t, err)
}

func TestWechatConfigRepository_Update_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewWechatConfigRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("update error"))
	mock.ExpectRollback()

	err := repo.Update(context.Background(), &entity.WechatConfig{ID: 1})
	assert.Error(t, err)
}

// ==================== AuditLogRepository ====================

func TestAuditLogRepository_List_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAuditLogRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(auditLogColumns).AddRow(1, 10, "article", "create", "{}", now),
	)

	logs, total, err := repo.List(context.Background(), 1, 10, "", "", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, logs, 1)
}

func TestAuditLogRepository_List_WithFilters(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAuditLogRepository(db)

	now := time.Now()
	start := "2024-01-01"
	end := "2024-12-31"

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(auditLogColumns).AddRow(1, 10, "article", "create", "{}", now),
	)

	_, _, err := repo.List(context.Background(), 1, 10, "article", "create", &start, &end)
	require.NoError(t, err)
}

func TestAuditLogRepository_List_CountError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAuditLogRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnError(fmt.Errorf("count error"))

	_, _, err := repo.List(context.Background(), 1, 10, "", "", nil, nil)
	assert.Error(t, err)
}

func TestAuditLogRepository_List_FindError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAuditLogRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("find error"))

	_, _, err := repo.List(context.Background(), 1, 10, "", "", nil, nil)
	assert.Error(t, err)
}

func TestAuditLogRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAuditLogRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), &entity.AuditLog{Module: "article", Action: "create"})
	require.NoError(t, err)
}

func TestAuditLogRepository_Create_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAuditLogRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), &entity.AuditLog{})
	assert.Error(t, err)
}

// ==================== LogConfigRepository ====================

func TestLogConfigRepository_Get_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewLogConfigRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(logConfigColumns).AddRow(1, "info", "/var/log/app.log", now, now),
	)

	cfg, err := repo.Get(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestLogConfigRepository_Get_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewLogConfigRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(logConfigColumns))

	cfg, err := repo.Get(context.Background())
	require.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestLogConfigRepository_Get_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewLogConfigRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.Get(context.Background())
	assert.Error(t, err)
}

func TestLogConfigRepository_Update_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewLogConfigRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	now := time.Now()
	err := repo.Update(context.Background(), &entity.LogConfig{ID: 1, UpdatedAt: now})
	require.NoError(t, err)
}

func TestLogConfigRepository_Update_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewLogConfigRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("update error"))
	mock.ExpectRollback()

	err := repo.Update(context.Background(), &entity.LogConfig{ID: 1})
	assert.Error(t, err)
}
