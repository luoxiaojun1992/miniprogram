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

func TestAttributeRepository_CRUDAndAssoc(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAttributeRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "性别"))
	attr, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, uint(1), attr.ID)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "性别"))
	attrs, err := repo.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, attrs, 1)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	require.NoError(t, repo.Create(context.Background(), &entity.Attribute{Name: "年级"}))

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	require.NoError(t, repo.Update(context.Background(), &entity.Attribute{ID: 1, Name: "职业"}))

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	require.NoError(t, repo.Delete(context.Background(), 1))

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	has, err := repo.HasUserAssociations(context.Background(), 1)
	require.NoError(t, err)
	assert.True(t, has)
}

func TestUserAttributeRepository_CRUDAndErrors(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserAttributeRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"id", "user_id", "attribute_id", "value"}).AddRow(1, 2, 3, "男"),
	)
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(3, "性别"))
	uas, err := repo.ListByUserID(context.Background(), 2)
	require.NoError(t, err)
	assert.Len(t, uas, 1)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	require.NoError(t, repo.Upsert(context.Background(), &entity.UserAttribute{UserID: 2, AttributeID: 3, Value: "女"}))

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	require.NoError(t, repo.Delete(context.Background(), 2, 3))

	db2, mock2 := newTestDB(t)
	repo2 := NewAttributeRepository(db2)
	mock2.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))
	_, err = repo2.GetByID(context.Background(), 1)
	assert.Error(t, err)
}
