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

// ── CourseController ──────────────────────────────────────────────────────────

func crsCtrl(svc *testutil.MockCourseService) *CourseController {
	return NewCourseController(svc, logrus.New())
}

func TestCourseCtrl_List_OK(t *testing.T) {
	svc := &testutil.MockCourseService{
		ListFn: func(_ context.Context, p, ps int, kw string, mid *uint, isFree *bool, uid *uint64) ([]*entity.Course, int64, error) {
			return []*entity.Course{{ID: 1}}, 1, nil
		},
	}
	r := newTestRouter()
	r.GET("/courses", crsCtrl(svc).List)
	assert.Equal(t, 200, doRequest(r, "GET", "/courses?module_id=1&is_free=true", "").Code)
}

func TestCourseCtrl_List_SvcErr(t *testing.T) {
	svc := &testutil.MockCourseService{
		ListFn: func(_ context.Context, p, ps int, kw string, mid *uint, isFree *bool, uid *uint64) ([]*entity.Course, int64, error) {
			return nil, 0, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouter()
	r.GET("/courses", crsCtrl(svc).List)
	assert.Equal(t, 500, doRequest(r, "GET", "/courses", "").Code)
}

func TestCourseCtrl_GetByID_OK(t *testing.T) {
	svc := &testutil.MockCourseService{
		GetByIDFn: func(_ context.Context, id uint64, uid *uint64) (*entity.Course, error) {
			return &entity.Course{ID: id}, nil
		},
	}
	r := newTestRouterWithAuth(1, 1)
	r.GET("/courses/:id", crsCtrl(svc).GetByID)
	assert.Equal(t, 200, doRequest(r, "GET", "/courses/1", "").Code)
}

func TestCourseCtrl_GetByID_BadID(t *testing.T) {
	r := newTestRouter()
	r.GET("/courses/:id", crsCtrl(&testutil.MockCourseService{}).GetByID)
	assert.Equal(t, 400, doRequest(r, "GET", "/courses/abc", "").Code)
}

func TestCourseCtrl_GetByID_SvcErr(t *testing.T) {
	svc := &testutil.MockCourseService{
		GetByIDFn: func(_ context.Context, id uint64, uid *uint64) (*entity.Course, error) {
			return nil, apperrors.NewNotFound("nf", nil)
		},
	}
	r := newTestRouter()
	r.GET("/courses/:id", crsCtrl(svc).GetByID)
	assert.Equal(t, 404, doRequest(r, "GET", "/courses/1", "").Code)
}

func TestCourseCtrl_AdminList_OK(t *testing.T) {
	svc := &testutil.MockCourseService{
		AdminListFn: func(_ context.Context, p, ps int, kw string, st *int8) ([]*entity.Course, int64, error) {
			return []*entity.Course{{ID: 1}}, 1, nil
		},
	}
	r := newTestRouter()
	r.GET("/admin/courses", crsCtrl(svc).AdminList)
	assert.Equal(t, 200, doRequest(r, "GET", "/admin/courses?status=1", "").Code)
}

func TestCourseCtrl_AdminList_SvcErr(t *testing.T) {
	svc := &testutil.MockCourseService{
		AdminListFn: func(_ context.Context, p, ps int, kw string, st *int8) ([]*entity.Course, int64, error) {
			return nil, 0, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouter()
	r.GET("/admin/courses", crsCtrl(svc).AdminList)
	assert.Equal(t, 500, doRequest(r, "GET", "/admin/courses", "").Code)
}

func TestCourseCtrl_AdminGetByID_OK(t *testing.T) {
	svc := &testutil.MockCourseService{
		AdminGetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return &entity.Course{ID: id}, nil
		},
	}
	r := newTestRouter()
	r.GET("/admin/courses/:id", crsCtrl(svc).AdminGetByID)
	assert.Equal(t, 200, doRequest(r, "GET", "/admin/courses/1", "").Code)
}

func TestCourseCtrl_AdminGetByID_BadID(t *testing.T) {
	r := newTestRouter()
	r.GET("/admin/courses/:id", crsCtrl(&testutil.MockCourseService{}).AdminGetByID)
	assert.Equal(t, 400, doRequest(r, "GET", "/admin/courses/abc", "").Code)
}

func TestCourseCtrl_AdminGetByID_SvcErr(t *testing.T) {
	svc := &testutil.MockCourseService{
		AdminGetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return nil, apperrors.NewNotFound("nf", nil)
		},
	}
	r := newTestRouter()
	r.GET("/admin/courses/:id", crsCtrl(svc).AdminGetByID)
	assert.Equal(t, 404, doRequest(r, "GET", "/admin/courses/1", "").Code)
}

func TestCourseCtrl_AdminCreate_OK(t *testing.T) {
	svc := &testutil.MockCourseService{
		CreateFn: func(_ context.Context, req *dto.CreateCourseRequest, authorID uint64) (uint64, error) {
			return 1, nil
		},
	}
	r := newTestRouterWithAuth(1, 2)
	r.POST("/admin/courses", crsCtrl(svc).AdminCreate)
	assert.Equal(t, 201, doRequest(r, "POST", "/admin/courses", `{"title":"T","price":0}`).Code)
}

func TestCourseCtrl_AdminCreate_BadJSON(t *testing.T) {
	r := newTestRouterWithAuth(1, 2)
	r.POST("/admin/courses", crsCtrl(&testutil.MockCourseService{}).AdminCreate)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/courses", `bad`).Code)
}

func TestCourseCtrl_AdminCreate_Validation(t *testing.T) {
	r := newTestRouterWithAuth(1, 2)
	r.POST("/admin/courses", crsCtrl(&testutil.MockCourseService{}).AdminCreate)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/courses", `{"title":""}`).Code)
}

func TestCourseCtrl_AdminCreate_SvcErr(t *testing.T) {
	svc := &testutil.MockCourseService{
		CreateFn: func(_ context.Context, req *dto.CreateCourseRequest, authorID uint64) (uint64, error) {
			return 0, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouterWithAuth(1, 2)
	r.POST("/admin/courses", crsCtrl(svc).AdminCreate)
	assert.Equal(t, 500, doRequest(r, "POST", "/admin/courses", `{"title":"T","price":0}`).Code)
}

func TestCourseCtrl_AdminUpdate_OK(t *testing.T) {
	svc := &testutil.MockCourseService{
		UpdateFn: func(_ context.Context, id uint64, req *dto.UpdateCourseRequest) error { return nil },
	}
	r := newTestRouter()
	r.PUT("/admin/courses/:id", crsCtrl(svc).AdminUpdate)
	assert.Equal(t, 200, doRequest(r, "PUT", "/admin/courses/1", `{"title":"T","price":0}`).Code)
}

func TestCourseCtrl_AdminUpdate_BadID(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/courses/:id", crsCtrl(&testutil.MockCourseService{}).AdminUpdate)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/courses/abc", `{}`).Code)
}

func TestCourseCtrl_AdminUpdate_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/courses/:id", crsCtrl(&testutil.MockCourseService{}).AdminUpdate)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/courses/1", `bad`).Code)
}

