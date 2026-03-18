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

var articleColumns = []string{"id", "title", "summary", "content", "content_type", "cover_image", "author_id", "module_id", "status", "view_count", "like_count", "collect_count", "sort_order", "created_at", "updated_at"}
var articleAuthorColumns = []string{"id", "open_id", "nickname", "user_type", "status", "created_at", "updated_at", "deleted_at"}

func nowTime() time.Time { return time.Now() }

func TestArticleRepository_GetByID_Found(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	now := nowTime()
	// GORM Preload("Author") → 2 queries
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(articleColumns).
			AddRow(1, "Title", "Sum", "Content", 1, "", 10, 2, 1, 0, 0, 0, 0, now, now),
	)
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(articleAuthorColumns).
			AddRow(10, "oid", "Author", 2, 1, now, now, nil),
	)

	art, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), art.ID)
}

func TestArticleRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(articleColumns))

	art, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, art)
}

func TestArticleRepository_GetByID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestArticleRepository_List_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	now := nowTime()
	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(articleColumns).
			AddRow(1, "Title", "Sum", "Content", 1, "", 10, 2, 1, 0, 0, 0, 0, now, now),
	)
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(articleAuthorColumns))

	arts, total, err := repo.List(context.Background(), 1, 10, "", nil, nil, "")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, arts, 1)
}

func TestArticleRepository_List_WithFilters(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	moduleID := uint(5)
	status := int8(1)
	now := nowTime()

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows(articleColumns).
			AddRow(1, "Title", "Sum", "Content", 1, "", 10, 5, 1, 0, 0, 0, 0, now, now),
	)
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(articleAuthorColumns))

	arts, total, err := repo.List(context.Background(), 1, 10, "Title", &moduleID, &status, "created_at")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	_ = arts
}

func TestArticleRepository_List_SortViewCount(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(articleColumns))

	_, _, err := repo.List(context.Background(), 1, 10, "", nil, nil, "-view_count")
	require.NoError(t, err)
}

func TestArticleRepository_List_SortLikeCount(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(articleColumns))

	_, _, err := repo.List(context.Background(), 1, 10, "", nil, nil, "-like_count")
	require.NoError(t, err)
}

func TestArticleRepository_List_CountError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnError(fmt.Errorf("count error"))

	_, _, err := repo.List(context.Background(), 1, 10, "", nil, nil, "")
	assert.Error(t, err)
}

func TestArticleRepository_List_FindError(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	mock.ExpectQuery("SELECT count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("find error"))

	_, _, err := repo.List(context.Background(), 1, 10, "", nil, nil, "")
	assert.Error(t, err)
}

func TestArticleRepository_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), &entity.Article{Title: "New", AuthorID: 1})
	require.NoError(t, err)
}

func TestArticleRepository_Create_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	err := repo.Create(context.Background(), &entity.Article{Title: "Fail"})
	assert.Error(t, err)
}

func TestArticleRepository_Update_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(context.Background(), &entity.Article{ID: 1, Title: "Updated"})
	require.NoError(t, err)
}

func TestArticleRepository_Update_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnError(fmt.Errorf("update error"))
	mock.ExpectRollback()

	err := repo.Update(context.Background(), &entity.Article{ID: 1})
	assert.Error(t, err)
}

func TestArticleRepository_Delete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestArticleRepository_Delete_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE").WillReturnError(fmt.Errorf("delete error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestArticleRepository_IncrViewCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	mock.ExpectExec("INSERT INTO article_attributes").WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.IncrViewCount(context.Background(), 1)
	require.NoError(t, err)
}

func TestArticleRepository_IncrViewCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)

	mock.ExpectExec("INSERT INTO article_attributes").WillReturnError(fmt.Errorf("exec error"))

	err := repo.IncrViewCount(context.Background(), 1)
	assert.Error(t, err)
}

func TestArticleRepository_IncrLikeCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)
	mock.ExpectExec("INSERT INTO article_attributes").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, repo.IncrLikeCount(context.Background(), 1))
}

func TestArticleRepository_IncrLikeCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)
	mock.ExpectExec("INSERT INTO article_attributes").WillReturnError(fmt.Errorf("exec error"))
	assert.Error(t, repo.IncrLikeCount(context.Background(), 1))
}

func TestArticleRepository_DecrLikeCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)
	mock.ExpectExec("INSERT INTO article_attributes").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, repo.DecrLikeCount(context.Background(), 1))
}

func TestArticleRepository_DecrLikeCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)
	mock.ExpectExec("INSERT INTO article_attributes").WillReturnError(fmt.Errorf("exec error"))
	assert.Error(t, repo.DecrLikeCount(context.Background(), 1))
}

func TestArticleRepository_IncrCollectCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)
	mock.ExpectExec("INSERT INTO article_attributes").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, repo.IncrCollectCount(context.Background(), 1))
}

func TestArticleRepository_IncrCollectCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)
	mock.ExpectExec("INSERT INTO article_attributes").WillReturnError(fmt.Errorf("exec error"))
	assert.Error(t, repo.IncrCollectCount(context.Background(), 1))
}

func TestArticleRepository_DecrCollectCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)
	mock.ExpectExec("INSERT INTO article_attributes").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, repo.DecrCollectCount(context.Background(), 1))
}

func TestArticleRepository_DecrCollectCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)
	mock.ExpectExec("INSERT INTO article_attributes").WillReturnError(fmt.Errorf("exec error"))
	assert.Error(t, repo.DecrCollectCount(context.Background(), 1))
}

func TestArticleRepository_IncrCommentCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)
	mock.ExpectExec("INSERT INTO article_attributes").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, repo.IncrCommentCount(context.Background(), 1))
}

func TestArticleRepository_IncrCommentCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)
	mock.ExpectExec("INSERT INTO article_attributes").WillReturnError(fmt.Errorf("exec error"))
	assert.Error(t, repo.IncrCommentCount(context.Background(), 1))
}

func TestArticleRepository_DecrCommentCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)
	mock.ExpectExec("INSERT INTO article_attributes").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, repo.DecrCommentCount(context.Background(), 1))
}

func TestArticleRepository_DecrCommentCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)
	mock.ExpectExec("INSERT INTO article_attributes").WillReturnError(fmt.Errorf("exec error"))
	assert.Error(t, repo.DecrCommentCount(context.Background(), 1))
}

func TestArticleRepository_IncrShareCount_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)
	mock.ExpectExec("INSERT INTO article_attributes").WillReturnResult(sqlmock.NewResult(1, 1))
	require.NoError(t, repo.IncrShareCount(context.Background(), 1))
}

func TestArticleRepository_IncrShareCount_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)
	mock.ExpectExec("INSERT INTO article_attributes").WillReturnError(fmt.Errorf("exec error"))
	assert.Error(t, repo.IncrShareCount(context.Background(), 1))
}

func TestArticleRepository_HasAssociations_Success(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(2))
	ok, err := repo.HasAssociations(context.Background(), 1)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestArticleRepository_HasAssociations_Error(t *testing.T) {
	db, mock := newTestDB(t)
	repo := NewArticleRepository(db)
	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("query error"))
	_, err := repo.HasAssociations(context.Background(), 1)
	assert.Error(t, err)
}
