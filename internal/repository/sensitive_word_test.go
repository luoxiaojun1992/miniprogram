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

var sensitiveWordColumns = []string{"id", "word", "status", "created_at", "updated_at"}

func TestNewSensitiveWordRepository(t *testing.T) {
	db, _ := newTestDB(t)
	repo := NewSensitiveWordRepository(db)
	require.NotNil(t, repo)
}

func TestSensitiveWordRepository_ListEnabledWords_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewSensitiveWordRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(sensitiveWordColumns).
			AddRow(1, "foo", 1, now, now).
			AddRow(2, "", 1, now, now),
	)

	words, err := repo.ListEnabledWords(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"foo"}, words)
}

func TestSensitiveWordRepository_ListEnabledWords_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewSensitiveWordRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("query error"))

	_, err := repo.ListEnabledWords(context.Background())
	assert.Error(t, err)
}

