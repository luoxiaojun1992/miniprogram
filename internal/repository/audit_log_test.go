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

var auditLogGetByIDColumns = []string{"id", "user_id", "username", "action", "module", "description", "ip_address", "user_agent", "request_data", "created_at"}

func TestAuditLogRepository_GetByID_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAuditLogRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(auditLogGetByIDColumns).AddRow(1, 1, "admin", "create", "article", "desc", "127.0.0.1", "ua", "{}", now),
	)

	log, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, log)
	assert.Equal(t, uint64(1), log.ID)
}

func TestAuditLogRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAuditLogRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(auditLogGetByIDColumns))

	log, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, log)
}

func TestAuditLogRepository_GetByID_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAuditLogRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}
