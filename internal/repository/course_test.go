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

var courseColumns = []string{"id", "title", "summary", "cover_image", "author_id", "module_id", "status", "price", "view_count", "like_count", "collect_count", "sort_order", "is_free", "created_at", "updated_at"}
var courseUnitColumns = []string{"id", "course_id", "title", "sort_order", "created_at", "updated_at"}
var courseAuthorColumns = []string{"id", "open_id", "nickname", "user_type", "status", "created_at", "updated_at", "deleted_at"}

// ==================== CourseRepository ====================

func TestCourseRepository_GetByID_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)

	now := time.Now()
	// Preload("Author") + Preload("Units") → 3 queries
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(courseColumns).
			AddRow(1, "Course1", "Sum", "", 10, 2, 1, 0, 0, 0, 0, 0, true, now, now),
	)
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(courseAuthorColumns).
		AddRow(10, "oid", "Author", 2, 1, now, now, nil))
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(courseUnitColumns))

	c, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestCourseRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(courseColumns))

	c, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, c)
}

func TestCourseRepository_GetByID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestCourseRepository_List_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(courseColumns).
			AddRow(1, "Course1", "Sum", "", 10, 2, 1, 0, 0, 0, 0, 0, true, now, now),
	)
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(courseAuthorColumns))

	courses, total, err := repo.List(context.Background(), 1, 10, "", nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, courses, 1)
}

func TestCourseRepository_List_WithFilters(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)

	moduleID := uint(2)
	status := int8(1)
	isFree := true
	now := time.Now()

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(courseColumns).
			AddRow(1, "Free Course", "Sum", "", 10, 2, 1, 0, 0, 0, 0, 0, true, now, now),
	)
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(courseAuthorColumns))

	courses, total, err := repo.List(context.Background(), 1, 10, "Free", &moduleID, &status, &isFree)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	_ = courses
}

func TestCourseRepository_List_WithoutModuleID_OnlyNullOrZeroModule(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `courses` WHERE \\(module_id IS NULL OR module_id = 0\\)").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT \\* FROM `courses` WHERE \\(module_id IS NULL OR module_id = 0\\) ORDER BY sort_order DESC, created_at DESC LIMIT \\?").
		WithArgs(10).
		WillReturnRows(
			sqlmock.NewRows(courseColumns).
				AddRow(1, "Course1", "Sum", "", 10, nil, 1, 0, 0, 0, 0, 0, true, now, now),
		)
	mock.ExpectQuery("SELECT \\* FROM `users` WHERE `users`.`id` = \\? AND `users`.`deleted_at` IS NULL").
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows(courseAuthorColumns).
			AddRow(10, "oid", "Author", 2, 1, now, now, nil))

	courses, total, err := repo.List(context.Background(), 1, 10, "", nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, courses, 1)
}

func TestCourseRepository_List_CountError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnError(fmt.Errorf("count error"))

	_, _, err := repo.List(context.Background(), 1, 10, "", nil, nil, nil)
	assert.Error(t, err)
}

func TestCourseRepository_List_FindError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("find error"))

	_, _, err := repo.List(context.Background(), 1, 10, "", nil, nil, nil)
	assert.Error(t, err)
}

func TestCourseRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), &entity.Course{Title: "New Course", AuthorID: 1})
	require.NoError(t, err)
}

func TestCourseRepository_Create_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), &entity.Course{Title: "Fail"})
	assert.Error(t, err)
}

func TestCourseRepository_Update_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(context.Background(), &entity.Course{ID: 1, Title: "Updated"})
	require.NoError(t, err)
}

func TestCourseRepository_Update_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("update error"))
	mock.ExpectRollback()

	err := repo.Update(context.Background(), &entity.Course{ID: 1})
	assert.Error(t, err)
}

func TestCourseRepository_Delete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestCourseRepository_Delete_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestCourseRepository_IncrViewCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)

	mock.ExpectExec("INSERT INTO course_attributes").WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.IncrViewCount(context.Background(), 1)
	require.NoError(t, err)
}

func TestCourseRepository_IncrViewCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)

	mock.ExpectExec("INSERT INTO course_attributes").WillReturnError(fmt.Errorf("exec error"))

	err := repo.IncrViewCount(context.Background(), 1)
	assert.Error(t, err)
}

// ==================== CourseUnitRepository ====================

func TestCourseUnitRepository_GetByID_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseUnitRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(courseUnitColumns).AddRow(1, 10, "Unit1", 1, now, now),
	)

	u, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, u)
}

func TestCourseUnitRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseUnitRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(courseUnitColumns))

	u, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, u)
}

func TestCourseUnitRepository_GetByID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseUnitRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestCourseUnitRepository_ListByCourseID_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseUnitRepository(db)

	now := time.Now()
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(courseUnitColumns).
			AddRow(1, 10, "Unit1", 1, now, now).
			AddRow(2, 10, "Unit2", 2, now, now),
	)

	units, err := repo.ListByCourseID(context.Background(), 10)
	require.NoError(t, err)
	assert.Len(t, units, 2)
}

