package service

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/testutil"
)

type articleAttachmentRepoStub struct {
	listFn    func(ctx context.Context, articleID uint64) ([]uint64, error)
	replaceFn func(ctx context.Context, articleID uint64, fileIDs []uint64) error
}

func (s *articleAttachmentRepoStub) ListFileIDs(ctx context.Context, articleID uint64) ([]uint64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, articleID)
	}
	return nil, nil
}

func (s *articleAttachmentRepoStub) Replace(ctx context.Context, articleID uint64, fileIDs []uint64) error {
	if s.replaceFn != nil {
		return s.replaceFn(ctx, articleID, fileIDs)
	}
	return nil
}

func (s *articleAttachmentRepoStub) ListByArticleID(ctx context.Context, articleID uint64) ([]*entity.ArticleAttachment, error) {
	ids, err := s.ListFileIDs(ctx, articleID)
	if err != nil {
		return nil, err
	}
	rows := make([]*entity.ArticleAttachment, 0, len(ids))
	for i, id := range ids {
		rows = append(rows, &entity.ArticleAttachment{ID: uint64(i + 1), ArticleID: articleID, FileID: id})
	}
	return rows, nil
}

func (s *articleAttachmentRepoStub) GetByFileID(_ context.Context, _ uint64) (*entity.ArticleAttachment, error) {
	return nil, nil
}

type courseAttachmentRepoStub struct {
	listFn    func(ctx context.Context, courseID uint64) ([]uint64, error)
	replaceFn func(ctx context.Context, courseID uint64, fileIDs []uint64) error
}

func (s *courseAttachmentRepoStub) ListFileIDs(ctx context.Context, courseID uint64) ([]uint64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, courseID)
	}
	return nil, nil
}

func (s *courseAttachmentRepoStub) Replace(ctx context.Context, courseID uint64, fileIDs []uint64) error {
	if s.replaceFn != nil {
		return s.replaceFn(ctx, courseID, fileIDs)
	}
	return nil
}

func (s *courseAttachmentRepoStub) ListByCourseID(ctx context.Context, courseID uint64) ([]*entity.CourseAttachment, error) {
	ids, err := s.ListFileIDs(ctx, courseID)
	if err != nil {
		return nil, err
	}
	rows := make([]*entity.CourseAttachment, 0, len(ids))
	for i, id := range ids {
		rows = append(rows, &entity.CourseAttachment{ID: uint64(i + 1), CourseID: courseID, FileID: id})
	}
	return rows, nil
}

func (s *courseAttachmentRepoStub) GetByFileID(_ context.Context, _ uint64) (*entity.CourseAttachment, error) {
	return nil, nil
}

type courseUnitAttachmentRepoStub struct {
	listFn    func(ctx context.Context, unitID uint64) ([]uint64, error)
	replaceFn func(ctx context.Context, unitID uint64, fileIDs []uint64) error
}

func (s *courseUnitAttachmentRepoStub) ListFileIDs(ctx context.Context, unitID uint64) ([]uint64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, unitID)
	}
	return nil, nil
}

func (s *courseUnitAttachmentRepoStub) Replace(ctx context.Context, unitID uint64, fileIDs []uint64) error {
	if s.replaceFn != nil {
		return s.replaceFn(ctx, unitID, fileIDs)
	}
	return nil
}

func (s *courseUnitAttachmentRepoStub) ListByUnitID(ctx context.Context, unitID uint64) ([]*entity.CourseUnitAttachment, error) {
	ids, err := s.ListFileIDs(ctx, unitID)
	if err != nil {
		return nil, err
	}
	rows := make([]*entity.CourseUnitAttachment, 0, len(ids))
	for i, id := range ids {
		rows = append(rows, &entity.CourseUnitAttachment{ID: uint64(i + 1), UnitID: unitID, FileID: id})
	}
	return rows, nil
}

func (s *courseUnitAttachmentRepoStub) GetByFileID(_ context.Context, _ uint64) (*entity.CourseUnitAttachment, error) {
	return nil, nil
}

type cosRemoverStub struct {
	deleteFn func(ctx context.Context, key string) error
}

func (s *cosRemoverStub) DeleteObject(ctx context.Context, key string) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, key)
	}
	return nil
}

// ==================== ModuleService ====================

func newModuleService(modRepo *testutil.MockModuleRepository, pageRepo *testutil.MockModulePageRepository) ModuleService {
	return NewModuleService(modRepo, pageRepo, logrus.New())
}

func TestModuleService_List(t *testing.T) {
	mods := []*entity.Module{{ID: 1, Title: "A"}}
	repo := &testutil.MockModuleRepository{
		ListFn: func(_ context.Context, status *int8) ([]*entity.Module, error) {
			return mods, nil
		},
	}
	svc := newModuleService(repo, nil)
	got, err := svc.List(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, mods, got)
}

func TestModuleService_Create(t *testing.T) {
	repo := &testutil.MockModuleRepository{
		CreateFn: func(_ context.Context, m *entity.Module) error {
			m.ID = 1
			return nil
		},
	}
	svc := newModuleService(repo, nil)
	id, err := svc.Create(context.Background(), &dto.CreateModuleRequest{Title: "Test", SortOrder: 1})
	require.NoError(t, err)
	assert.Equal(t, uint(1), id)
}

func TestModuleService_Update_Found(t *testing.T) {
	mod := &entity.Module{ID: 1, Title: "Old"}
	repo := &testutil.MockModuleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Module, error) {
			return mod, nil
		},
		UpdateFn: func(_ context.Context, m *entity.Module) error {
			return nil
		},
	}
	svc := newModuleService(repo, nil)
	err := svc.Update(context.Background(), 1, &dto.CreateModuleRequest{Title: "New"})
	require.NoError(t, err)
}

func TestModuleService_Update_NotFound(t *testing.T) {
	repo := &testutil.MockModuleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Module, error) {
			return nil, nil
		},
	}
	svc := newModuleService(repo, nil)
	err := svc.Update(context.Background(), 1, &dto.CreateModuleRequest{})
	require.Error(t, err)
}

func TestModuleService_Update_DBError(t *testing.T) {
	repo := &testutil.MockModuleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Module, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := newModuleService(repo, nil)
	err := svc.Update(context.Background(), 1, &dto.CreateModuleRequest{})
	require.Error(t, err)
}

func TestModuleService_Delete_Found(t *testing.T) {
	repo := &testutil.MockModuleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Module, error) {
			return &entity.Module{ID: 1}, nil
		},
		DeleteFn: func(_ context.Context, id uint) error {
			return nil
		},
	}
	svc := newModuleService(repo, nil)
	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestModuleService_Delete_NotFound(t *testing.T) {
	repo := &testutil.MockModuleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Module, error) {
			return nil, nil
		},
	}
	svc := newModuleService(repo, nil)
	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
}

