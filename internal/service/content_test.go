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
	pageRepo := &testutil.MockModulePageRepository{
		CreateFn: func(_ context.Context, p *entity.ModulePage) error {
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
	repo := &testutil.MockArticleRepository{
		CreateFn: func(_ context.Context, a *entity.Article) error {
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

func TestArticleService_Update_Success(t *testing.T) {
	article := &entity.Article{ID: 1}
	repo := &testutil.MockArticleRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return article, nil
		},
		UpdateFn: func(_ context.Context, a *entity.Article) error {
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
		DeleteFn: func(_ context.Context, id uint64) error {
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
		DeleteFn: func(_ context.Context, id uint64) error {
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
		DeleteFn: func(_ context.Context, id uint64) error {
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
