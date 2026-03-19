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

var collectionColumns = []string{"id", "user_id", "content_type", "content_id", "created_at"}
var likeColumns = []string{"id", "user_id", "content_type", "content_id", "created_at"}
var followColumns = []string{"id", "follower_id", "followed_id", "created_at"}
var commentColumns = []string{"id", "user_id", "content_type", "content_id", "parent_id", "content", "status", "created_at"}
var studyRecordColumns = []string{"id", "user_id", "course_id", "unit_id", "progress", "status", "created_at", "updated_at"}

// ==================== ContentPermissionRepository ====================

func TestContentPermissionRepository_GetByContent_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewContentPermissionRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"id", "content_type", "content_id", "role_id", "created_at"}).
			AddRow(1, 1, 100, 2, now),
	)

	perms, err := repo.GetByContent(context.Background(), 1, 100)
	require.NoError(t, err)
	assert.Len(t, perms, 1)
}

func TestContentPermissionRepository_GetByContent_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewContentPermissionRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByContent(context.Background(), 1, 100)
	assert.Error(t, err)
}

func TestContentPermissionRepository_SetContentPermissions_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewContentPermissionRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM content_permissions").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	roleID1 := uint(1)
	err := repo.SetContentPermissions(context.Background(), 1, 100, []uint{roleID1})
	require.NoError(t, err)
}

func TestContentPermissionRepository_SetContentPermissions_Empty(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewContentPermissionRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM content_permissions").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := repo.SetContentPermissions(context.Background(), 1, 100, []uint{})
	require.NoError(t, err)
}

func TestContentPermissionRepository_SetContentPermissions_DeleteError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewContentPermissionRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM content_permissions").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.SetContentPermissions(context.Background(), 1, 100, []uint{1})
	assert.Error(t, err)
}

func TestContentPermissionRepository_SetContentPermissions_InsertError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewContentPermissionRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM content_permissions").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.SetContentPermissions(context.Background(), 1, 100, []uint{1})
	assert.Error(t, err)
}

// ==================== StudyRecordRepository ====================

func TestStudyRecordRepository_GetByUserAndUnit_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewStudyRecordRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(studyRecordColumns).AddRow(1, 10, 5, 20, 60, 1, now, now),
	)

	rec, err := repo.GetByUserAndUnit(context.Background(), 10, 20)
	require.NoError(t, err)
	assert.NotNil(t, rec)
}

func TestStudyRecordRepository_GetByUserAndUnit_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewStudyRecordRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(studyRecordColumns))

	rec, err := repo.GetByUserAndUnit(context.Background(), 10, 20)
	require.NoError(t, err)
	assert.Nil(t, rec)
}

func TestStudyRecordRepository_GetByUserAndUnit_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewStudyRecordRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByUserAndUnit(context.Background(), 1, 1)
	assert.Error(t, err)
}

func TestStudyRecordRepository_ListByUser_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewStudyRecordRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(studyRecordColumns).AddRow(1, 10, 5, 20, 60, 1, now, now),
	)

	records, total, err := repo.ListByUser(context.Background(), 10, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, records, 1)
}

func TestStudyRecordRepository_ListByUser_CountError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewStudyRecordRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnError(fmt.Errorf("count error"))

	_, _, err := repo.ListByUser(context.Background(), 1, 1, 10)
	assert.Error(t, err)
}

func TestStudyRecordRepository_ListByUser_FindError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewStudyRecordRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("find error"))

	_, _, err := repo.ListByUser(context.Background(), 1, 1, 10)
	assert.Error(t, err)
}

func TestStudyRecordRepository_Upsert_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewStudyRecordRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	rec := &entity.UserStudyRecord{UserID: 1, CourseID: 5, UnitID: 20, Progress: 60, Status: 1}
	err := repo.Upsert(context.Background(), rec)
	require.NoError(t, err)
}

func TestStudyRecordRepository_Upsert_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewStudyRecordRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("upsert error"))
	mock.ExpectRollback()

	rec := &entity.UserStudyRecord{UserID: 1, UnitID: 1}
	err := repo.Upsert(context.Background(), rec)
	assert.Error(t, err)
}

// ==================== CollectionRepository ====================

func TestCollectionRepository_Get_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCollectionRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(collectionColumns).AddRow(1, 10, 1, 100, now),
	)

	c, err := repo.Get(context.Background(), 10, 1, 100)
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestCollectionRepository_Get_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCollectionRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(collectionColumns))

	c, err := repo.Get(context.Background(), 10, 1, 100)
	require.NoError(t, err)
	assert.Nil(t, c)
}

func TestCollectionRepository_Get_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCollectionRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.Get(context.Background(), 1, 1, 1)
	assert.Error(t, err)
}

func TestCollectionRepository_List_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCollectionRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(collectionColumns).AddRow(1, 10, 1, 100, now),
	)

	cols, total, err := repo.List(context.Background(), 10, 1, 10, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, cols, 1)
}

func TestCollectionRepository_List_WithContentType(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCollectionRepository(db)

	ct := int8(1)
	now := time.Now()
	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(collectionColumns).AddRow(1, 10, 1, 100, now),
	)

	_, _, err := repo.List(context.Background(), 10, 1, 10, &ct)
	require.NoError(t, err)
}

func TestCollectionRepository_List_CountError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCollectionRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnError(fmt.Errorf("count error"))

	_, _, err := repo.List(context.Background(), 1, 1, 10, nil)
	assert.Error(t, err)
}

