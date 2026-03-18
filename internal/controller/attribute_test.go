package controller

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/testutil"
)

// ── AttributeController ───────────────────────────────────────────────────────

func attrCtrl(svc *testutil.MockAttributeService) *AttributeController {
	return NewAttributeController(svc, logrus.New())
}

func TestAttrCtrl_List_OK(t *testing.T) {
	svc := &testutil.MockAttributeService{
		ListFn: func(_ context.Context) ([]*entity.Attribute, error) {
			return []*entity.Attribute{{ID: 1, Name: "性别"}}, nil
		},
	}
	r := newTestRouter()
	r.GET("/admin/attributes", attrCtrl(svc).List)
	assert.Equal(t, 200, doRequest(r, "GET", "/admin/attributes", "").Code)
}

func TestAttrCtrl_List_SvcErr(t *testing.T) {
	svc := &testutil.MockAttributeService{
		ListFn: func(_ context.Context) ([]*entity.Attribute, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouter()
	r.GET("/admin/attributes", attrCtrl(svc).List)
	assert.Equal(t, 500, doRequest(r, "GET", "/admin/attributes", "").Code)
}

func TestAttrCtrl_Create_OK(t *testing.T) {
	svc := &testutil.MockAttributeService{
		CreateFn: func(_ context.Context, req *dto.CreateAttributeRequest) (uint, error) { return 1, nil },
	}
	r := newTestRouter()
	r.POST("/admin/attributes", attrCtrl(svc).Create)
	assert.Equal(t, 201, doRequest(r, "POST", "/admin/attributes", `{"name":"性别"}`).Code)
}

func TestAttrCtrl_Create_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/attributes", attrCtrl(&testutil.MockAttributeService{}).Create)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/attributes", `bad`).Code)
}

func TestAttrCtrl_Create_Validation(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/attributes", attrCtrl(&testutil.MockAttributeService{}).Create)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/attributes", `{"name":""}`).Code)
}

func TestAttrCtrl_Create_SvcErr(t *testing.T) {
	svc := &testutil.MockAttributeService{
		CreateFn: func(_ context.Context, req *dto.CreateAttributeRequest) (uint, error) {
			return 0, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouter()
	r.POST("/admin/attributes", attrCtrl(svc).Create)
	assert.Equal(t, 500, doRequest(r, "POST", "/admin/attributes", `{"name":"性别"}`).Code)
}

func TestAttrCtrl_Update_OK(t *testing.T) {
	svc := &testutil.MockAttributeService{
		UpdateFn: func(_ context.Context, id uint, req *dto.UpdateAttributeRequest) error { return nil },
	}
	r := newTestRouter()
	r.PUT("/admin/attributes/:id", attrCtrl(svc).Update)
	assert.Equal(t, 200, doRequest(r, "PUT", "/admin/attributes/1", `{"name":"年龄"}`).Code)
}

func TestAttrCtrl_Update_BadID(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/attributes/:id", attrCtrl(&testutil.MockAttributeService{}).Update)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/attributes/abc", `{"name":"年龄"}`).Code)
}

func TestAttrCtrl_Update_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/attributes/:id", attrCtrl(&testutil.MockAttributeService{}).Update)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/attributes/1", `bad`).Code)
}

func TestAttrCtrl_Update_Validation(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/attributes/:id", attrCtrl(&testutil.MockAttributeService{}).Update)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/attributes/1", `{"name":""}`).Code)
}

func TestAttrCtrl_Update_SvcErr(t *testing.T) {
	svc := &testutil.MockAttributeService{
		UpdateFn: func(_ context.Context, id uint, req *dto.UpdateAttributeRequest) error {
			return apperrors.NewNotFound("not found", nil)
		},
	}
	r := newTestRouter()
	r.PUT("/admin/attributes/:id", attrCtrl(svc).Update)
	assert.Equal(t, 404, doRequest(r, "PUT", "/admin/attributes/1", `{"name":"年龄"}`).Code)
}

func TestAttrCtrl_Delete_OK(t *testing.T) {
	svc := &testutil.MockAttributeService{
		DeleteFn: func(_ context.Context, id uint) error { return nil },
	}
	r := newTestRouter()
	r.DELETE("/admin/attributes/:id", attrCtrl(svc).Delete)
	assert.Equal(t, 204, doRequest(r, "DELETE", "/admin/attributes/1", "").Code)
}

func TestAttrCtrl_Delete_BadID(t *testing.T) {
	r := newTestRouter()
	r.DELETE("/admin/attributes/:id", attrCtrl(&testutil.MockAttributeService{}).Delete)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/admin/attributes/abc", "").Code)
}

func TestAttrCtrl_Delete_SvcErr(t *testing.T) {
	svc := &testutil.MockAttributeService{
		DeleteFn: func(_ context.Context, id uint) error {
			return apperrors.NewNotFound("not found", nil)
		},
	}
	r := newTestRouter()
	r.DELETE("/admin/attributes/:id", attrCtrl(svc).Delete)
	assert.Equal(t, 404, doRequest(r, "DELETE", "/admin/attributes/1", "").Code)
}

