package controller

import (
	"context"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/testutil"
)

// ── ArticleController ─────────────────────────────────────────────────────────

func artCtrl(svc *testutil.MockArticleService) *ArticleController {
	return NewArticleController(svc, logrus.New())
}

func TestArticleCtrl_List_OK(t *testing.T) {
	svc := &testutil.MockArticleService{
		ListFn: func(_ context.Context, p, ps int, kw string, mid *uint, sort string, uid *uint64) ([]*entity.Article, int64, error) {
			return []*entity.Article{{ID: 1}}, 1, nil
		},
	}
	r := newTestRouter()
	r.GET("/articles", artCtrl(svc).List)
	assert.Equal(t, 200, doRequest(r, "GET", "/articles?module_id=1", "").Code)
}

func TestArticleCtrl_List_SvcErr(t *testing.T) {
	svc := &testutil.MockArticleService{
		ListFn: func(_ context.Context, p, ps int, kw string, mid *uint, sort string, uid *uint64) ([]*entity.Article, int64, error) {
			return nil, 0, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouter()
	r.GET("/articles", artCtrl(svc).List)
	assert.Equal(t, 500, doRequest(r, "GET", "/articles", "").Code)
}

func TestArticleCtrl_GetByID_OK(t *testing.T) {
	svc := &testutil.MockArticleService{
		GetByIDFn: func(_ context.Context, id uint64, uid *uint64) (*entity.Article, error) {
			return &entity.Article{ID: id}, nil
		},
	}
	r := newTestRouterWithAuth(1, 1)
	r.GET("/articles/:id", artCtrl(svc).GetByID)
	assert.Equal(t, 200, doRequest(r, "GET", "/articles/1", "").Code)
}

func TestArticleCtrl_GetByID_BadID(t *testing.T) {
	r := newTestRouter()
	r.GET("/articles/:id", artCtrl(&testutil.MockArticleService{}).GetByID)
	assert.Equal(t, 400, doRequest(r, "GET", "/articles/abc", "").Code)
}

func TestArticleCtrl_GetByID_SvcErr(t *testing.T) {
	svc := &testutil.MockArticleService{
		GetByIDFn: func(_ context.Context, id uint64, uid *uint64) (*entity.Article, error) {
			return nil, apperrors.NewNotFound("nf", nil)
		},
	}
	r := newTestRouter()
	r.GET("/articles/:id", artCtrl(svc).GetByID)
	assert.Equal(t, 404, doRequest(r, "GET", "/articles/1", "").Code)
}

func TestArticleCtrl_AdminList_OK(t *testing.T) {
	svc := &testutil.MockArticleService{
		AdminListFn: func(_ context.Context, p, ps int, kw string, mid *uint, st *int8) ([]*entity.Article, int64, error) {
			return []*entity.Article{{ID: 1}}, 1, nil
		},
	}
	r := newTestRouter()
	r.GET("/admin/articles", artCtrl(svc).AdminList)
	assert.Equal(t, 200, doRequest(r, "GET", "/admin/articles?module_id=1&status=1", "").Code)
}

func TestArticleCtrl_AdminList_SvcErr(t *testing.T) {
	svc := &testutil.MockArticleService{
		AdminListFn: func(_ context.Context, p, ps int, kw string, mid *uint, st *int8) ([]*entity.Article, int64, error) {
			return nil, 0, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouter()
	r.GET("/admin/articles", artCtrl(svc).AdminList)
	assert.Equal(t, 500, doRequest(r, "GET", "/admin/articles", "").Code)
}

func TestArticleCtrl_AdminGetByID_OK(t *testing.T) {
	svc := &testutil.MockArticleService{
		AdminGetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return &entity.Article{ID: id}, nil
		},
	}
	r := newTestRouter()
	r.GET("/admin/articles/:id", artCtrl(svc).AdminGetByID)
	assert.Equal(t, 200, doRequest(r, "GET", "/admin/articles/1", "").Code)
}

func TestArticleCtrl_AdminGetByID_BadID(t *testing.T) {
	r := newTestRouter()
	r.GET("/admin/articles/:id", artCtrl(&testutil.MockArticleService{}).AdminGetByID)
	assert.Equal(t, 400, doRequest(r, "GET", "/admin/articles/abc", "").Code)
}

func TestArticleCtrl_AdminGetByID_SvcErr(t *testing.T) {
	svc := &testutil.MockArticleService{
		AdminGetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return nil, apperrors.NewNotFound("nf", nil)
		},
	}
	r := newTestRouter()
	r.GET("/admin/articles/:id", artCtrl(svc).AdminGetByID)
	assert.Equal(t, 404, doRequest(r, "GET", "/admin/articles/1", "").Code)
}

func TestArticleCtrl_AdminCreate_OK(t *testing.T) {
	svc := &testutil.MockArticleService{
		CreateFn: func(_ context.Context, req *dto.CreateArticleRequest, authorID uint64) (uint64, error) {
			return 1, nil
		},
	}
	r := newTestRouterWithAuth(1, 2)
	r.POST("/admin/articles", artCtrl(svc).AdminCreate)
	body := `{"title":"T","content":"C","module_id":1,"summary":"S"}`
	assert.Equal(t, 201, doRequest(r, "POST", "/admin/articles", body).Code)
}

func TestArticleCtrl_AdminCreate_BadJSON(t *testing.T) {
	r := newTestRouterWithAuth(1, 2)
	r.POST("/admin/articles", artCtrl(&testutil.MockArticleService{}).AdminCreate)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/articles", `bad`).Code)
}

func TestArticleCtrl_AdminCreate_Validation(t *testing.T) {
	r := newTestRouterWithAuth(1, 2)
	r.POST("/admin/articles", artCtrl(&testutil.MockArticleService{}).AdminCreate)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/articles", `{"title":"","content":""}`).Code)
}

func TestArticleCtrl_AdminCreate_SvcErr(t *testing.T) {
	svc := &testutil.MockArticleService{
		CreateFn: func(_ context.Context, req *dto.CreateArticleRequest, authorID uint64) (uint64, error) {
			return 0, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouterWithAuth(1, 2)
	r.POST("/admin/articles", artCtrl(svc).AdminCreate)
	assert.Equal(t, 500, doRequest(r, "POST", "/admin/articles", `{"title":"T","content":"C","module_id":1}`).Code)
}

func TestArticleCtrl_AdminUpdate_OK(t *testing.T) {
	svc := &testutil.MockArticleService{
		UpdateFn: func(_ context.Context, id uint64, req *dto.UpdateArticleRequest) error { return nil },
	}
	r := newTestRouter()
	r.PUT("/admin/articles/:id", artCtrl(svc).AdminUpdate)
	assert.Equal(t, 200, doRequest(r, "PUT", "/admin/articles/1", `{"title":"T","content":"C"}`).Code)
}

func TestArticleCtrl_AdminUpdate_BadID(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/articles/:id", artCtrl(&testutil.MockArticleService{}).AdminUpdate)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/articles/abc", `{}`).Code)
}

func TestArticleCtrl_AdminUpdate_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/articles/:id", artCtrl(&testutil.MockArticleService{}).AdminUpdate)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/articles/1", `bad`).Code)
}