func TestModuleService_Delete_WithAssociations(t *testing.T) {
	repo := &testutil.MockModuleRepository{
		GetByIDFn:         func(_ context.Context, id uint) (*entity.Module, error) { return &entity.Module{ID: id}, nil },
		HasAssociationsFn: func(_ context.Context, id uint) (bool, error) { return true, nil },
	}
	svc := newModuleService(repo, nil)
	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
}

func TestModuleService_GetPages(t *testing.T) {
	pages := []*entity.ModulePage{{ID: 1, Title: "P1"}}
	pageRepo := &testutil.MockModulePageRepository{
		ListByModuleIDFn: func(_ context.Context, moduleID uint) ([]*entity.ModulePage, error) {
			return pages, nil
		},
	}
	svc := newModuleService(nil, pageRepo)
	got, err := svc.GetPages(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, pages, got)
}

func TestModuleService_CreatePage(t *testing.T) {
	pageRepo := &testutil.MockModulePageRepository{
		CreateFn: func(_ context.Context, p *entity.ModulePage) error {
			p.ID = 1
			return nil
		},
	}
	svc := newModuleService(nil, pageRepo)
	id, err := svc.CreatePage(context.Background(), 1, &dto.CreateModulePageRequest{
		Title: "Page 1", Content: "hello",
	})
	require.NoError(t, err)
	assert.Equal(t, uint(1), id)
}

func TestModuleService_CreatePage_WithContentType(t *testing.T) {
	var saved *entity.ModulePage
	pageRepo := &testutil.MockModulePageRepository{
		CreateFn: func(_ context.Context, p *entity.ModulePage) error {
			saved = p
			p.ID = 2
			return nil
		},
	}
	svc := newModuleService(nil, pageRepo)
	id, err := svc.CreatePage(context.Background(), 1, &dto.CreateModulePageRequest{
		Title: "Page 1", Content: "hello", ContentType: 2,
	})
	require.NoError(t, err)
	assert.Equal(t, uint(2), id)
	require.NotNil(t, saved)
	assert.Equal(t, int8(1), saved.ContentType)
}

func TestModuleService_UpdatePage_Found(t *testing.T) {
	page := &entity.ModulePage{ID: 1, ModuleID: 1, Title: "Old"}
	pageRepo := &testutil.MockModulePageRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.ModulePage, error) {
			return page, nil
		},
		UpdateFn: func(_ context.Context, p *entity.ModulePage) error {
			return nil
		},
	}
	svc := newModuleService(nil, pageRepo)
	err := svc.UpdatePage(context.Background(), 1, 1, &dto.CreateModulePageRequest{Title: "New"})
	require.NoError(t, err)
}

func TestModuleService_UpdatePage_NotFound(t *testing.T) {
	pageRepo := &testutil.MockModulePageRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.ModulePage, error) {
			return nil, nil
		},
	}
	svc := newModuleService(nil, pageRepo)
	err := svc.UpdatePage(context.Background(), 1, 1, &dto.CreateModulePageRequest{})
	require.Error(t, err)
}

func TestModuleService_UpdatePage_WrongModule(t *testing.T) {
	page := &entity.ModulePage{ID: 1, ModuleID: 99}
	pageRepo := &testutil.MockModulePageRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.ModulePage, error) {
			return page, nil
		},
	}
	svc := newModuleService(nil, pageRepo)
	err := svc.UpdatePage(context.Background(), 1, 1, &dto.CreateModulePageRequest{})
	require.Error(t, err)
}

func TestModuleService_UpdatePage_DBError(t *testing.T) {
	pageRepo := &testutil.MockModulePageRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.ModulePage, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := newModuleService(nil, pageRepo)
	err := svc.UpdatePage(context.Background(), 1, 1, &dto.CreateModulePageRequest{})
	require.Error(t, err)
}

func TestModuleService_DeletePage_Found(t *testing.T) {
	page := &entity.ModulePage{ID: 1, ModuleID: 1}
	pageRepo := &testutil.MockModulePageRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.ModulePage, error) {
			return page, nil
		},
		DeleteFn: func(_ context.Context, id uint) error {
			return nil
		},
	}
	svc := newModuleService(nil, pageRepo)
	err := svc.DeletePage(context.Background(), 1, 1)
	require.NoError(t, err)
}

func TestModuleService_DeletePage_NotFound(t *testing.T) {
	pageRepo := &testutil.MockModulePageRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.ModulePage, error) {
			return nil, nil
		},
	}
	svc := newModuleService(nil, pageRepo)
	err := svc.DeletePage(context.Background(), 1, 1)
	require.Error(t, err)
}

func TestModuleService_DeletePage_WrongModule(t *testing.T) {
	page := &entity.ModulePage{ID: 1, ModuleID: 99}
	pageRepo := &testutil.MockModulePageRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.ModulePage, error) {
			return page, nil
		},
	}
	svc := newModuleService(nil, pageRepo)
	err := svc.DeletePage(context.Background(), 1, 1)
	require.Error(t, err)
}

// ==================== ArticleService ====================

func newArticleService(articleRepo *testutil.MockArticleRepository, permRepo *testutil.MockContentPermissionRepository) ArticleService {
	return NewArticleService(articleRepo, permRepo, logrus.New())
}

func TestArticleService_List(t *testing.T) {
	articles := []*entity.Article{{ID: 1}}
	repo := &testutil.MockArticleRepository{
		ListFn: func(_ context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8, sort string) ([]*entity.Article, int64, error) {
			return articles, 1, nil
		},
	}
	svc := newArticleService(repo, nil)
	got, total, err := svc.List(context.Background(), 1, 10, "", nil, "", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, articles, got)
}

func TestArticleService_GetByID_Found(t *testing.T) {
	article := &entity.Article{ID: 1, Status: 1}
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return article, nil
		},
		IncrViewCountFn: func(_ context.Context, id uint64) error {
			return nil
		},
	}
	svc := newArticleService(repo, nil)
	got, err := svc.GetByID(context.Background(), 1, nil)
	require.NoError(t, err)
	assert.Equal(t, article, got)
}

func TestArticleService_GetByID_NotFound(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return nil, nil
		},
	}
	svc := newArticleService(repo, nil)
	_, err := svc.GetByID(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestArticleService_GetByID_NotPublished(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return &entity.Article{ID: 1, Status: 0}, nil
		},
	}
	svc := newArticleService(repo, nil)
	_, err := svc.GetByID(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestArticleService_GetByID_DBError(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := newArticleService(repo, nil)
	_, err := svc.GetByID(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestArticleService_AdminList(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		ListFn: func(_ context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8, sort string) ([]*entity.Article, int64, error) {
			return []*entity.Article{{ID: 1}}, 1, nil
		},
	}
	svc := newArticleService(repo, nil)
	got, total, err := svc.AdminList(context.Background(), 1, 10, "", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, got, 1)
}

func TestArticleService_AdminGetByID_Found(t *testing.T) {
	article := &entity.Article{ID: 1}
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return article, nil
		},
	}
	svc := newArticleService(repo, nil)
	got, err := svc.AdminGetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, article, got)
}