func TestCollectionRepository_List_FindError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCollectionRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("find error"))

	_, _, err := repo.List(context.Background(), 1, 1, 10, nil)
	assert.Error(t, err)
}

func TestCollectionRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCollectionRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), &entity.Collection{UserID: 1, ContentType: 1, ContentID: 100})
	require.NoError(t, err)
}

func TestCollectionRepository_Create_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCollectionRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), &entity.Collection{})
	assert.Error(t, err)
}

func TestCollectionRepository_Delete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCollectionRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1, 1, 100)
	require.NoError(t, err)
}

func TestCollectionRepository_Delete_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCollectionRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1, 1, 100)
	assert.Error(t, err)
}

// ==================== LikeRepository ====================

func TestLikeRepository_Get_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewLikeRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(likeColumns).AddRow(1, 10, 1, 100, now),
	)

	l, err := repo.Get(context.Background(), 10, 1, 100)
	require.NoError(t, err)
	assert.NotNil(t, l)
}

func TestLikeRepository_Get_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewLikeRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(likeColumns))

	l, err := repo.Get(context.Background(), 10, 1, 100)
	require.NoError(t, err)
	assert.Nil(t, l)
}

func TestLikeRepository_Get_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewLikeRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.Get(context.Background(), 1, 1, 1)
	assert.Error(t, err)
}

func TestLikeRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewLikeRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), &entity.Like{UserID: 1, ContentType: 1, ContentID: 100})
	require.NoError(t, err)
}

func TestLikeRepository_Create_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewLikeRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), &entity.Like{})
	assert.Error(t, err)
}

func TestLikeRepository_Delete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewLikeRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1, 1, 100)
	require.NoError(t, err)
}

func TestLikeRepository_Delete_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewLikeRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1, 1, 100)
	assert.Error(t, err)
}

// ==================== FollowRepository ====================

func TestFollowRepository_Get_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewFollowRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(followColumns).AddRow(1, 10, 11, now),
	)

	f, err := repo.Get(context.Background(), 10, 11)
	require.NoError(t, err)
	assert.NotNil(t, f)
}

func TestFollowRepository_Get_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewFollowRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(followColumns))

	f, err := repo.Get(context.Background(), 10, 11)
	require.NoError(t, err)
	assert.Nil(t, f)
}

func TestFollowRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewFollowRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), &entity.Follow{FollowerID: 1, FollowedID: 2})
	require.NoError(t, err)
}

func TestFollowRepository_Delete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewFollowRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1, 2)
	require.NoError(t, err)
}

// ==================== CommentRepository ====================

func TestCommentRepository_GetByID_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	now := time.Now()
	// Preload("User") → 2 queries
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(commentColumns).AddRow(1, 10, 1, 100, 0, "Great content", 1, now),
	)
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"id", "open_id", "nickname", "user_type", "status", "created_at", "updated_at", "deleted_at"}).
			AddRow(10, "oid", "User10", 1, 1, now, now, nil),
	)

	c, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestCommentRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(commentColumns))

	c, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, c)
}

func TestCommentRepository_GetByID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestCommentRepository_List_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(commentColumns).AddRow(1, 10, 1, 100, 0, "Nice", 1, now),
	)
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "open_id", "nickname", "user_type", "status", "created_at", "updated_at", "deleted_at"}))

	comments, total, err := repo.List(context.Background(), 1, 100, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, comments, 1)
}

func TestCommentRepository_List_CountError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnError(fmt.Errorf("count error"))

	_, _, err := repo.List(context.Background(), 1, 100, 1, 10)
	assert.Error(t, err)
}

func TestCommentRepository_List_FindError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("find error"))

	_, _, err := repo.List(context.Background(), 1, 100, 1, 10)
	assert.Error(t, err)
}

func TestCommentRepository_ListAdmin_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	now := time.Now()
	status := int8(1)
	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(commentColumns).AddRow(1, 10, 1, 100, 0, "Nice", 1, now),
	)
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "open_id", "nickname", "user_type", "status", "created_at", "updated_at", "deleted_at"}))

	comments, total, err := repo.ListAdmin(context.Background(), 1, 10, &status)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, comments, 1)
}

func TestCommentRepository_ListAdmin_NoFilter(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(commentColumns))

	_, _, err := repo.ListAdmin(context.Background(), 1, 10, nil)
	require.NoError(t, err)
}

func TestCommentRepository_HasReplies_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	ok, err := repo.HasReplies(context.Background(), 1)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCommentRepository_ListAdmin_CountError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnError(fmt.Errorf("count error"))

	_, _, err := repo.ListAdmin(context.Background(), 1, 10, nil)
	assert.Error(t, err)
}

func TestCommentRepository_ListAdmin_FindError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("find error"))

	_, _, err := repo.ListAdmin(context.Background(), 1, 10, nil)
	assert.Error(t, err)
}

func TestCommentRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), &entity.Comment{UserID: 1, ContentType: 1, ContentID: 100, Content: "Nice"})
	require.NoError(t, err)
}

func TestCommentRepository_Create_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), &entity.Comment{})
	assert.Error(t, err)
}

func TestCommentRepository_UpdateStatus_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateStatus(context.Background(), 1, 1)
	require.NoError(t, err)
}

func TestCommentRepository_UpdateStatus_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("update error"))
	mock.ExpectRollback()

	err := repo.UpdateStatus(context.Background(), 1, 1)
	assert.Error(t, err)
}

func TestCommentRepository_Delete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestCommentRepository_Delete_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCommentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}