func TestArticleCtrl_AdminUpdate_Validation(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/articles/:id", artCtrl(&testutil.MockArticleService{}).AdminUpdate)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/articles/1", `{"title":"","content":""}`).Code)
}

func TestArticleCtrl_AdminUpdate_SvcErr(t *testing.T) {
	svc := &testutil.MockArticleService{
		UpdateFn: func(_ context.Context, id uint64, req *dto.UpdateArticleRequest) error {
			return apperrors.NewNotFound("nf", nil)
		},
	}
	r := newTestRouter()
	r.PUT("/admin/articles/:id", artCtrl(svc).AdminUpdate)
	assert.Equal(t, 404, doRequest(r, "PUT", "/admin/articles/1", `{"title":"T","content":"C"}`).Code)
}

func TestArticleCtrl_AdminDelete_OK(t *testing.T) {
	svc := &testutil.MockArticleService{
		DeleteFn: func(_ context.Context, id uint64) error { return nil },
	}
	r := newTestRouter()
	r.DELETE("/admin/articles/:id", artCtrl(svc).AdminDelete)
	assert.Equal(t, http.StatusNoContent, doRequest(r, "DELETE", "/admin/articles/1", "").Code)
}

func TestArticleCtrl_AdminDelete_BadID(t *testing.T) {
	r := newTestRouter()
	r.DELETE("/admin/articles/:id", artCtrl(&testutil.MockArticleService{}).AdminDelete)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/admin/articles/abc", "").Code)
}

func TestArticleCtrl_AdminDelete_SvcErr(t *testing.T) {
	svc := &testutil.MockArticleService{
		DeleteFn: func(_ context.Context, id uint64) error { return apperrors.NewNotFound("nf", nil) },
	}
	r := newTestRouter()
	r.DELETE("/admin/articles/:id", artCtrl(svc).AdminDelete)
	assert.Equal(t, 404, doRequest(r, "DELETE", "/admin/articles/1", "").Code)
}