func TestArticleService_AdminGetByID_NotFound(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return nil, nil
		},
	}
	svc := newArticleService(repo, nil)
	_, err := svc.AdminGetByID(context.Background(), 1)
	require.Error(t, err)
}

func TestArticleService_Create_WithPermissions(t *testing.T) {
	var saved *entity.Article
	repo := &testutil.MockArticleRepository{
		CreateFn: func(_ context.Context, a *entity.Article) error {
			saved = a
			a.ID = 1
			return nil
		},
	}
	permRepo := &testutil.MockContentPermissionRepository{
		SetContentPermsFn: func(_ context.Context, contentType int8, contentID uint64, roleIDs []uint) error {
			return nil
		},
	}
	svc := newArticleService(repo, permRepo)
	id, err := svc.Create(context.Background(), &dto.CreateArticleRequest{
		Title: "Test", Status: 1, RolePermissions: []uint{1},
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), id)
	require.NotNil(t, saved)
	assert.Equal(t, int8(1), saved.ContentType)
}

func TestArticleService_Create_NoPermissions(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		CreateFn: func(_ context.Context, a *entity.Article) error {
			a.ID = 2
			return nil
		},
	}
	svc := newArticleService(repo, nil)
	id, err := svc.Create(context.Background(), &dto.CreateArticleRequest{
		Title: "Test", Status: 0,
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), id)
}

func TestArticleService_Create_Error(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		CreateFn: func(_ context.Context, a *entity.Article) error {
			return apperrors.NewInternal("db error", nil)
		},
	}
	svc := newArticleService(repo, nil)
	_, err := svc.Create(context.Background(), &dto.CreateArticleRequest{Title: "Test"}, 1)
	require.Error(t, err)
}

func TestArticleService_Create_MaskSensitiveWords(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		CreateFn: func(_ context.Context, a *entity.Article) error {
			a.ID = 3
			assert.Equal(t, "含***标题", a.Title)
			assert.Equal(t, "摘要***", a.Summary)
			assert.Equal(t, "正文***内容", a.Content)
			return nil
		},
	}
	wordsRepo := &testutil.MockSensitiveWordRepository{
		ListEnabledWordsFn: func(_ context.Context) ([]string, error) {
			return []string{"敏感词"}, nil
		},
	}
	svc := NewArticleService(repo, nil, logrus.New(), wordsRepo)
	id, err := svc.Create(context.Background(), &dto.CreateArticleRequest{
		Title:   "含敏感词标题",
		Summary: "摘要敏感词",
		Content: "正文敏感词内容",
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(3), id)
}

func TestArticleService_Update_Success(t *testing.T) {
	article := &entity.Article{ID: 1}
	var updated *entity.Article
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return article, nil
		},
		UpdateFn: func(_ context.Context, a *entity.Article) error {
			updated = a
			return nil
		},
	}
	permRepo := &testutil.MockContentPermissionRepository{
		SetContentPermsFn: func(_ context.Context, contentType int8, contentID uint64, roleIDs []uint) error {
			return nil
		},
	}
	svc := newArticleService(repo, permRepo)
	err := svc.Update(context.Background(), 1, &dto.UpdateArticleRequest{Title: "New"})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, int8(1), updated.ContentType)
}

func TestArticleService_Update_NotFound(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return nil, nil
		},
	}
	svc := newArticleService(repo, nil)
	err := svc.Update(context.Background(), 1, &dto.UpdateArticleRequest{})
	require.Error(t, err)
}

func TestArticleService_Update_DBError(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := newArticleService(repo, nil)
	err := svc.Update(context.Background(), 1, &dto.UpdateArticleRequest{})
	require.Error(t, err)
}

func TestArticleService_Delete_Success(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return &entity.Article{ID: 1}, nil
		},
		DeleteCascadeFn: func(_ context.Context, id uint64, fileIDs []uint64) error {
			return nil
		},
	}
	svc := newArticleService(repo, nil)
	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestArticleService_Delete_NotFound(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return nil, nil
		},
	}
	svc := newArticleService(repo, nil)
	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
}

func TestArticleService_Delete_WithAssociations(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) { return &entity.Article{ID: id}, nil },
		DeleteCascadeFn: func(_ context.Context, id uint64, fileIDs []uint64) error {
			return nil
		},
	}
	svc := newArticleService(repo, nil)
	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestArticleService_Publish_Success(t *testing.T) {
	article := &entity.Article{ID: 1, Status: 0}
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return article, nil
		},
		UpdateFn: func(_ context.Context, a *entity.Article) error {
			return nil
		},
	}
	svc := newArticleService(repo, nil)
	err := svc.Publish(context.Background(), 1, &dto.PublishArticleRequest{Status: 1})
	require.NoError(t, err)
}

func TestArticleService_Publish_NotFound(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return nil, nil
		},
	}
	svc := newArticleService(repo, nil)
	err := svc.Publish(context.Background(), 1, &dto.PublishArticleRequest{Status: 1})
	require.Error(t, err)
}

func TestArticleService_Publish_Unpublish(t *testing.T) {
	article := &entity.Article{ID: 1, Status: 1}
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return article, nil
		},
		UpdateFn: func(_ context.Context, a *entity.Article) error {
			return nil
		},
	}
	svc := newArticleService(repo, nil)
	err := svc.Publish(context.Background(), 1, &dto.PublishArticleRequest{Status: 0})
	require.NoError(t, err)
}

// ==================== CourseService ====================

func newCourseService(
	courseRepo *testutil.MockCourseRepository,
	unitRepo *testutil.MockCourseUnitRepository,
	permRepo *testutil.MockContentPermissionRepository,
) CourseService {
	if unitRepo == nil {
		unitRepo = &testutil.MockCourseUnitRepository{}
	}
	return NewCourseService(courseRepo, unitRepo, permRepo, logrus.New())
}

func TestCourseService_List(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		ListFn: func(_ context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8, isFree *bool) ([]*entity.Course, int64, error) {
			return []*entity.Course{{ID: 1}}, 1, nil
		},
	}
	svc := newCourseService(repo, nil, nil)
	got, total, err := svc.List(context.Background(), 1, 10, "", nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, got, 1)
}