func TestCourseCtrl_AdminUpdate_Validation(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/courses/:id", crsCtrl(&testutil.MockCourseService{}).AdminUpdate)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/courses/1", `{"title":""}`).Code)
}

func TestCourseCtrl_AdminUpdate_SvcErr(t *testing.T) {
	svc := &testutil.MockCourseService{
		UpdateFn: func(_ context.Context, id uint64, req *dto.UpdateCourseRequest) error {
			return apperrors.NewNotFound("nf", nil)
		},
	}
	r := newTestRouter()
	r.PUT("/admin/courses/:id", crsCtrl(svc).AdminUpdate)
	assert.Equal(t, 404, doRequest(r, "PUT", "/admin/courses/1", `{"title":"T","price":0}`).Code)
}

func TestCourseCtrl_AdminDelete_OK(t *testing.T) {
	svc := &testutil.MockCourseService{
		DeleteFn: func(_ context.Context, id uint64) error { return nil },
	}
	r := newTestRouter()
	r.DELETE("/admin/courses/:id", crsCtrl(svc).AdminDelete)
	assert.Equal(t, http.StatusNoContent, doRequest(r, "DELETE", "/admin/courses/1", "").Code)
}

func TestCourseCtrl_AdminDelete_BadID(t *testing.T) {
	r := newTestRouter()
	r.DELETE("/admin/courses/:id", crsCtrl(&testutil.MockCourseService{}).AdminDelete)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/admin/courses/abc", "").Code)
}

func TestCourseCtrl_AdminDelete_SvcErr(t *testing.T) {
	svc := &testutil.MockCourseService{
		DeleteFn: func(_ context.Context, id uint64) error { return apperrors.NewNotFound("nf", nil) },
	}
	r := newTestRouter()
	r.DELETE("/admin/courses/:id", crsCtrl(svc).AdminDelete)
	assert.Equal(t, 404, doRequest(r, "DELETE", "/admin/courses/1", "").Code)
}

