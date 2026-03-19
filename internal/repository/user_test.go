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

// ==================== UserRepository ====================

func TestUserRepository_GetByID_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "open_id", "nickname", "user_type", "status", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, "openid1", "Nick", 1, 1, now, now, nil)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	user, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, uint64(1), user.ID)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	user, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_GetByID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db connection error"))

	user, err := repo.GetByID(context.Background(), 1)
	assert.Nil(t, user)
	assert.Error(t, err)
}

func TestUserRepository_GetByOpenID_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "open_id", "nickname", "user_type", "status", "created_at", "updated_at", "deleted_at"}).
		AddRow(5, "oid_abc", "User5", 1, 1, now, now, nil)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	user, err := repo.GetByOpenID(context.Background(), "oid_abc")
	require.NoError(t, err)
	assert.Equal(t, uint64(5), user.ID)
}

func TestUserRepository_GetByOpenID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	user, err := repo.GetByOpenID(context.Background(), "no_such_id")
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_GetByOpenID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByOpenID(context.Background(), "oid")
	assert.Error(t, err)
}

func TestUserRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	u := &entity.User{Nickname: "new user", UserType: 1}
	err := repo.Create(context.Background(), u)
	require.NoError(t, err)
}

func TestUserRepository_Create_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("duplicate entry"))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), &entity.User{Nickname: "x"})
	assert.Error(t, err)
}

func TestUserRepository_Update_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	now := time.Now()
	u := &entity.User{ID: 1, Nickname: "updated", UpdatedAt: now}
	err := repo.Update(context.Background(), u)
	require.NoError(t, err)
}

func TestUserRepository_Update_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("update error"))
	mock.ExpectRollback()

	err := repo.Update(context.Background(), &entity.User{ID: 1})
	assert.Error(t, err)
}

func TestUserRepository_Delete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1)) // soft delete
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestUserRepository_Delete_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestUserRepository_List_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"id", "open_id", "nickname", "user_type", "status", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "o1", "N1", 1, 1, now, now, nil).
			AddRow(2, "o2", "N2", 1, 1, now, now, nil),
	)

	users, total, err := repo.List(context.Background(), 1, 10, "", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, users, 2)
}

func TestUserRepository_List_WithFilters(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	userType := int8(1)
	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"id", "open_id", "nickname", "user_type", "status", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "o1", "Keyword", 1, 1, now, now, nil),
	)

	users, total, err := repo.List(context.Background(), 1, 10, "Keyword", &userType)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, users, 1)
}

func TestUserRepository_List_CountError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnError(fmt.Errorf("count error"))

	_, _, err := repo.List(context.Background(), 1, 10, "", nil)
	assert.Error(t, err)
}

func TestUserRepository_List_FindError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("find error"))

	_, _, err := repo.List(context.Background(), 1, 10, "", nil)
	assert.Error(t, err)
}

func TestUserRepository_GetWithTags_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	now := time.Now()
	// GORM Preload("Tags") issues a second SELECT
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"id", "open_id", "nickname", "user_type", "status", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "o1", "N1", 1, 1, now, now, nil),
	)
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "tag_name", "created_at"}))

	user, err := repo.GetWithTags(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, user)
}

func TestUserRepository_GetWithTags_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	user, err := repo.GetWithTags(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_GetWithTags_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetWithTags(context.Background(), 1)
	assert.Error(t, err)
}

// ==================== AdminUserRepository ====================

func TestAdminUserRepository_GetByEmail_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAdminUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "user_id", "email", "password_hash"}).
		AddRow(1, 10, "admin@test.com", "hash")
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	admin, err := repo.GetByEmail(context.Background(), "admin@test.com")
	require.NoError(t, err)
	assert.NotNil(t, admin)
}

func TestAdminUserRepository_GetByEmail_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAdminUserRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	admin, err := repo.GetByEmail(context.Background(), "no@test.com")
	require.NoError(t, err)
	assert.Nil(t, admin)
}

func TestAdminUserRepository_GetByEmail_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAdminUserRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByEmail(context.Background(), "a@b.com")
	assert.Error(t, err)
}

func TestAdminUserRepository_GetByUserID_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAdminUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "user_id", "email"}).AddRow(1, 10, "a@b.com")
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	admin, err := repo.GetByUserID(context.Background(), 10)
	require.NoError(t, err)
	assert.NotNil(t, admin)
}

func TestAdminUserRepository_GetByUserID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAdminUserRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	admin, err := repo.GetByUserID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, admin)
}

func TestAdminUserRepository_GetByUserID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAdminUserRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByUserID(context.Background(), 1)
	assert.Error(t, err)
}

func TestAdminUserRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAdminUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), &entity.AdminUser{UserID: 10, Email: "a@b.com", PasswordHash: "hash"})
	require.NoError(t, err)
}

func TestAdminUserRepository_Create_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAdminUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("duplicate"))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), &entity.AdminUser{Email: "dup@b.com"})
	assert.Error(t, err)
}

func TestAdminUserRepository_UpdateLastLogin_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAdminUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateLastLogin(context.Background(), 1)
	require.NoError(t, err)
}

func TestAdminUserRepository_UpdateLastLogin_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewAdminUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("update error"))
	mock.ExpectRollback()

	err := repo.UpdateLastLogin(context.Background(), 1)
	assert.Error(t, err)
}

// ==================== UserTagRepository ====================

func TestUserTagRepository_GetByUserID_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserTagRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"id", "user_id", "tag_name", "created_at"}).
			AddRow(1, 10, "vip", now),
	)

	tags, err := repo.GetByUserID(context.Background(), 10)
	require.NoError(t, err)
	assert.Len(t, tags, 1)
}

func TestUserTagRepository_GetByUserID_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserTagRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByUserID(context.Background(), 1)
	assert.Error(t, err)
}

func TestUserTagRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserTagRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), &entity.UserTag{UserID: 1, TagName: "vip"})
	require.NoError(t, err)
}

func TestUserTagRepository_Create_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserTagRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), &entity.UserTag{TagName: "x"})
	assert.Error(t, err)
}

func TestUserTagRepository_Delete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserTagRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestUserTagRepository_Delete_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserTagRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestUserRepository_HasAssociations_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(4))
	ok, err := repo.HasAssociations(context.Background(), 1)
	require.NoError(t, err)
	assert.True(t, ok)
}