func TestCourseService_GetByID_Found(t *testing.T) {
	course := &entity.Course{ID: 1, Status: 1}
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return course, nil
		},
		IncrViewCountFn: func(_ context.Context, id uint64) error {
			return nil
		},
	}
	svc := newCourseService(repo, nil, nil)
	got, err := svc.GetByID(context.Background(), 1, nil)
	require.NoError(t, err)
	assert.Equal(t, course, got)
}

func TestCourseService_GetByID_NotFound(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return nil, nil
		},
	}
	svc := newCourseService(repo, nil, nil)
	_, err := svc.GetByID(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestCourseService_GetByID_NotPublished(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return &entity.Course{ID: 1, Status: 0}, nil
		},
	}
	svc := newCourseService(repo, nil, nil)
	_, err := svc.GetByID(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestCourseService_GetByID_DBError(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := newCourseService(repo, nil, nil)
	_, err := svc.GetByID(context.Background(), 1, nil)
	require.Error(t, err)
}

func TestCourseService_AdminList(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		ListFn: func(_ context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8, isFree *bool) ([]*entity.Course, int64, error) {
			return []*entity.Course{{ID: 1}}, 1, nil
		},
	}
	svc := newCourseService(repo, nil, nil)
	got, total, err := svc.AdminList(context.Background(), 1, 10, "", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, got, 1)
}

func TestCourseService_AdminGetByID_Found(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return &entity.Course{ID: 1}, nil
		},
	}
	svc := newCourseService(repo, nil, nil)
	got, err := svc.AdminGetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), got.ID)
}

func TestCourseService_AdminGetByID_NotFound(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return nil, nil
		},
	}
	svc := newCourseService(repo, nil, nil)
	_, err := svc.AdminGetByID(context.Background(), 1)
	require.Error(t, err)
}

func TestCourseService_Create_WithPermissions(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		CreateFn: func(_ context.Context, c *entity.Course) error {
			c.ID = 1
			return nil
		},
	}
	permRepo := &testutil.MockContentPermissionRepository{
		SetContentPermsFn: func(_ context.Context, contentType int8, contentID uint64, roleIDs []uint) error {
			return nil
		},
	}
	svc := newCourseService(repo, nil, permRepo)
	id, err := svc.Create(context.Background(), &dto.CreateCourseRequest{
		Title: "Course 1", Status: 1, RolePermissions: []uint{1},
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), id)
}

func TestCourseService_Create_NoPermissions(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		CreateFn: func(_ context.Context, c *entity.Course) error {
			c.ID = 2
			return nil
		},
	}
	svc := newCourseService(repo, nil, nil)
	id, err := svc.Create(context.Background(), &dto.CreateCourseRequest{Title: "Course 1"}, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), id)
}

func TestCourseService_Create_MaskSensitiveWords(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		CreateFn: func(_ context.Context, c *entity.Course) error {
			c.ID = 3
			assert.Equal(t, "***课程", c.Title)
			assert.Equal(t, "课程***简介", c.Description)
			return nil
		},
	}
	wordsRepo := &testutil.MockSensitiveWordRepository{
		ListEnabledWordsFn: func(_ context.Context) ([]string, error) {
			return []string{"敏感词"}, nil
		},
	}
	svc := NewCourseService(repo, nil, nil, logrus.New(), wordsRepo)
	id, err := svc.Create(context.Background(), &dto.CreateCourseRequest{
		Title:       "敏感词课程",
		Description: "课程敏感词简介",
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(3), id)
}

func TestCourseService_Update_Success(t *testing.T) {
	course := &entity.Course{ID: 1}
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return course, nil
		},
		UpdateFn: func(_ context.Context, c *entity.Course) error {
			return nil
		},
	}
	permRepo := &testutil.MockContentPermissionRepository{
		SetContentPermsFn: func(_ context.Context, contentType int8, contentID uint64, roleIDs []uint) error {
			return nil
		},
	}
	svc := newCourseService(repo, nil, permRepo)
	err := svc.Update(context.Background(), 1, &dto.UpdateCourseRequest{Title: "New"})
	require.NoError(t, err)
}

func TestCourseService_Update_NotFound(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return nil, nil
		},
	}
	svc := newCourseService(repo, nil, nil)
	err := svc.Update(context.Background(), 1, &dto.UpdateCourseRequest{})
	require.Error(t, err)
}

func TestCourseService_Delete_Success(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return &entity.Course{ID: 1}, nil
		},
		DeleteCascadeFn: func(_ context.Context, id uint64, fileIDs []uint64) error {
			return nil
		},
	}
	svc := newCourseService(repo, nil, nil)
	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestCourseService_Delete_NotFound(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return nil, nil
		},
	}
	svc := newCourseService(repo, nil, nil)
	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
}

func TestCourseService_Delete_WithAssociations(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) { return &entity.Course{ID: id}, nil },
		DeleteCascadeFn: func(_ context.Context, id uint64, fileIDs []uint64) error {
			return nil
		},
	}
	svc := newCourseService(repo, nil, nil)
	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestCourseService_Publish_Success(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return &entity.Course{ID: 1}, nil
		},
		UpdateFn: func(_ context.Context, c *entity.Course) error {
			return nil
		},
	}
	svc := newCourseService(repo, nil, nil)
	err := svc.Publish(context.Background(), 1, &dto.PublishCourseRequest{Status: 1})
	require.NoError(t, err)
}

func TestCourseService_Publish_Unpublish(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return &entity.Course{ID: 1, Status: 1}, nil
		},
		UpdateFn: func(_ context.Context, c *entity.Course) error {
			return nil
		},
	}
	svc := newCourseService(repo, nil, nil)
	err := svc.Publish(context.Background(), 1, &dto.PublishCourseRequest{Status: 0})
	require.NoError(t, err)
}

func TestCourseService_Publish_NotFound(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return nil, nil
		},
	}
	svc := newCourseService(repo, nil, nil)
	err := svc.Publish(context.Background(), 1, &dto.PublishCourseRequest{Status: 1})
	require.Error(t, err)
}

func TestCourseService_GetUnits(t *testing.T) {
	unitRepo := &testutil.MockCourseUnitRepository{
		ListByCourseIDFn: func(_ context.Context, courseID uint64) ([]*entity.CourseUnit, error) {
			return []*entity.CourseUnit{{ID: 1, CourseID: courseID}}, nil
		},
	}
	svc := newCourseService(nil, unitRepo, nil)
	units, err := svc.GetUnits(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, units, 1)
}

func TestCourseService_CreateUnit(t *testing.T) {
	unitRepo := &testutil.MockCourseUnitRepository{
		CreateFn: func(_ context.Context, u *entity.CourseUnit) error {
			u.ID = 1
			return nil
		},
	}
	svc := newCourseService(nil, unitRepo, nil)
	id, err := svc.CreateUnit(context.Background(), 1, &dto.CreateCourseUnitRequest{Title: "Unit 1"})
	require.NoError(t, err)
	assert.Equal(t, uint64(1), id)
}