func TestCourseCtrl_AdminPublish_OK(t *testing.T) {
	svc := &testutil.MockCourseService{
		PublishFn: func(_ context.Context, id uint64, req *dto.PublishCourseRequest) error { return nil },
	}
	r := newTestRouter()
	r.POST("/admin/courses/:id/publish", crsCtrl(svc).AdminPublish)
	assert.Equal(t, 200, doRequest(r, "POST", "/admin/courses/1/publish", `{"status":1}`).Code)
}

func TestCourseCtrl_AdminPublish_BadID(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/courses/:id/publish", crsCtrl(&testutil.MockCourseService{}).AdminPublish)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/courses/abc/publish", `{"status":1}`).Code)
}

func TestCourseCtrl_AdminPublish_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/courses/:id/publish", crsCtrl(&testutil.MockCourseService{}).AdminPublish)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/courses/1/publish", `bad`).Code)
}

func TestCourseCtrl_AdminPublish_Validation(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/courses/:id/publish", crsCtrl(&testutil.MockCourseService{}).AdminPublish)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/courses/1/publish", `{"status":99}`).Code)
}

func TestCourseCtrl_AdminPublish_SvcErr(t *testing.T) {
	svc := &testutil.MockCourseService{
		PublishFn: func(_ context.Context, id uint64, req *dto.PublishCourseRequest) error {
			return apperrors.NewNotFound("nf", nil)
		},
	}
	r := newTestRouter()
	r.POST("/admin/courses/:id/publish", crsCtrl(svc).AdminPublish)
	assert.Equal(t, 404, doRequest(r, "POST", "/admin/courses/1/publish", `{"status":1}`).Code)
}

func TestCourseCtrl_AdminPin_OK(t *testing.T) {
	svc := &testutil.MockCourseService{
		PinFn: func(_ context.Context, id uint64, req *dto.PinCourseRequest) error { return nil },
	}
	r := newTestRouter()
	r.POST("/admin/courses/:id/pin", crsCtrl(svc).AdminPin)
	assert.Equal(t, 200, doRequest(r, "POST", "/admin/courses/1/pin", `{"sort_order":999}`).Code)
}

func TestCourseCtrl_AdminPin_BadID(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/courses/:id/pin", crsCtrl(&testutil.MockCourseService{}).AdminPin)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/courses/abc/pin", `{"sort_order":1}`).Code)
}

func TestCourseCtrl_AdminPin_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/courses/:id/pin", crsCtrl(&testutil.MockCourseService{}).AdminPin)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/courses/1/pin", `bad`).Code)
}

func TestCourseCtrl_AdminPin_Validation(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/courses/:id/pin", crsCtrl(&testutil.MockCourseService{}).AdminPin)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/courses/1/pin", `{}`).Code)
}

func TestCourseCtrl_AdminPin_SvcErr(t *testing.T) {
	svc := &testutil.MockCourseService{
		PinFn: func(_ context.Context, _ uint64, _ *dto.PinCourseRequest) error {
			return apperrors.NewInternal("pin failed", nil)
		},
	}
	r := newTestRouter()
	r.POST("/admin/courses/:id/pin", crsCtrl(svc).AdminPin)
	assert.Equal(t, 500, doRequest(r, "POST", "/admin/courses/1/pin", `{"sort_order":1}`).Code)
}

func TestCourseCtrl_AdminCopy_OK(t *testing.T) {
	svc := &testutil.MockCourseService{
		CopyFn: func(_ context.Context, id uint64, authorID uint64) (uint64, error) { return 66, nil },
	}
	r := newTestRouter()
	r.POST("/admin/courses/:id/copy", crsCtrl(svc).AdminCopy)
	assert.Equal(t, 201, doRequest(r, "POST", "/admin/courses/1/copy", "").Code)
}

func TestCourseCtrl_AdminCopy_BadID(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/courses/:id/copy", crsCtrl(&testutil.MockCourseService{}).AdminCopy)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/courses/abc/copy", "").Code)
}

func TestCourseCtrl_AdminCopy_SvcErr(t *testing.T) {
	svc := &testutil.MockCourseService{
		CopyFn: func(_ context.Context, _ uint64, _ uint64) (uint64, error) {
			return 0, apperrors.NewInternal("copy failed", nil)
		},
	}
	r := newTestRouterWithAuth(1, 2)
	r.POST("/admin/courses/:id/copy", crsCtrl(svc).AdminCopy)
	assert.Equal(t, 500, doRequest(r, "POST", "/admin/courses/1/copy", "").Code)
}