// ── User Attributes ───────────────────────────────────────────────────────────

func TestAttrCtrl_ListUserAttributes_OK(t *testing.T) {
	svc := &testutil.MockAttributeService{
		ListUserAttrsFn: func(_ context.Context, userID uint64) ([]*entity.UserAttribute, error) {
			return []*entity.UserAttribute{{ID: 1, UserID: userID, AttributeID: 1, ValueString: "男"}}, nil
		},
	}
	r := newTestRouter()
	r.GET("/admin/users/:id/attributes", attrCtrl(svc).ListUserAttributes)
	assert.Equal(t, 200, doRequest(r, "GET", "/admin/users/1/attributes", "").Code)
}

func TestAttrCtrl_ListUserAttributes_BadID(t *testing.T) {
	r := newTestRouter()
	r.GET("/admin/users/:id/attributes", attrCtrl(&testutil.MockAttributeService{}).ListUserAttributes)
	assert.Equal(t, 400, doRequest(r, "GET", "/admin/users/abc/attributes", "").Code)
}

func TestAttrCtrl_ListUserAttributes_SvcErr(t *testing.T) {
	svc := &testutil.MockAttributeService{
		ListUserAttrsFn: func(_ context.Context, userID uint64) ([]*entity.UserAttribute, error) {
			return nil, apperrors.NewNotFound("user not found", nil)
		},
	}
	r := newTestRouter()
	r.GET("/admin/users/:id/attributes", attrCtrl(svc).ListUserAttributes)
	assert.Equal(t, 404, doRequest(r, "GET", "/admin/users/1/attributes", "").Code)
}

func TestAttrCtrl_SetUserAttribute_OK(t *testing.T) {
	svc := &testutil.MockAttributeService{
		SetUserAttrFn: func(_ context.Context, userID uint64, req *dto.SetUserAttributeRequest) error { return nil },
	}
	r := newTestRouter()
	r.POST("/admin/users/:id/attributes", attrCtrl(svc).SetUserAttribute)
	assert.Equal(t, 200, doRequest(r, "POST", "/admin/users/1/attributes", `{"attribute_id":1,"value":"男"}`).Code)
}

func TestAttrCtrl_SetUserAttribute_BadID(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/users/:id/attributes", attrCtrl(&testutil.MockAttributeService{}).SetUserAttribute)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/users/abc/attributes", `{"attribute_id":1,"value":"男"}`).Code)
}

func TestAttrCtrl_SetUserAttribute_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/users/:id/attributes", attrCtrl(&testutil.MockAttributeService{}).SetUserAttribute)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/users/1/attributes", `bad`).Code)
}

func TestAttrCtrl_SetUserAttribute_Validation(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/users/:id/attributes", attrCtrl(&testutil.MockAttributeService{}).SetUserAttribute)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/users/1/attributes", `{"attribute_id":0,"value":"男"}`).Code)
}

func TestAttrCtrl_SetUserAttribute_SvcErr(t *testing.T) {
	svc := &testutil.MockAttributeService{
		SetUserAttrFn: func(_ context.Context, userID uint64, req *dto.SetUserAttributeRequest) error {
			return apperrors.NewNotFound("attr not found", nil)
		},
	}
	r := newTestRouter()
	r.POST("/admin/users/:id/attributes", attrCtrl(svc).SetUserAttribute)
	assert.Equal(t, 404, doRequest(r, "POST", "/admin/users/1/attributes", `{"attribute_id":1,"value":"男"}`).Code)
}

func TestAttrCtrl_DeleteUserAttribute_OK(t *testing.T) {
	svc := &testutil.MockAttributeService{
		DeleteUserAttrFn: func(_ context.Context, userID uint64, attributeID uint) error { return nil },
	}
	r := newTestRouter()
	r.DELETE("/admin/users/:id/attributes", attrCtrl(svc).DeleteUserAttribute)
	assert.Equal(t, 204, doRequest(r, "DELETE", "/admin/users/1/attributes?attribute_id=5", "").Code)
}

func TestAttrCtrl_DeleteUserAttribute_BadUserID(t *testing.T) {
	r := newTestRouter()
	r.DELETE("/admin/users/:id/attributes", attrCtrl(&testutil.MockAttributeService{}).DeleteUserAttribute)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/admin/users/abc/attributes?attribute_id=5", "").Code)
}

func TestAttrCtrl_DeleteUserAttribute_BadAttrID(t *testing.T) {
	r := newTestRouter()
	r.DELETE("/admin/users/:id/attributes", attrCtrl(&testutil.MockAttributeService{}).DeleteUserAttribute)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/admin/users/1/attributes?attribute_id=abc", "").Code)
}

func TestAttrCtrl_DeleteUserAttribute_SvcErr(t *testing.T) {
	svc := &testutil.MockAttributeService{
		DeleteUserAttrFn: func(_ context.Context, userID uint64, attributeID uint) error {
			return apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouter()
	r.DELETE("/admin/users/:id/attributes", attrCtrl(svc).DeleteUserAttribute)
	assert.Equal(t, 500, doRequest(r, "DELETE", "/admin/users/1/attributes?attribute_id=5", "").Code)
}