func TestCourseService_UpdateUnit_Success(t *testing.T) {
	unit := &entity.CourseUnit{ID: 1, CourseID: 1}
	unitRepo := &testutil.MockCourseUnitRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.CourseUnit, error) {
			return unit, nil
		},
		UpdateFn: func(_ context.Context, u *entity.CourseUnit) error {
			return nil
		},
	}
	svc := newCourseService(nil, unitRepo, nil)
	err := svc.UpdateUnit(context.Background(), 1, 1, &dto.CreateCourseUnitRequest{Title: "Updated"})
	require.NoError(t, err)
}

func TestCourseService_UpdateUnit_NotFound(t *testing.T) {
	unitRepo := &testutil.MockCourseUnitRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.CourseUnit, error) {
			return nil, nil
		},
	}
	svc := newCourseService(nil, unitRepo, nil)
	err := svc.UpdateUnit(context.Background(), 1, 1, &dto.CreateCourseUnitRequest{})
	require.Error(t, err)
}

func TestCourseService_UpdateUnit_WrongCourse(t *testing.T) {
	unit := &entity.CourseUnit{ID: 1, CourseID: 99}
	unitRepo := &testutil.MockCourseUnitRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.CourseUnit, error) {
			return unit, nil
		},
	}
	svc := newCourseService(nil, unitRepo, nil)
	err := svc.UpdateUnit(context.Background(), 1, 1, &dto.CreateCourseUnitRequest{})
	require.Error(t, err)
}

func TestCourseService_UpdateUnit_DBError(t *testing.T) {
	unitRepo := &testutil.MockCourseUnitRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.CourseUnit, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := newCourseService(nil, unitRepo, nil)
	err := svc.UpdateUnit(context.Background(), 1, 1, &dto.CreateCourseUnitRequest{})
	require.Error(t, err)
}

func TestCourseService_DeleteUnit_Success(t *testing.T) {
	unit := &entity.CourseUnit{ID: 1, CourseID: 1}
	unitRepo := &testutil.MockCourseUnitRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.CourseUnit, error) {
			return unit, nil
		},
		DeleteCascadeFn: func(_ context.Context, id uint64, fileIDs []uint64) error {
			return nil
		},
	}
	svc := newCourseService(nil, unitRepo, nil)
	err := svc.DeleteUnit(context.Background(), 1, 1)
	require.NoError(t, err)
}

func TestCourseService_DeleteUnit_NotFound(t *testing.T) {
	unitRepo := &testutil.MockCourseUnitRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.CourseUnit, error) {
			return nil, nil
		},
	}
	svc := newCourseService(nil, unitRepo, nil)
	err := svc.DeleteUnit(context.Background(), 1, 1)
	require.Error(t, err)
}

func TestCourseService_DeleteUnit_HasStudyRecords(t *testing.T) {
	unitRepo := &testutil.MockCourseUnitRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.CourseUnit, error) {
			return &entity.CourseUnit{ID: id, CourseID: 1}, nil
		},
		DeleteCascadeFn: func(_ context.Context, id uint64, fileIDs []uint64) error {
			return nil
		},
	}
	svc := newCourseService(nil, unitRepo, nil)
	err := svc.DeleteUnit(context.Background(), 1, 1)
	require.NoError(t, err)
}

func TestArticleService_Delete_CascadeWithCOS(t *testing.T) {
	deletedKeys := make([]string, 0, 2)
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			cover := uint64(10)
			return &entity.Article{ID: id, CoverFileID: &cover}, nil
		},
		DeleteCascadeFn: func(_ context.Context, id uint64, fileIDs []uint64) error {
			assert.ElementsMatch(t, []uint64{10, 11}, fileIDs)
			return nil
		},
	}
	attachRepo := &articleAttachmentRepoStub{
		listFn: func(_ context.Context, articleID uint64) ([]uint64, error) {
			return []uint64{11, 11}, nil
		},
	}
	fileRepo := &testutil.MockFileRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.File, error) {
			if id == 10 {
				return &entity.File{ID: 10, Key: "k/cover"}, nil
			}
			return &entity.File{ID: 11, Key: "k/attach"}, nil
		},
	}
	remover := &cosRemoverStub{
		deleteFn: func(_ context.Context, key string) error {
			deletedKeys = append(deletedKeys, key)
			return nil
		},
	}
	svc := NewArticleService(repo, nil, logrus.New(), attachRepo, fileRepo, remover)
	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"k/cover", "k/attach"}, deletedKeys)
}

func TestCourseService_DeleteUnit_CascadeWithCOS(t *testing.T) {
	deletedKeys := make([]string, 0, 2)
	unitRepo := &testutil.MockCourseUnitRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.CourseUnit, error) {
			video := uint64(31)
			return &entity.CourseUnit{ID: id, CourseID: 1, VideoFileID: &video}, nil
		},
		DeleteCascadeFn: func(_ context.Context, id uint64, fileIDs []uint64) error {
			assert.ElementsMatch(t, []uint64{31, 32}, fileIDs)
			return nil
		},
	}
	unitAttachRepo := &courseUnitAttachmentRepoStub{
		listFn: func(_ context.Context, unitID uint64) ([]uint64, error) {
			return []uint64{32, 32}, nil
		},
	}
	fileRepo := &testutil.MockFileRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.File, error) {
			if id == 31 {
				return &entity.File{ID: 31, Key: "k/video"}, nil
			}
			return &entity.File{ID: 32, Key: "k/unit-attach"}, nil
		},
	}
	remover := &cosRemoverStub{
		deleteFn: func(_ context.Context, key string) error {
			deletedKeys = append(deletedKeys, key)
			return nil
		},
	}
	svc := NewCourseService(nil, unitRepo, nil, logrus.New(), unitAttachRepo, fileRepo, remover)
	err := svc.DeleteUnit(context.Background(), 1, 9)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"k/video", "k/unit-attach"}, deletedKeys)
}

func TestArticleService_Pin_Success(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return &entity.Article{ID: id}, nil
		},
		UpdateFn: func(_ context.Context, a *entity.Article) error { return nil },
	}
	svc := newArticleService(repo, &testutil.MockContentPermissionRepository{
		SetContentPermsFn: func(_ context.Context, contentType int8, contentID uint64, roleIDs []uint) error { return nil },
	})
	err := svc.Pin(context.Background(), 1, &dto.PinArticleRequest{SortOrder: 100})
	require.NoError(t, err)
}