func TestArticleCtrl_AdminPublish_OK(t *testing.T) {
	svc := &testutil.MockArticleService{
		PublishFn: func(_ context.Context, id uint64, req *dto.PublishArticleRequest) error { return nil },
	}
	r := newTestRouter()
	r.POST("/admin/articles/:id/publish", artCtrl(svc).AdminPublish)
	assert.Equal(t, 200, doRequest(r, "POST", "/admin/articles/1/publish", `{"status":1}`).Code)
}

func TestArticleCtrl_AdminPublish_BadID(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/articles/:id/publish", artCtrl(&testutil.MockArticleService{}).AdminPublish)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/articles/abc/publish", `{"status":1}`).Code)
}

func TestArticleCtrl_AdminPublish_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/articles/:id/publish", artCtrl(&testutil.MockArticleService{}).AdminPublish)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/articles/1/publish", `bad`).Code)
}

func TestArticleCtrl_AdminPublish_Validation(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/articles/:id/publish", artCtrl(&testutil.MockArticleService{}).AdminPublish)
	// status field not present / invalid
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/articles/1/publish", `{"status":99}`).Code)
}

func TestArticleCtrl_AdminPublish_SvcErr(t *testing.T) {
	svc := &testutil.MockArticleService{
		PublishFn: func(_ context.Context, id uint64, req *dto.PublishArticleRequest) error {
			return apperrors.NewNotFound("nf", nil)
		},
	}
	r := newTestRouter()
	r.POST("/admin/articles/:id/publish", artCtrl(svc).AdminPublish)
	assert.Equal(t, 404, doRequest(r, "POST", "/admin/articles/1/publish", `{"status":1}`).Code)
}

// ── ModuleController ──────────────────────────────────────────────────────────

func TestModuleCtrl_List_OK(t *testing.T) {
	svc := &testutil.MockModuleService{
		ListFn: func(_ context.Context, st *int8) ([]*entity.Module, error) {
			return []*entity.Module{{ID: 1}}, nil
		},
	}
	r := newTestRouter()
	r.GET("/modules", NewModuleController(svc, logrus.New()).List)
	assert.Equal(t, 200, doRequest(r, "GET", "/modules?status=1", "").Code)
}

func TestModuleCtrl_List_SvcErr(t *testing.T) {
	svc := &testutil.MockModuleService{
		ListFn: func(_ context.Context, st *int8) ([]*entity.Module, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouter()
	r.GET("/modules", NewModuleController(svc, logrus.New()).List)
	assert.Equal(t, 500, doRequest(r, "GET", "/modules", "").Code)
}

func TestModuleCtrl_Create_OK(t *testing.T) {
	svc := &testutil.MockModuleService{
		CreateFn: func(_ context.Context, req *dto.CreateModuleRequest) (uint, error) { return 1, nil },
	}
	r := newTestRouter()
	r.POST("/admin/modules", NewModuleController(svc, logrus.New()).Create)
	assert.Equal(t, 201, doRequest(r, "POST", "/admin/modules", `{"title":"M","sort_order":1}`).Code)
}

func TestModuleCtrl_Create_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/modules", NewModuleController(&testutil.MockModuleService{}, logrus.New()).Create)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/modules", `bad`).Code)
}

func TestModuleCtrl_Create_Validation(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/modules", NewModuleController(&testutil.MockModuleService{}, logrus.New()).Create)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/modules", `{"title":""}`).Code)
}

func TestModuleCtrl_Create_SvcErr(t *testing.T) {
	svc := &testutil.MockModuleService{
		CreateFn: func(_ context.Context, req *dto.CreateModuleRequest) (uint, error) {
			return 0, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouter()
	r.POST("/admin/modules", NewModuleController(svc, logrus.New()).Create)
	assert.Equal(t, 500, doRequest(r, "POST", "/admin/modules", `{"title":"M","sort_order":1}`).Code)
}

func TestModuleCtrl_Update_OK(t *testing.T) {
	svc := &testutil.MockModuleService{
		UpdateFn: func(_ context.Context, id uint, req *dto.CreateModuleRequest) error { return nil },
	}
	r := newTestRouter()
	r.PUT("/admin/modules/:id", NewModuleController(svc, logrus.New()).Update)
	assert.Equal(t, 200, doRequest(r, "PUT", "/admin/modules/1", `{"title":"M","sort_order":1}`).Code)
}

func TestModuleCtrl_Update_BadID(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/modules/:id", NewModuleController(&testutil.MockModuleService{}, logrus.New()).Update)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/modules/abc", `{}`).Code)
}