func TestCourseCtrl_AdminGetUnits_OK(t *testing.T) {
	svc := &testutil.MockCourseService{
		GetUnitsFn: func(_ context.Context, cid uint64) ([]*entity.CourseUnit, error) {
			return []*entity.CourseUnit{{ID: 1}}, nil
		},
	}
	r := newTestRouter()
	r.GET("/admin/courses/:id/units", crsCtrl(svc).AdminGetUnits)
	assert.Equal(t, 200, doRequest(r, "GET", "/admin/courses/1/units", "").Code)
}

func TestCourseCtrl_AdminGetUnits_BadID(t *testing.T) {
	r := newTestRouter()
	r.GET("/admin/courses/:id/units", crsCtrl(&testutil.MockCourseService{}).AdminGetUnits)
	assert.Equal(t, 400, doRequest(r, "GET", "/admin/courses/abc/units", "").Code)
}

func TestCourseCtrl_AdminGetUnits_SvcErr(t *testing.T) {
	svc := &testutil.MockCourseService{
		GetUnitsFn: func(_ context.Context, cid uint64) ([]*entity.CourseUnit, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouter()
	r.GET("/admin/courses/:id/units", crsCtrl(svc).AdminGetUnits)
	assert.Equal(t, 500, doRequest(r, "GET", "/admin/courses/1/units", "").Code)
}

func TestCourseCtrl_AdminCreateUnit_OK(t *testing.T) {
	svc := &testutil.MockCourseService{
		CreateUnitFn: func(_ context.Context, cid uint64, req *dto.CreateCourseUnitRequest) (uint64, error) {
			return 1, nil
		},
	}
	r := newTestRouter()
	r.POST("/admin/courses/:id/units", crsCtrl(svc).AdminCreateUnit)
	assert.Equal(t, 201, doRequest(r, "POST", "/admin/courses/1/units", `{"title":"U"}`).Code)
}

func TestCourseCtrl_AdminCreateUnit_BadID(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/courses/:id/units", crsCtrl(&testutil.MockCourseService{}).AdminCreateUnit)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/courses/abc/units", `{"title":"U"}`).Code)
}

func TestCourseCtrl_AdminCreateUnit_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/courses/:id/units", crsCtrl(&testutil.MockCourseService{}).AdminCreateUnit)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/courses/1/units", `bad`).Code)
}

func TestCourseCtrl_AdminCreateUnit_Validation(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/courses/:id/units", crsCtrl(&testutil.MockCourseService{}).AdminCreateUnit)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/courses/1/units", `{"title":""}`).Code)
}

func TestCourseCtrl_AdminCreateUnit_SvcErr(t *testing.T) {
	svc := &testutil.MockCourseService{
		CreateUnitFn: func(_ context.Context, cid uint64, req *dto.CreateCourseUnitRequest) (uint64, error) {
			return 0, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouter()
	r.POST("/admin/courses/:id/units", crsCtrl(svc).AdminCreateUnit)
	assert.Equal(t, 500, doRequest(r, "POST", "/admin/courses/1/units", `{"title":"U"}`).Code)
}

func TestCourseCtrl_AdminUpdateUnit_OK(t *testing.T) {
	svc := &testutil.MockCourseService{
		UpdateUnitFn: func(_ context.Context, cid, uid uint64, req *dto.CreateCourseUnitRequest) error { return nil },
	}
	r := newTestRouter()
	r.PUT("/admin/courses/:id/units/:unit_id", crsCtrl(svc).AdminUpdateUnit)
	assert.Equal(t, 200, doRequest(r, "PUT", "/admin/courses/1/units/2", `{"title":"U"}`).Code)
}

func TestCourseCtrl_AdminUpdateUnit_BadCourseID(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/courses/:id/units/:unit_id", crsCtrl(&testutil.MockCourseService{}).AdminUpdateUnit)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/courses/abc/units/1", `{}`).Code)
}

func TestCourseCtrl_AdminUpdateUnit_BadUnitID(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/courses/:id/units/:unit_id", crsCtrl(&testutil.MockCourseService{}).AdminUpdateUnit)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/courses/1/units/abc", `{}`).Code)
}