func TestArticleService_Pin_NotFoundAndGetError(t *testing.T) {
	repoNotFound := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) { return nil, nil },
	}
	svcNotFound := newArticleService(repoNotFound, nil)
	require.Error(t, svcNotFound.Pin(context.Background(), 1, &dto.PinArticleRequest{SortOrder: 1}))

	repoErr := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svcErr := newArticleService(repoErr, nil)
	require.Error(t, svcErr.Pin(context.Background(), 1, &dto.PinArticleRequest{SortOrder: 1}))
}

func TestArticleService_Copy_Success(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return &entity.Article{ID: id, Title: "A", AuthorID: 1, ModuleID: 1, ContentType: 1}, nil
		},
		CreateFn: func(_ context.Context, a *entity.Article) error {
			a.ID = 99
			return nil
		},
	}
	permRepo := &testutil.MockContentPermissionRepository{
		GetByContentFn: func(_ context.Context, contentType int8, contentID uint64) ([]*entity.ContentPermission, error) {
			return nil, nil
		},
		SetContentPermsFn: func(_ context.Context, contentType int8, contentID uint64, roleIDs []uint) error { return nil },
	}
	svc := newArticleService(repo, permRepo)
	id, err := svc.Copy(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.Equal(t, uint64(99), id)
}

func TestArticleService_Copy_ErrorBranches(t *testing.T) {
	repoGetErr := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	_, err := newArticleService(repoGetErr, nil).Copy(context.Background(), 1, 2)
	require.Error(t, err)

	repoNotFound := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) { return nil, nil },
	}
	_, err = newArticleService(repoNotFound, nil).Copy(context.Background(), 1, 2)
	require.Error(t, err)

	repoCreateErr := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return &entity.Article{ID: id, Title: "A", ModuleID: 1, ContentType: 1}, nil
		},
		CreateFn: func(_ context.Context, a *entity.Article) error { return apperrors.NewInternal("db", nil) },
	}
	_, err = newArticleService(repoCreateErr, nil).Copy(context.Background(), 1, 2)
	require.Error(t, err)
}

func TestArticleService_BindAttachmentIDs(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return &entity.Article{ID: id, Title: "A", Status: 1}, nil
		},
	}
	attachmentRepo := &articleAttachmentRepoStub{
		listFn: func(_ context.Context, articleID uint64) ([]uint64, error) {
			return []uint64{11, 12}, nil
		},
	}
	svc := NewArticleService(repo, nil, logrus.New(), attachmentRepo)
	article, err := svc.AdminGetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, []uint64{11, 12}, article.AttachmentFileIDs)
}

func TestArticleService_Copy_WithAttachmentAndPermissions(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return &entity.Article{ID: id, Title: "A", ModuleID: 1, ContentType: 1}, nil
		},
		CreateFn: func(_ context.Context, a *entity.Article) error {
			a.ID = 101
			return nil
		},
	}
	attachmentReplaced := false
	attachmentRepo := &articleAttachmentRepoStub{
		listFn: func(_ context.Context, articleID uint64) ([]uint64, error) {
			assert.Equal(t, uint64(1), articleID)
			return []uint64{1, 2}, nil
		},
		replaceFn: func(_ context.Context, articleID uint64, fileIDs []uint64) error {
			attachmentReplaced = true
			assert.Equal(t, uint64(101), articleID)
			assert.Equal(t, []uint64{1, 2}, fileIDs)
			return nil
		},
	}
	permSet := false
	permRepo := &testutil.MockContentPermissionRepository{
		GetByContentFn: func(_ context.Context, contentType int8, contentID uint64) ([]*entity.ContentPermission, error) {
			assert.Equal(t, int8(1), contentType)
			assert.Equal(t, uint64(1), contentID)
			roleID := uint(9)
			return []*entity.ContentPermission{{RoleID: &roleID}, {RoleID: nil}}, nil
		},
		SetContentPermsFn: func(_ context.Context, contentType int8, contentID uint64, roleIDs []uint) error {
			permSet = true
			assert.Equal(t, int8(1), contentType)
			assert.Equal(t, uint64(101), contentID)
			assert.Equal(t, []uint{9}, roleIDs)
			return nil
		},
	}
	svc := NewArticleService(repo, permRepo, logrus.New(), attachmentRepo)
	newID, err := svc.Copy(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.Equal(t, uint64(101), newID)
	assert.True(t, attachmentReplaced)
	assert.True(t, permSet)
}

func TestArticleService_Copy_IgnoreAttachmentAndPermErrors(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return &entity.Article{ID: id, Title: "A", ModuleID: 1, ContentType: 1}, nil
		},
		CreateFn: func(_ context.Context, a *entity.Article) error {
			a.ID = 102
			return nil
		},
	}
	attachmentRepo := &articleAttachmentRepoStub{
		listFn: func(_ context.Context, _ uint64) ([]uint64, error) { return nil, apperrors.NewInternal("list", nil) },
	}
	permRepo := &testutil.MockContentPermissionRepository{
		GetByContentFn: func(_ context.Context, _ int8, _ uint64) ([]*entity.ContentPermission, error) {
			return nil, apperrors.NewInternal("perm", nil)
		},
	}
	svc := NewArticleService(repo, permRepo, logrus.New(), attachmentRepo)
	newID, err := svc.Copy(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.Equal(t, uint64(102), newID)
}

func TestCourseService_Pin_Success(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return &entity.Course{ID: id}, nil
		},
		UpdateFn: func(_ context.Context, c *entity.Course) error { return nil },
	}
	svc := newCourseService(repo, nil, &testutil.MockContentPermissionRepository{
		SetContentPermsFn: func(_ context.Context, contentType int8, contentID uint64, roleIDs []uint) error { return nil },
	})
	err := svc.Pin(context.Background(), 1, &dto.PinCourseRequest{SortOrder: 100})
	require.NoError(t, err)
}

func TestCourseService_Pin_NotFoundAndGetError(t *testing.T) {
	repoNotFound := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) { return nil, nil },
	}
	require.Error(t, newCourseService(repoNotFound, nil, nil).Pin(context.Background(), 1, &dto.PinCourseRequest{SortOrder: 1}))

	repoErr := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	require.Error(t, newCourseService(repoErr, nil, nil).Pin(context.Background(), 1, &dto.PinCourseRequest{SortOrder: 1}))
}

func TestCourseService_Copy_Success(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return &entity.Course{ID: id, Title: "C", AuthorID: 1, ModuleID: 1}, nil
		},
		CreateFn: func(_ context.Context, c *entity.Course) error {
			c.ID = 88
			return nil
		},
	}
	unitRepo := &testutil.MockCourseUnitRepository{
		ListByCourseIDFn: func(_ context.Context, courseID uint64) ([]*entity.CourseUnit, error) {
			return []*entity.CourseUnit{{ID: 1, CourseID: courseID, Title: "U1"}}, nil
		},
		CreateFn: func(_ context.Context, unit *entity.CourseUnit) error { return nil },
	}
	permRepo := &testutil.MockContentPermissionRepository{
		GetByContentFn: func(_ context.Context, contentType int8, contentID uint64) ([]*entity.ContentPermission, error) {
			return nil, nil
		},
		SetContentPermsFn: func(_ context.Context, contentType int8, contentID uint64, roleIDs []uint) error { return nil },
	}
	svc := newCourseService(repo, unitRepo, permRepo)
	id, err := svc.Copy(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.Equal(t, uint64(88), id)
}