func TestModuleCtrl_Update_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/modules/:id", NewModuleController(&testutil.MockModuleService{}, logrus.New()).Update)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/modules/1", `bad`).Code)
}

func TestModuleCtrl_Update_Validation(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/modules/:id", NewModuleController(&testutil.MockModuleService{}, logrus.New()).Update)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/modules/1", `{"title":""}`).Code)
}

func TestModuleCtrl_Update_SvcErr(t *testing.T) {
	svc := &testutil.MockModuleService{
		UpdateFn: func(_ context.Context, id uint, req *dto.CreateModuleRequest) error {
			return apperrors.NewNotFound("nf", nil)
		},
	}
	r := newTestRouter()
	r.PUT("/admin/modules/:id", NewModuleController(svc, logrus.New()).Update)
	assert.Equal(t, 404, doRequest(r, "PUT", "/admin/modules/1", `{"title":"M","sort_order":1}`).Code)
}

func TestModuleCtrl_Delete_OK(t *testing.T) {
	svc := &testutil.MockModuleService{
		DeleteFn: func(_ context.Context, id uint) error { return nil },
	}
	r := newTestRouter()
	r.DELETE("/admin/modules/:id", NewModuleController(svc, logrus.New()).Delete)
	assert.Equal(t, http.StatusNoContent, doRequest(r, "DELETE", "/admin/modules/1", "").Code)
}

func TestModuleCtrl_Delete_BadID(t *testing.T) {
	r := newTestRouter()
	r.DELETE("/admin/modules/:id", NewModuleController(&testutil.MockModuleService{}, logrus.New()).Delete)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/admin/modules/abc", "").Code)
}

func TestModuleCtrl_Delete_SvcErr(t *testing.T) {
	svc := &testutil.MockModuleService{
		DeleteFn: func(_ context.Context, id uint) error { return apperrors.NewNotFound("nf", nil) },
	}
	r := newTestRouter()
	r.DELETE("/admin/modules/:id", NewModuleController(svc, logrus.New()).Delete)
	assert.Equal(t, 404, doRequest(r, "DELETE", "/admin/modules/1", "").Code)
}

func TestModuleCtrl_GetPages_OK(t *testing.T) {
	svc := &testutil.MockModuleService{
		GetPagesFn: func(_ context.Context, mid uint) ([]*entity.ModulePage, error) {
			return []*entity.ModulePage{{ID: 1}}, nil
		},
	}
	r := newTestRouter()
	r.GET("/admin/modules/:id/pages", NewModuleController(svc, logrus.New()).GetPages)
	assert.Equal(t, 200, doRequest(r, "GET", "/admin/modules/1/pages", "").Code)
}

func TestModuleCtrl_GetPages_BadID(t *testing.T) {
	r := newTestRouter()
	r.GET("/admin/modules/:id/pages", NewModuleController(&testutil.MockModuleService{}, logrus.New()).GetPages)
	assert.Equal(t, 400, doRequest(r, "GET", "/admin/modules/abc/pages", "").Code)
}

func TestModuleCtrl_GetPages_SvcErr(t *testing.T) {
	svc := &testutil.MockModuleService{
		GetPagesFn: func(_ context.Context, mid uint) ([]*entity.ModulePage, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouter()
	r.GET("/admin/modules/:id/pages", NewModuleController(svc, logrus.New()).GetPages)
	assert.Equal(t, 500, doRequest(r, "GET", "/admin/modules/1/pages", "").Code)
}

func TestModuleCtrl_CreatePage_OK(t *testing.T) {
	svc := &testutil.MockModuleService{
		CreatePageFn: func(_ context.Context, mid uint, req *dto.CreateModulePageRequest) (uint, error) {
			return 1, nil
		},
	}
	r := newTestRouter()
	r.POST("/admin/modules/:id/pages", NewModuleController(svc, logrus.New()).CreatePage)
	assert.Equal(t, 201, doRequest(r, "POST", "/admin/modules/1/pages", `{"title":"P","content":"C","sort_order":1}`).Code)
}

func TestModuleCtrl_CreatePage_BadID(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/modules/:id/pages", NewModuleController(&testutil.MockModuleService{}, logrus.New()).CreatePage)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/modules/abc/pages", `{}`).Code)
}

func TestModuleCtrl_CreatePage_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/modules/:id/pages", NewModuleController(&testutil.MockModuleService{}, logrus.New()).CreatePage)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/modules/1/pages", `bad`).Code)
}