func TestCourseUnitRepository_ListByCourseID_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseUnitRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.ListByCourseID(context.Background(), 1)
	assert.Error(t, err)
}

func TestCourseUnitRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseUnitRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), &entity.CourseUnit{CourseID: 1, Title: "Unit1"})
	require.NoError(t, err)
}

func TestCourseUnitRepository_Create_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseUnitRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), &entity.CourseUnit{Title: "Fail"})
	assert.Error(t, err)
}

func TestCourseUnitRepository_Update_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseUnitRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(context.Background(), &entity.CourseUnit{ID: 1, Title: "Updated"})
	require.NoError(t, err)
}

func TestCourseUnitRepository_Update_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseUnitRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("update error"))
	mock.ExpectRollback()

	err := repo.Update(context.Background(), &entity.CourseUnit{ID: 1})
	assert.Error(t, err)
}

func TestCourseUnitRepository_Delete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseUnitRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestCourseUnitRepository_Delete_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseUnitRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestCourseRepository_IncrLikeCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectExec("INSERT INTO course_attributes").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, repo.IncrLikeCount(context.Background(), 1))
}

func TestCourseRepository_IncrLikeCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectExec("INSERT INTO course_attributes").WillReturnError(fmt.Errorf("exec error"))
	assert.Error(t, repo.IncrLikeCount(context.Background(), 1))
}

func TestCourseRepository_DecrLikeCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectExec("INSERT INTO course_attributes").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, repo.DecrLikeCount(context.Background(), 1))
}

func TestCourseRepository_DecrLikeCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectExec("INSERT INTO course_attributes").WillReturnError(fmt.Errorf("exec error"))
	assert.Error(t, repo.DecrLikeCount(context.Background(), 1))
}

func TestCourseRepository_IncrCollectCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectExec("INSERT INTO course_attributes").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, repo.IncrCollectCount(context.Background(), 1))
}

func TestCourseRepository_IncrCollectCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectExec("INSERT INTO course_attributes").WillReturnError(fmt.Errorf("exec error"))
	assert.Error(t, repo.IncrCollectCount(context.Background(), 1))
}

func TestCourseRepository_DecrCollectCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectExec("INSERT INTO course_attributes").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, repo.DecrCollectCount(context.Background(), 1))
}

func TestCourseRepository_DecrCollectCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectExec("INSERT INTO course_attributes").WillReturnError(fmt.Errorf("exec error"))
	assert.Error(t, repo.DecrCollectCount(context.Background(), 1))
}

func TestCourseRepository_IncrCommentCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectExec("INSERT INTO course_attributes").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, repo.IncrCommentCount(context.Background(), 1))
}

func TestCourseRepository_IncrCommentCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectExec("INSERT INTO course_attributes").WillReturnError(fmt.Errorf("exec error"))
	assert.Error(t, repo.IncrCommentCount(context.Background(), 1))
}

func TestCourseRepository_DecrCommentCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectExec("INSERT INTO course_attributes").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, repo.DecrCommentCount(context.Background(), 1))
}

func TestCourseRepository_DecrCommentCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectExec("INSERT INTO course_attributes").WillReturnError(fmt.Errorf("exec error"))
	assert.Error(t, repo.DecrCommentCount(context.Background(), 1))
}

func TestCourseRepository_IncrShareCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectExec("INSERT INTO course_attributes").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, repo.IncrShareCount(context.Background(), 1))
}

func TestCourseRepository_IncrShareCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectExec("INSERT INTO course_attributes").WillReturnError(fmt.Errorf("exec error"))
	assert.Error(t, repo.IncrShareCount(context.Background(), 1))
}

func TestCourseRepository_IncrStudyCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectExec("INSERT INTO course_attributes").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, repo.IncrStudyCount(context.Background(), 1))
}

func TestCourseRepository_IncrStudyCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectExec("INSERT INTO course_attributes").WillReturnError(fmt.Errorf("exec error"))
	assert.Error(t, repo.IncrStudyCount(context.Background(), 1))
}

func TestCourseRepository_HasAssociations_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(3))
	ok, err := repo.HasAssociations(context.Background(), 1)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCourseRepository_HasAssociations_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseRepository(db)
	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("query error"))
	_, err := repo.HasAssociations(context.Background(), 1)
	assert.Error(t, err)
}

func TestCourseUnitRepository_HasStudyRecords_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseUnitRepository(db)
	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	ok, err := repo.HasStudyRecords(context.Background(), 1)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCourseUnitRepository_HasStudyRecords_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewCourseUnitRepository(db)
	mock.ExpectQuery("SELECT count").WillReturnError(fmt.Errorf("query error"))
	_, err := repo.HasStudyRecords(context.Background(), 1)
	assert.Error(t, err)
}