func TestCourseService_Copy_ErrorBranches(t *testing.T) {
	repoGetErr := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	_, err := newCourseService(repoGetErr, &testutil.MockCourseUnitRepository{}, nil).Copy(context.Background(), 1, 2)
	require.Error(t, err)

	repoNotFound := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) { return nil, nil },
	}
	_, err = newCourseService(repoNotFound, &testutil.MockCourseUnitRepository{}, nil).Copy(context.Background(), 1, 2)
	require.Error(t, err)

	repoCreateErr := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return &entity.Course{ID: id, Title: "C", ModuleID: 1}, nil
		},
		CreateFn: func(_ context.Context, c *entity.Course) error { return apperrors.NewInternal("db", nil) },
	}
	_, err = newCourseService(repoCreateErr, &testutil.MockCourseUnitRepository{}, nil).Copy(context.Background(), 1, 2)
	require.Error(t, err)
}

func TestCourseService_BindAttachmentIDs(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return &entity.Course{ID: id, Title: "C", Status: 1}, nil
		},
	}
	attachmentRepo := &courseAttachmentRepoStub{
		listFn: func(_ context.Context, courseID uint64) ([]uint64, error) {
			return []uint64{21, 22}, nil
		},
	}
	svc := NewCourseService(repo, &testutil.MockCourseUnitRepository{}, nil, logrus.New(), attachmentRepo)
	course, err := svc.AdminGetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, []uint64{21, 22}, course.AttachmentFileIDs)
}

func TestCourseService_Copy_WithAttachmentAndPermissions(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return &entity.Course{ID: id, Title: "C", ModuleID: 1}, nil
		},
		CreateFn: func(_ context.Context, c *entity.Course) error {
			c.ID = 202
			return nil
		},
	}
	unitRepo := &testutil.MockCourseUnitRepository{
		ListByCourseIDFn: func(_ context.Context, courseID uint64) ([]*entity.CourseUnit, error) {
			return []*entity.CourseUnit{{ID: 1, CourseID: courseID, Title: "U"}}, nil
		},
		CreateFn: func(_ context.Context, unit *entity.CourseUnit) error {
			assert.Equal(t, uint64(202), unit.CourseID)
			return nil
		},
	}
	attachmentReplaced := false
	attachmentRepo := &courseAttachmentRepoStub{
		listFn: func(_ context.Context, courseID uint64) ([]uint64, error) {
			assert.Equal(t, uint64(1), courseID)
			return []uint64{6, 7}, nil
		},
		replaceFn: func(_ context.Context, courseID uint64, fileIDs []uint64) error {
			attachmentReplaced = true
			assert.Equal(t, uint64(202), courseID)
			assert.Equal(t, []uint64{6, 7}, fileIDs)
			return nil
		},
	}
	permSet := false
	permRepo := &testutil.MockContentPermissionRepository{
		GetByContentFn: func(_ context.Context, contentType int8, contentID uint64) ([]*entity.ContentPermission, error) {
			assert.Equal(t, int8(2), contentType)
			assert.Equal(t, uint64(1), contentID)
			roleID := uint(3)
			return []*entity.ContentPermission{{RoleID: &roleID}}, nil
		},
		SetContentPermsFn: func(_ context.Context, contentType int8, contentID uint64, roleIDs []uint) error {
			permSet = true
			assert.Equal(t, int8(2), contentType)
			assert.Equal(t, uint64(202), contentID)
			assert.Equal(t, []uint{3}, roleIDs)
			return nil
		},
	}
	svc := NewCourseService(repo, unitRepo, permRepo, logrus.New(), attachmentRepo)
	newID, err := svc.Copy(context.Background(), 1, 9)
	require.NoError(t, err)
	assert.Equal(t, uint64(202), newID)
	assert.True(t, attachmentReplaced)
	assert.True(t, permSet)
}

func TestCourseService_Copy_IgnoreUnitAttachmentPermErrors(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return &entity.Course{ID: id, Title: "C", ModuleID: 1}, nil
		},
		CreateFn: func(_ context.Context, c *entity.Course) error {
			c.ID = 203
			return nil
		},
	}
	unitRepo := &testutil.MockCourseUnitRepository{
		ListByCourseIDFn: func(_ context.Context, courseID uint64) ([]*entity.CourseUnit, error) {
			return nil, apperrors.NewInternal("unit", nil)
		},
	}
	attachmentRepo := &courseAttachmentRepoStub{
		listFn: func(_ context.Context, courseID uint64) ([]uint64, error) {
			return nil, apperrors.NewInternal("att", nil)
		},
	}
	permRepo := &testutil.MockContentPermissionRepository{
		GetByContentFn: func(_ context.Context, _ int8, _ uint64) ([]*entity.ContentPermission, error) {
			return nil, apperrors.NewInternal("perm", nil)
		},
	}
	svc := NewCourseService(repo, unitRepo, permRepo, logrus.New(), attachmentRepo)
	newID, err := svc.Copy(context.Background(), 1, 9)
	require.NoError(t, err)
	assert.Equal(t, uint64(203), newID)
}

// ==================== Missing module/article/course error paths ====================

func TestModuleService_Create_Error(t *testing.T) {
	repo := &testutil.MockModuleRepository{
		CreateFn: func(_ context.Context, m *entity.Module) error {
			return apperrors.NewInternal("db", nil)
		},
	}
	svc := newModuleService(repo, nil)
	_, err := svc.Create(context.Background(), &dto.CreateModuleRequest{Title: "T"})
	require.Error(t, err)
}

func TestModuleService_Delete_DBError(t *testing.T) {
	repo := &testutil.MockModuleRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Module, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := newModuleService(repo, nil)
	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
}

func TestModuleService_CreatePage_Error(t *testing.T) {
	pageRepo := &testutil.MockModulePageRepository{
		CreateFn: func(_ context.Context, p *entity.ModulePage) error {
			return apperrors.NewInternal("db", nil)
		},
	}
	svc := newModuleService(&testutil.MockModuleRepository{}, pageRepo)
	_, err := svc.CreatePage(context.Background(), 1, &dto.CreateModulePageRequest{Title: "T"})
	require.Error(t, err)
}