func TestCourseCtrl_AdminUpdateUnit_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/courses/:id/units/:unit_id", crsCtrl(&testutil.MockCourseService{}).AdminUpdateUnit)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/courses/1/units/2", `bad`).Code)
}

func TestCourseCtrl_AdminUpdateUnit_Validation(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/courses/:id/units/:unit_id", crsCtrl(&testutil.MockCourseService{}).AdminUpdateUnit)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/courses/1/units/2", `{"title":""}`).Code)
}

func TestCourseCtrl_AdminUpdateUnit_SvcErr(t *testing.T) {
	svc := &testutil.MockCourseService{
		UpdateUnitFn: func(_ context.Context, cid, uid uint64, req *dto.CreateCourseUnitRequest) error {
			return apperrors.NewNotFound("nf", nil)
		},
	}
	r := newTestRouter()
	r.PUT("/admin/courses/:id/units/:unit_id", crsCtrl(svc).AdminUpdateUnit)
	assert.Equal(t, 404, doRequest(r, "PUT", "/admin/courses/1/units/2", `{"title":"U"}`).Code)
}

func TestCourseCtrl_AdminDeleteUnit_OK(t *testing.T) {
	svc := &testutil.MockCourseService{
		DeleteUnitFn: func(_ context.Context, cid, uid uint64) error { return nil },
	}
	r := newTestRouter()
	r.DELETE("/admin/courses/:id/units/:unit_id", crsCtrl(svc).AdminDeleteUnit)
	assert.Equal(t, http.StatusNoContent, doRequest(r, "DELETE", "/admin/courses/1/units/2", "").Code)
}

func TestCourseCtrl_AdminDeleteUnit_BadCourseID(t *testing.T) {
	r := newTestRouter()
	r.DELETE("/admin/courses/:id/units/:unit_id", crsCtrl(&testutil.MockCourseService{}).AdminDeleteUnit)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/admin/courses/abc/units/1", "").Code)
}

func TestCourseCtrl_AdminDeleteUnit_BadUnitID(t *testing.T) {
	r := newTestRouter()
	r.DELETE("/admin/courses/:id/units/:unit_id", crsCtrl(&testutil.MockCourseService{}).AdminDeleteUnit)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/admin/courses/1/units/abc", "").Code)
}

func TestCourseCtrl_AdminDeleteUnit_SvcErr(t *testing.T) {
	svc := &testutil.MockCourseService{
		DeleteUnitFn: func(_ context.Context, cid, uid uint64) error { return apperrors.NewNotFound("nf", nil) },
	}
	r := newTestRouter()
	r.DELETE("/admin/courses/:id/units/:unit_id", crsCtrl(svc).AdminDeleteUnit)
	assert.Equal(t, 404, doRequest(r, "DELETE", "/admin/courses/1/units/2", "").Code)
}

// ── Additional coverage tests ─────────────────────────────────────────────────

func TestCourseCtrl_List_WithAuth(t *testing.T) {
svc := &testutil.MockCourseService{
ListFn: func(_ context.Context, p, ps int, kw string, mid *uint, isFree *bool, uid *uint64) ([]*entity.Course, int64, error) {
return []*entity.Course{{ID: 1}}, 1, nil
},
}
r := newTestRouterWithAuth(1, 1)
r.GET("/courses", crsCtrl(svc).List)
assert.Equal(t, 200, doRequest(r, "GET", "/courses", "").Code)
}

func TestCourseCtrl_List_BindQueryErr(t *testing.T) {
r := newTestRouter()
r.GET("/courses", crsCtrl(&testutil.MockCourseService{}).List)
assert.Equal(t, 400, doRequest(r, "GET", "/courses?page=abc", "").Code)
}

func TestCourseCtrl_AdminList_BindQueryErr(t *testing.T) {
r := newTestRouter()
r.GET("/admin/courses", crsCtrl(&testutil.MockCourseService{}).AdminList)
assert.Equal(t, 400, doRequest(r, "GET", "/admin/courses?page=abc", "").Code)
}

func TestCourseCtrl_GetByID_WithAuth(t *testing.T) {
svc := &testutil.MockCourseService{
GetByIDFn: func(_ context.Context, id uint64, uid *uint64) (*entity.Course, error) {
return &entity.Course{ID: id}, nil
},
}
r := newTestRouterWithAuth(1, 1)
r.GET("/courses/:id", crsCtrl(svc).GetByID)
assert.Equal(t, 200, doRequest(r, "GET", "/courses/1", "").Code)
}