func TestModuleCtrl_CreatePage_Validation(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/modules/:id/pages", NewModuleController(&testutil.MockModuleService{}, logrus.New()).CreatePage)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/modules/1/pages", `{"title":""}`).Code)
}

func TestModuleCtrl_CreatePage_SvcErr(t *testing.T) {
	svc := &testutil.MockModuleService{
		CreatePageFn: func(_ context.Context, mid uint, req *dto.CreateModulePageRequest) (uint, error) {
			return 0, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouter()
	r.POST("/admin/modules/:id/pages", NewModuleController(svc, logrus.New()).CreatePage)
	assert.Equal(t, 500, doRequest(r, "POST", "/admin/modules/1/pages", `{"title":"P","content":"C","sort_order":1}`).Code)
}

func TestModuleCtrl_UpdatePage_OK(t *testing.T) {
	svc := &testutil.MockModuleService{
		UpdatePageFn: func(_ context.Context, mid, pid uint, req *dto.CreateModulePageRequest) error { return nil },
	}
	r := newTestRouter()
	r.PUT("/admin/modules/:id/pages/:page_id", NewModuleController(svc, logrus.New()).UpdatePage)
	assert.Equal(t, 200, doRequest(r, "PUT", "/admin/modules/1/pages/2", `{"title":"P","content":"C","sort_order":1}`).Code)
}

func TestModuleCtrl_UpdatePage_BadModuleID(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/modules/:id/pages/:page_id", NewModuleController(&testutil.MockModuleService{}, logrus.New()).UpdatePage)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/modules/abc/pages/1", `{}`).Code)
}

func TestModuleCtrl_UpdatePage_BadPageID(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/modules/:id/pages/:page_id", NewModuleController(&testutil.MockModuleService{}, logrus.New()).UpdatePage)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/modules/1/pages/abc", `{}`).Code)
}

func TestModuleCtrl_UpdatePage_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/modules/:id/pages/:page_id", NewModuleController(&testutil.MockModuleService{}, logrus.New()).UpdatePage)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/modules/1/pages/2", `bad`).Code)
}

func TestModuleCtrl_UpdatePage_Validation(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/modules/:id/pages/:page_id", NewModuleController(&testutil.MockModuleService{}, logrus.New()).UpdatePage)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/modules/1/pages/2", `{"title":""}`).Code)
}

func TestModuleCtrl_UpdatePage_SvcErr(t *testing.T) {
	svc := &testutil.MockModuleService{
		UpdatePageFn: func(_ context.Context, mid, pid uint, req *dto.CreateModulePageRequest) error {
			return apperrors.NewNotFound("nf", nil)
		},
	}
	r := newTestRouter()
	r.PUT("/admin/modules/:id/pages/:page_id", NewModuleController(svc, logrus.New()).UpdatePage)
	assert.Equal(t, 404, doRequest(r, "PUT", "/admin/modules/1/pages/2", `{"title":"P","content":"C","sort_order":1}`).Code)
}

func TestModuleCtrl_DeletePage_OK(t *testing.T) {
	svc := &testutil.MockModuleService{
		DeletePageFn: func(_ context.Context, mid, pid uint) error { return nil },
	}
	r := newTestRouter()
	r.DELETE("/admin/modules/:id/pages/:page_id", NewModuleController(svc, logrus.New()).DeletePage)
	assert.Equal(t, http.StatusNoContent, doRequest(r, "DELETE", "/admin/modules/1/pages/2", "").Code)
}

func TestModuleCtrl_DeletePage_BadModuleID(t *testing.T) {
	r := newTestRouter()
	r.DELETE("/admin/modules/:id/pages/:page_id", NewModuleController(&testutil.MockModuleService{}, logrus.New()).DeletePage)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/admin/modules/abc/pages/1", "").Code)
}

func TestModuleCtrl_DeletePage_BadPageID(t *testing.T) {
	r := newTestRouter()
	r.DELETE("/admin/modules/:id/pages/:page_id", NewModuleController(&testutil.MockModuleService{}, logrus.New()).DeletePage)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/admin/modules/1/pages/abc", "").Code)
}

func TestModuleCtrl_DeletePage_SvcErr(t *testing.T) {
	svc := &testutil.MockModuleService{
		DeletePageFn: func(_ context.Context, mid, pid uint) error { return apperrors.NewNotFound("nf", nil) },
	}
	r := newTestRouter()
	r.DELETE("/admin/modules/:id/pages/:page_id", NewModuleController(svc, logrus.New()).DeletePage)
	assert.Equal(t, 404, doRequest(r, "DELETE", "/admin/modules/1/pages/2", "").Code)
}