func TestModuleService_DeletePage_DBError(t *testing.T) {
	pageRepo := &testutil.MockModulePageRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.ModulePage, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := newModuleService(&testutil.MockModuleRepository{}, pageRepo)
	err := svc.DeletePage(context.Background(), 1, 1)
	require.Error(t, err)
}

func TestArticleService_Create_PermFail(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		CreateFn: func(_ context.Context, a *entity.Article) error {
			a.ID = 1
			return nil
		},
	}
	permRepo := &testutil.MockContentPermissionRepository{
		SetContentPermsFn: func(_ context.Context, contentType int8, contentID uint64, roleIDs []uint) error {
			return apperrors.NewInternal("perm db", nil)
		},
	}
	svc := newArticleService(repo, permRepo)
	// Permission failure should only warn, not fail Create
	id, err := svc.Create(context.Background(), &dto.CreateArticleRequest{
		Title: "Test", RolePermissions: []uint{1},
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), id)
}

func TestArticleService_AdminGetByID_DBError(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := newArticleService(repo, nil)
	_, err := svc.AdminGetByID(context.Background(), 1)
	require.Error(t, err)
}

func TestArticleService_Update_GetByIDError(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := newArticleService(repo, nil)
	err := svc.Update(context.Background(), 1, &dto.UpdateArticleRequest{Title: "T"})
	require.Error(t, err)
}

func TestArticleService_Delete_DBError(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := newArticleService(repo, nil)
	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
}

func TestArticleService_Publish_DBError(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := newArticleService(repo, nil)
	err := svc.Publish(context.Background(), 1, &dto.PublishArticleRequest{Status: 1})
	require.Error(t, err)
}

func TestCourseService_Create_PermFail(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		CreateFn: func(_ context.Context, c *entity.Course) error {
			c.ID = 1
			return nil
		},
	}
	permRepo := &testutil.MockContentPermissionRepository{
		SetContentPermsFn: func(_ context.Context, contentType int8, contentID uint64, roleIDs []uint) error {
			return apperrors.NewInternal("perm db", nil)
		},
	}
	svc := newCourseService(repo, nil, permRepo)
	// Permission failure should only warn, not fail Create
	id, err := svc.Create(context.Background(), &dto.CreateCourseRequest{
		Title: "Test", RolePermissions: []uint{1},
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), id)
}

func TestCourseService_AdminGetByID_DBError(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := newCourseService(repo, nil, nil)
	_, err := svc.AdminGetByID(context.Background(), 1)
	require.Error(t, err)
}

func TestCourseService_Update_GetByIDError(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := newCourseService(repo, nil, nil)
	err := svc.Update(context.Background(), 1, &dto.UpdateCourseRequest{Title: "T"})
	require.Error(t, err)
}

func TestCourseService_Delete_DBError(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := newCourseService(repo, nil, nil)
	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
}

func TestCourseService_Publish_DBError(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := newCourseService(repo, nil, nil)
	err := svc.Publish(context.Background(), 1, &dto.PublishCourseRequest{Status: 1})
	require.Error(t, err)
}

func TestCourseService_CreateUnit_Error(t *testing.T) {
	unitRepo := &testutil.MockCourseUnitRepository{
		CreateFn: func(_ context.Context, u *entity.CourseUnit) error {
			return apperrors.NewInternal("db", nil)
		},
	}
	svc := newCourseService(nil, unitRepo, nil)
	_, err := svc.CreateUnit(context.Background(), 1, &dto.CreateCourseUnitRequest{Title: "T"})
	require.Error(t, err)
}

func TestCourseService_DeleteUnit_DBError(t *testing.T) {
	unitRepo := &testutil.MockCourseUnitRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.CourseUnit, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := newCourseService(nil, unitRepo, nil)
	err := svc.DeleteUnit(context.Background(), 1, 1)
	require.Error(t, err)
}

func TestArticleService_GetByID_DefaultPublicWithoutPermissions(t *testing.T) {
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return &entity.Article{ID: id, Status: 1}, nil
		},
		IncrViewCountFn: func(_ context.Context, id uint64) error { return nil },
	}
	permRepo := &testutil.MockContentPermissionRepository{
		GetByContentFn: func(_ context.Context, contentType int8, contentID uint64) ([]*entity.ContentPermission, error) {
			return nil, nil
		},
	}
	svc := NewArticleService(repo, permRepo, logrus.New(), &testutil.MockRoleRepository{})
	uid := uint64(100)
	_, err := svc.GetByID(context.Background(), 1, &uid)
	require.NoError(t, err)
}

func TestCourseService_GetByID_DeniedWhenRoleNotMatched(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return &entity.Course{ID: id, Status: 1}, nil
		},
		IncrViewCountFn: func(_ context.Context, id uint64) error { return nil },
	}
	permRepo := &testutil.MockContentPermissionRepository{
		GetByContentFn: func(_ context.Context, contentType int8, contentID uint64) ([]*entity.ContentPermission, error) {
			roleID := uint(2)
			return []*entity.ContentPermission{{RoleID: &roleID}}, nil
		},
	}
	roleRepo := &testutil.MockRoleRepository{
		GetUserRolesFn: func(_ context.Context, userID uint64) ([]*entity.Role, error) {
			return []*entity.Role{{ID: 1}}, nil
		},
	}
	svc := NewCourseService(repo, &testutil.MockCourseUnitRepository{}, permRepo, logrus.New(), roleRepo)
	uid := uint64(100)
	_, err := svc.GetByID(context.Background(), 1, &uid)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "无权限")
}

func TestCourseService_GetByID_AllowsByRoleHierarchy(t *testing.T) {
	repo := &testutil.MockCourseRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return &entity.Course{ID: id, Status: 1}, nil
		},
		IncrViewCountFn: func(_ context.Context, id uint64) error { return nil },
	}
	permRepo := &testutil.MockContentPermissionRepository{
		GetByContentFn: func(_ context.Context, contentType int8, contentID uint64) ([]*entity.ContentPermission, error) {
			roleID := uint(1)
			return []*entity.ContentPermission{{RoleID: &roleID}}, nil
		},
	}
	roleRepo := &testutil.MockRoleRepository{
		GetUserRolesFn: func(_ context.Context, userID uint64) ([]*entity.Role, error) {
			return []*entity.Role{{ID: 2, ParentID: 1}}, nil
		},
		ListFn: func(_ context.Context) ([]*entity.Role, error) {
			return []*entity.Role{{ID: 1, ParentID: 0}, {ID: 2, ParentID: 1}}, nil
		},
	}
	svc := NewCourseService(repo, &testutil.MockCourseUnitRepository{}, permRepo, logrus.New(), roleRepo)
	uid := uint64(100)
	_, err := svc.GetByID(context.Background(), 1, &uid)
	require.NoError(t, err)
}
