package controller

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/testutil"
)

func newTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	log := logrus.New()
	log.SetOutput(new(bytes.Buffer))
	r.Use(middleware.ErrorMiddleware(log))
	return r
}

func newTestEngineWithAuth(userID uint64, userType int8) *gin.Engine {
	r := newTestEngine()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("user_type", userType)
		c.Next()
	})
	return r
}

func performRequest(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req, _ = http.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ==================== AuthController ====================

func TestAuthController_WechatLogin_Success(t *testing.T) {
	svc := &testutil.MockAuthService{
		WechatLoginFn: func(_ context.Context, req *dto.WechatLoginRequest) (*dto.LoginResponseData, error) {
			return &dto.LoginResponseData{AccessToken: "tok", TokenType: "Bearer", ExpiresIn: 3600}, nil
		},
	}
	r := newTestEngine()
	ctrl := NewAuthController(svc, logrus.New())
	r.POST("/auth/wechat-login", ctrl.WechatLogin)

	w := performRequest(r, "POST", "/auth/wechat-login", `{"code":"code123"}`)
	assert.Equal(t, 200, w.Code)
}

func TestAuthController_WechatLogin_InvalidJSON(t *testing.T) {
	r := newTestEngine()
	ctrl := NewAuthController(&testutil.MockAuthService{}, logrus.New())
	r.POST("/auth/wechat-login", ctrl.WechatLogin)

	w := performRequest(r, "POST", "/auth/wechat-login", `not-json`)
	assert.Equal(t, 400, w.Code)
}

func TestAuthController_WechatLogin_ValidationFail(t *testing.T) {
	r := newTestEngine()
	ctrl := NewAuthController(&testutil.MockAuthService{}, logrus.New())
	r.POST("/auth/wechat-login", ctrl.WechatLogin)

	// Empty code fails validation
	w := performRequest(r, "POST", "/auth/wechat-login", `{"code":""}`)
	assert.Equal(t, 422, w.Code)
}

func TestAuthController_WechatLogin_ServiceError(t *testing.T) {
	svc := &testutil.MockAuthService{
		WechatLoginFn: func(_ context.Context, req *dto.WechatLoginRequest) (*dto.LoginResponseData, error) {
			return nil, apperrors.NewUnauthorized("wechat error", nil)
		},
	}
	r := newTestEngine()
	ctrl := NewAuthController(svc, logrus.New())
	r.POST("/auth/wechat-login", ctrl.WechatLogin)

	w := performRequest(r, "POST", "/auth/wechat-login", `{"code":"code123"}`)
	assert.Equal(t, 401, w.Code)
}

func TestAuthController_AdminLogin_Success(t *testing.T) {
	svc := &testutil.MockAuthService{
		AdminLoginFn: func(_ context.Context, req *dto.AdminLoginRequest) (*dto.LoginResponseData, error) {
			return &dto.LoginResponseData{AccessToken: "tok", TokenType: "Bearer"}, nil
		},
	}
	r := newTestEngine()
	ctrl := NewAuthController(svc, logrus.New())
	r.POST("/auth/admin-login", ctrl.AdminLogin)

	w := performRequest(r, "POST", "/auth/admin-login", `{"email":"a@b.com","password":"password123"}`)
	assert.Equal(t, 200, w.Code)
}

func TestAuthController_AdminLogin_InvalidJSON(t *testing.T) {
	r := newTestEngine()
	ctrl := NewAuthController(&testutil.MockAuthService{}, logrus.New())
	r.POST("/auth/admin-login", ctrl.AdminLogin)

	w := performRequest(r, "POST", "/auth/admin-login", `not-json`)
	assert.Equal(t, 400, w.Code)
}

func TestAuthController_AdminLogin_ValidationFail(t *testing.T) {
	r := newTestEngine()
	ctrl := NewAuthController(&testutil.MockAuthService{}, logrus.New())
	r.POST("/auth/admin-login", ctrl.AdminLogin)

	w := performRequest(r, "POST", "/auth/admin-login", `{"email":"","password":""}`)
	assert.Equal(t, 422, w.Code)
}

func TestAuthController_AdminLogin_ServiceError(t *testing.T) {
	svc := &testutil.MockAuthService{
		AdminLoginFn: func(_ context.Context, req *dto.AdminLoginRequest) (*dto.LoginResponseData, error) {
			return nil, apperrors.NewUnauthorized("wrong credentials", nil)
		},
	}
	r := newTestEngine()
	ctrl := NewAuthController(svc, logrus.New())
	r.POST("/auth/admin-login", ctrl.AdminLogin)

	w := performRequest(r, "POST", "/auth/admin-login", `{"email":"a@b.com","password":"password123"}`)
	assert.Equal(t, 401, w.Code)
}

func TestAuthController_RefreshToken_Success(t *testing.T) {
	svc := &testutil.MockAuthService{
		RefreshTokenFn: func(_ context.Context, userID uint64, userType int8) (*dto.LoginResponseData, error) {
			return &dto.LoginResponseData{AccessToken: "new_tok"}, nil
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewAuthController(svc, logrus.New())
	r.POST("/auth/refresh", ctrl.RefreshToken)

	w := performRequest(r, "POST", "/auth/refresh", "")
	assert.Equal(t, 200, w.Code)
}

func TestAuthController_RefreshToken_Unauthorized(t *testing.T) {
	r := newTestEngine() // no auth set
	ctrl := NewAuthController(&testutil.MockAuthService{}, logrus.New())
	r.POST("/auth/refresh", ctrl.RefreshToken)

	w := performRequest(r, "POST", "/auth/refresh", "")
	assert.Equal(t, 401, w.Code)
}

func TestAuthController_RefreshToken_ServiceError(t *testing.T) {
	svc := &testutil.MockAuthService{
		RefreshTokenFn: func(_ context.Context, userID uint64, userType int8) (*dto.LoginResponseData, error) {
			return nil, apperrors.NewUnauthorized("user not found", nil)
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewAuthController(svc, logrus.New())
	r.POST("/auth/refresh", ctrl.RefreshToken)

	w := performRequest(r, "POST", "/auth/refresh", "")
	assert.Equal(t, 401, w.Code)
}

// ==================== UserController ====================

func TestUserController_GetProfile_Success(t *testing.T) {
	svc := &testutil.MockUserService{
		GetProfileFn: func(_ context.Context, userID uint64) (*entity.User, error) {
			return &entity.User{ID: 1, Nickname: "Alice"}, nil
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewUserController(svc, logrus.New())
	r.GET("/users/profile", ctrl.GetProfile)

	w := performRequest(r, "GET", "/users/profile", "")
	assert.Equal(t, 200, w.Code)
}

func TestUserController_GetProfile_Unauthorized(t *testing.T) {
	r := newTestEngine()
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.GET("/users/profile", ctrl.GetProfile)

	w := performRequest(r, "GET", "/users/profile", "")
	assert.Equal(t, 401, w.Code)
}

func TestUserController_GetProfile_ServiceError(t *testing.T) {
	svc := &testutil.MockUserService{
		GetProfileFn: func(_ context.Context, userID uint64) (*entity.User, error) {
			return nil, apperrors.NewNotFound("not found", nil)
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewUserController(svc, logrus.New())
	r.GET("/users/profile", ctrl.GetProfile)

	w := performRequest(r, "GET", "/users/profile", "")
	assert.Equal(t, 404, w.Code)
}

func TestUserController_UpdateProfile_Success(t *testing.T) {
	svc := &testutil.MockUserService{
		UpdateProfileFn: func(_ context.Context, userID uint64, req *dto.UserProfileUpdateRequest) error {
			return nil
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewUserController(svc, logrus.New())
	r.PUT("/users/profile", ctrl.UpdateProfile)

	w := performRequest(r, "PUT", "/users/profile", `{"nickname":"Bob","avatar_url":"http://x.com/a.jpg"}`)
	assert.Equal(t, 200, w.Code)
}

func TestUserController_UpdateProfile_Unauthorized(t *testing.T) {
	r := newTestEngine()
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.PUT("/users/profile", ctrl.UpdateProfile)

	w := performRequest(r, "PUT", "/users/profile", `{"nickname":"Bob"}`)
	assert.Equal(t, 401, w.Code)
}

func TestUserController_UpdateProfile_InvalidJSON(t *testing.T) {
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.PUT("/users/profile", ctrl.UpdateProfile)

	w := performRequest(r, "PUT", "/users/profile", `not-json`)
	assert.Equal(t, 400, w.Code)
}

func TestUserController_UpdateProfile_ServiceError(t *testing.T) {
	svc := &testutil.MockUserService{
		UpdateProfileFn: func(_ context.Context, userID uint64, req *dto.UserProfileUpdateRequest) error {
			return apperrors.NewInternal("db error", nil)
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewUserController(svc, logrus.New())
	r.PUT("/users/profile", ctrl.UpdateProfile)

	w := performRequest(r, "PUT", "/users/profile", `{"nickname":"Bob"}`)
	assert.Equal(t, 500, w.Code)
}

func TestUserController_GetPermissions_Success(t *testing.T) {
	svc := &testutil.MockUserService{
		GetPermissionsFn: func(_ context.Context, userID uint64) ([]string, []string, error) {
			return []string{"admin"}, []string{"user:list"}, nil
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewUserController(svc, logrus.New())
	r.GET("/users/permissions", ctrl.GetPermissions)

	w := performRequest(r, "GET", "/users/permissions", "")
	assert.Equal(t, 200, w.Code)
}

func TestUserController_GetPermissions_Unauthorized(t *testing.T) {
	r := newTestEngine()
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.GET("/users/permissions", ctrl.GetPermissions)

	w := performRequest(r, "GET", "/users/permissions", "")
	assert.Equal(t, 401, w.Code)
}

func TestUserController_GetPermissions_ServiceError(t *testing.T) {
	svc := &testutil.MockUserService{
		GetPermissionsFn: func(_ context.Context, userID uint64) ([]string, []string, error) {
			return nil, nil, apperrors.NewInternal("db error", nil)
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewUserController(svc, logrus.New())
	r.GET("/users/permissions", ctrl.GetPermissions)

	w := performRequest(r, "GET", "/users/permissions", "")
	assert.Equal(t, 500, w.Code)
}

// ==================== Admin User endpoints ====================

func TestUserController_List_Success(t *testing.T) {
	users := []*entity.User{{ID: 1}, {ID: 2}}
	svc := &testutil.MockUserService{
		ListFn: func(_ context.Context, page, pageSize int, keyword string, userType, status *int8) ([]*entity.User, int64, error) {
			return users, 2, nil
		},
	}
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(svc, logrus.New())
	r.GET("/admin/users", ctrl.List)

	w := performRequest(r, "GET", "/admin/users", "")
	assert.Equal(t, 200, w.Code)
}

func TestUserController_List_ServiceError(t *testing.T) {
	svc := &testutil.MockUserService{
		ListFn: func(_ context.Context, page, pageSize int, keyword string, userType, status *int8) ([]*entity.User, int64, error) {
			return nil, 0, apperrors.NewInternal("db error", nil)
		},
	}
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(svc, logrus.New())
	r.GET("/admin/users", ctrl.List)

	w := performRequest(r, "GET", "/admin/users", "")
	assert.Equal(t, 500, w.Code)
}

func TestUserController_GetByID_Success(t *testing.T) {
	svc := &testutil.MockUserService{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return &entity.User{ID: 1}, nil
		},
	}
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(svc, logrus.New())
	r.GET("/admin/users/:id", ctrl.GetByID)

	w := performRequest(r, "GET", "/admin/users/1", "")
	assert.Equal(t, 200, w.Code)
}

func TestUserController_GetByID_InvalidID(t *testing.T) {
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.GET("/admin/users/:id", ctrl.GetByID)

	w := performRequest(r, "GET", "/admin/users/abc", "")
	assert.Equal(t, 400, w.Code)
}

func TestUserController_GetByID_ServiceError(t *testing.T) {
	svc := &testutil.MockUserService{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return nil, apperrors.NewNotFound("not found", nil)
		},
	}
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(svc, logrus.New())
	r.GET("/admin/users/:id", ctrl.GetByID)

	w := performRequest(r, "GET", "/admin/users/1", "")
	assert.Equal(t, 404, w.Code)
}

func TestUserController_CreateAdmin_Success(t *testing.T) {
	svc := &testutil.MockUserService{
		CreateAdminUserFn: func(_ context.Context, req *dto.CreateAdminUserRequest) (uint64, error) {
			return 10, nil
		},
	}
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(svc, logrus.New())
	r.POST("/admin/users", ctrl.CreateAdminUser)

	body := `{"email":"a@b.com","password":"password123","nickname":"Admin","user_type":2}`
	w := performRequest(r, "POST", "/admin/users", body)
	assert.Equal(t, 201, w.Code)
}

func TestUserController_CreateAdmin_InvalidJSON(t *testing.T) {
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.POST("/admin/users", ctrl.CreateAdminUser)

	w := performRequest(r, "POST", "/admin/users", `not-json`)
	assert.Equal(t, 400, w.Code)
}

func TestUserController_CreateAdmin_ValidationFail(t *testing.T) {
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.POST("/admin/users", ctrl.CreateAdminUser)

	w := performRequest(r, "POST", "/admin/users", `{"email":"","password":"","user_type":2}`)
	assert.Equal(t, 422, w.Code)
}

func TestUserController_CreateAdmin_ServiceError(t *testing.T) {
	svc := &testutil.MockUserService{
		CreateAdminUserFn: func(_ context.Context, req *dto.CreateAdminUserRequest) (uint64, error) {
			return 0, apperrors.NewConflict("email exists", nil)
		},
	}
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(svc, logrus.New())
	r.POST("/admin/users", ctrl.CreateAdminUser)

	body := `{"email":"a@b.com","password":"password123","nickname":"Admin","user_type":2}`
	w := performRequest(r, "POST", "/admin/users", body)
	assert.Equal(t, 409, w.Code)
}

func TestUserController_UpdateUser_Success(t *testing.T) {
	svc := &testutil.MockUserService{
		UpdateUserFn: func(_ context.Context, id uint64, req *dto.UpdateUserRequest, operatorID uint64) error {
			return nil
		},
	}
	r := newTestEngineWithAuth(99, 2)
	ctrl := NewUserController(svc, logrus.New())
	r.PUT("/admin/users/:id", ctrl.UpdateUser)

	w := performRequest(r, "PUT", "/admin/users/1", `{"nickname":"NewName","status":1}`)
	assert.Equal(t, 200, w.Code)
}

func TestUserController_UpdateUser_InvalidID(t *testing.T) {
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.PUT("/admin/users/:id", ctrl.UpdateUser)

	w := performRequest(r, "PUT", "/admin/users/abc", `{}`)
	assert.Equal(t, 400, w.Code)
}

func TestUserController_UpdateUser_InvalidJSON(t *testing.T) {
	r := newTestEngineWithAuth(99, 2)
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.PUT("/admin/users/:id", ctrl.UpdateUser)

	w := performRequest(r, "PUT", "/admin/users/1", `not-json`)
	assert.Equal(t, 400, w.Code)
}

func TestUserController_UpdateUser_Unauthorized(t *testing.T) {
	r := newTestEngineWithAuth(99, 2)
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.PUT("/admin/users/:id", ctrl.UpdateUser)
	// No userID in context
	r2 := newTestEngine()
	ctrl2 := NewUserController(&testutil.MockUserService{}, logrus.New())
	r2.PUT("/admin/users/:id", ctrl2.UpdateUser)

	w := performRequest(r2, "PUT", "/admin/users/1", `{"nickname":"x"}`)
	assert.Equal(t, 401, w.Code)
}

func TestUserController_UpdateUser_ServiceError(t *testing.T) {
	svc := &testutil.MockUserService{
		UpdateUserFn: func(_ context.Context, id uint64, req *dto.UpdateUserRequest, operatorID uint64) error {
			return apperrors.NewForbidden("forbidden", nil)
		},
	}
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(svc, logrus.New())
	r.PUT("/admin/users/:id", ctrl.UpdateUser)

	w := performRequest(r, "PUT", "/admin/users/1", `{"user_type":3}`)
	assert.Equal(t, 403, w.Code)
}

func TestUserController_DeleteUser_Success(t *testing.T) {
	svc := &testutil.MockUserService{
		DeleteUserFn: func(_ context.Context, id uint64) error {
			return nil
		},
	}
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(svc, logrus.New())
	r.DELETE("/admin/users/:id", ctrl.DeleteUser)

	w := performRequest(r, "DELETE", "/admin/users/1", "")
	assert.Equal(t, 200, w.Code)
}

func TestUserController_DeleteUser_InvalidID(t *testing.T) {
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.DELETE("/admin/users/:id", ctrl.DeleteUser)

	w := performRequest(r, "DELETE", "/admin/users/abc", "")
	assert.Equal(t, 400, w.Code)
}

func TestUserController_DeleteUser_ServiceError(t *testing.T) {
	svc := &testutil.MockUserService{
		DeleteUserFn: func(_ context.Context, id uint64) error {
			return apperrors.NewNotFound("not found", nil)
		},
	}
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(svc, logrus.New())
	r.DELETE("/admin/users/:id", ctrl.DeleteUser)

	w := performRequest(r, "DELETE", "/admin/users/1", "")
	assert.Equal(t, 404, w.Code)
}

func TestUserController_AssignRoles_Success(t *testing.T) {
	svc := &testutil.MockUserService{
		AssignRolesFn: func(_ context.Context, userID uint64, req *dto.AssignRolesRequest) error {
			return nil
		},
	}
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(svc, logrus.New())
	r.POST("/admin/users/:id/roles", ctrl.AssignRoles)

	w := performRequest(r, "POST", "/admin/users/1/roles", `{"role_ids":[1,2]}`)
	assert.Equal(t, 200, w.Code)
}

func TestUserController_AssignRoles_InvalidID(t *testing.T) {
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.POST("/admin/users/:id/roles", ctrl.AssignRoles)

	w := performRequest(r, "POST", "/admin/users/abc/roles", `{"role_ids":[1]}`)
	assert.Equal(t, 400, w.Code)
}

func TestUserController_AssignRoles_InvalidJSON(t *testing.T) {
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.POST("/admin/users/:id/roles", ctrl.AssignRoles)

	w := performRequest(r, "POST", "/admin/users/1/roles", `not-json`)
	assert.Equal(t, 400, w.Code)
}

func TestUserController_AssignRoles_ServiceError(t *testing.T) {
	svc := &testutil.MockUserService{
		AssignRolesFn: func(_ context.Context, userID uint64, req *dto.AssignRolesRequest) error {
			return apperrors.NewNotFound("not found", nil)
		},
	}
	r := newTestEngineWithAuth(1, 2)
	ctrl := NewUserController(svc, logrus.New())
	r.POST("/admin/users/:id/roles", ctrl.AssignRoles)

	w := performRequest(r, "POST", "/admin/users/1/roles", `{"role_ids":[1]}`)
	assert.Equal(t, 404, w.Code)
}

func TestUserController_AddTag_Success(t *testing.T) {
	svc := &testutil.MockUserService{
		AddTagFn: func(_ context.Context, userID uint64, req *dto.AddTagRequest) (uint, error) {
			return 1, nil
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewUserController(svc, logrus.New())
	r.POST("/users/:id/tags", ctrl.AddTag)

	w := performRequest(r, "POST", "/users/1/tags", `{"tag_name":"go"}`)
	assert.Equal(t, 201, w.Code)
}

func TestUserController_AddTag_InvalidID(t *testing.T) {
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.POST("/users/:id/tags", ctrl.AddTag)

	w := performRequest(r, "POST", "/users/abc/tags", `{"tag_name":"go"}`)
	assert.Equal(t, 400, w.Code)
}

func TestUserController_AddTag_InvalidJSON(t *testing.T) {
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.POST("/users/:id/tags", ctrl.AddTag)

	w := performRequest(r, "POST", "/users/1/tags", `not-json`)
	assert.Equal(t, 400, w.Code)
}

func TestUserController_AddTag_ValidationFail(t *testing.T) {
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.POST("/users/:id/tags", ctrl.AddTag)

	w := performRequest(r, "POST", "/users/1/tags", `{"tag_name":""}`)
	assert.Equal(t, 422, w.Code)
}

func TestUserController_AddTag_ServiceError(t *testing.T) {
	svc := &testutil.MockUserService{
		AddTagFn: func(_ context.Context, userID uint64, req *dto.AddTagRequest) (uint, error) {
			return 0, apperrors.NewBadRequest("limit exceeded", nil)
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewUserController(svc, logrus.New())
	r.POST("/users/:id/tags", ctrl.AddTag)

	w := performRequest(r, "POST", "/users/1/tags", `{"tag_name":"go"}`)
	assert.Equal(t, 400, w.Code)
}

func TestUserController_DeleteTag_Success(t *testing.T) {
	svc := &testutil.MockUserService{
		DeleteTagFn: func(_ context.Context, userID, tagID uint64) error {
			return nil
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewUserController(svc, logrus.New())
	r.DELETE("/users/:id/tags/:tag_id", ctrl.DeleteTag)

	w := performRequest(r, "DELETE", "/users/1/tags/5", "")
	assert.Equal(t, 200, w.Code)
}

func TestUserController_DeleteTag_InvalidUserID(t *testing.T) {
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.DELETE("/users/:id/tags/:tag_id", ctrl.DeleteTag)

	w := performRequest(r, "DELETE", "/users/abc/tags/5", "")
	assert.Equal(t, 400, w.Code)
}

func TestUserController_DeleteTag_InvalidTagID(t *testing.T) {
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewUserController(&testutil.MockUserService{}, logrus.New())
	r.DELETE("/users/:id/tags/:tag_id", ctrl.DeleteTag)

	w := performRequest(r, "DELETE", "/users/1/tags/abc", "")
	assert.Equal(t, 400, w.Code)
}

func TestUserController_DeleteTag_ServiceError(t *testing.T) {
	svc := &testutil.MockUserService{
		DeleteTagFn: func(_ context.Context, userID, tagID uint64) error {
			return apperrors.NewInternal("db error", nil)
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewUserController(svc, logrus.New())
	r.DELETE("/users/:id/tags/:tag_id", ctrl.DeleteTag)

	w := performRequest(r, "DELETE", "/users/1/tags/5", "")
	assert.Equal(t, 500, w.Code)
}

// ==================== SystemController ====================

func TestSystemController_GetWechatConfig_Success(t *testing.T) {
	wechatSvc := &testutil.MockWechatConfigService{
		GetFn: func(_ context.Context) (*entity.WechatConfig, error) {
			return &entity.WechatConfig{AppID: "wx123"}, nil
		},
	}
	r := newTestEngine()
	ctrl := NewSystemController(wechatSvc, &testutil.MockAuditLogService{}, &testutil.MockLogConfigService{}, logrus.New())
	r.GET("/admin/wechat-config", ctrl.GetWechatConfig)

	w := performRequest(r, "GET", "/admin/wechat-config", "")
	assert.Equal(t, 200, w.Code)
}

func TestSystemController_GetWechatConfig_Error(t *testing.T) {
	wechatSvc := &testutil.MockWechatConfigService{
		GetFn: func(_ context.Context) (*entity.WechatConfig, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	r := newTestEngine()
	ctrl := NewSystemController(wechatSvc, &testutil.MockAuditLogService{}, &testutil.MockLogConfigService{}, logrus.New())
	r.GET("/admin/wechat-config", ctrl.GetWechatConfig)

	w := performRequest(r, "GET", "/admin/wechat-config", "")
	assert.Equal(t, 500, w.Code)
}

func TestSystemController_UpdateWechatConfig_Success(t *testing.T) {
	wechatSvc := &testutil.MockWechatConfigService{
		UpdateFn: func(_ context.Context, req *dto.UpdateWechatConfigRequest) error {
			return nil
		},
	}
	r := newTestEngine()
	ctrl := NewSystemController(wechatSvc, &testutil.MockAuditLogService{}, &testutil.MockLogConfigService{}, logrus.New())
	r.PUT("/admin/wechat-config", ctrl.UpdateWechatConfig)

	w := performRequest(r, "PUT", "/admin/wechat-config", `{"app_id":"wx123","app_secret":"secret","api_token":"tok"}`)
	assert.Equal(t, 200, w.Code)
}

func TestSystemController_UpdateWechatConfig_InvalidJSON(t *testing.T) {
	r := newTestEngine()
	ctrl := NewSystemController(&testutil.MockWechatConfigService{}, &testutil.MockAuditLogService{}, &testutil.MockLogConfigService{}, logrus.New())
	r.PUT("/admin/wechat-config", ctrl.UpdateWechatConfig)

	w := performRequest(r, "PUT", "/admin/wechat-config", `not-json`)
	assert.Equal(t, 400, w.Code)
}

func TestSystemController_UpdateWechatConfig_ValidationFail(t *testing.T) {
	r := newTestEngine()
	ctrl := NewSystemController(&testutil.MockWechatConfigService{}, &testutil.MockAuditLogService{}, &testutil.MockLogConfigService{}, logrus.New())
	r.PUT("/admin/wechat-config", ctrl.UpdateWechatConfig)

	w := performRequest(r, "PUT", "/admin/wechat-config", `{"app_id":"","app_secret":""}`)
	assert.Equal(t, 422, w.Code)
}

func TestSystemController_UpdateWechatConfig_ServiceError(t *testing.T) {
	wechatSvc := &testutil.MockWechatConfigService{
		UpdateFn: func(_ context.Context, req *dto.UpdateWechatConfigRequest) error {
			return apperrors.NewInternal("db error", nil)
		},
	}
	r := newTestEngine()
	ctrl := NewSystemController(wechatSvc, &testutil.MockAuditLogService{}, &testutil.MockLogConfigService{}, logrus.New())
	r.PUT("/admin/wechat-config", ctrl.UpdateWechatConfig)

	w := performRequest(r, "PUT", "/admin/wechat-config", `{"app_id":"wx123","app_secret":"secret"}`)
	assert.Equal(t, 500, w.Code)
}

func TestSystemController_ListAuditLogs_Success(t *testing.T) {
	auditSvc := &testutil.MockAuditLogService{
		ListFn: func(_ context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error) {
			return []*entity.AuditLog{{ID: 1}}, 1, nil
		},
	}
	r := newTestEngine()
	ctrl := NewSystemController(&testutil.MockWechatConfigService{}, auditSvc, &testutil.MockLogConfigService{}, logrus.New())
	r.GET("/admin/audit-logs", ctrl.ListAuditLogs)

	w := performRequest(r, "GET", "/admin/audit-logs?start_time=2026-01-01&end_time=2026-12-31", "")
	assert.Equal(t, 200, w.Code)
}

func TestSystemController_ListAuditLogs_ServiceError(t *testing.T) {
	auditSvc := &testutil.MockAuditLogService{
		ListFn: func(_ context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error) {
			return nil, 0, apperrors.NewInternal("db error", nil)
		},
	}
	r := newTestEngine()
	ctrl := NewSystemController(&testutil.MockWechatConfigService{}, auditSvc, &testutil.MockLogConfigService{}, logrus.New())
	r.GET("/admin/audit-logs", ctrl.ListAuditLogs)

	w := performRequest(r, "GET", "/admin/audit-logs", "")
	assert.Equal(t, 500, w.Code)
}

func TestSystemController_GetLogConfig_Success(t *testing.T) {
	logConfigSvc := &testutil.MockLogConfigService{
		GetFn: func(_ context.Context) (*entity.LogConfig, error) {
			return &entity.LogConfig{RetentionDays: 30}, nil
		},
	}
	r := newTestEngine()
	ctrl := NewSystemController(&testutil.MockWechatConfigService{}, &testutil.MockAuditLogService{}, logConfigSvc, logrus.New())
	r.GET("/admin/log-config", ctrl.GetLogConfig)

	w := performRequest(r, "GET", "/admin/log-config", "")
	assert.Equal(t, 200, w.Code)
}

func TestSystemController_GetLogConfig_Error(t *testing.T) {
	logConfigSvc := &testutil.MockLogConfigService{
		GetFn: func(_ context.Context) (*entity.LogConfig, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	r := newTestEngine()
	ctrl := NewSystemController(&testutil.MockWechatConfigService{}, &testutil.MockAuditLogService{}, logConfigSvc, logrus.New())
	r.GET("/admin/log-config", ctrl.GetLogConfig)

	w := performRequest(r, "GET", "/admin/log-config", "")
	assert.Equal(t, 500, w.Code)
}

func TestSystemController_UpdateLogConfig_Success(t *testing.T) {
	logConfigSvc := &testutil.MockLogConfigService{
		UpdateFn: func(_ context.Context, req *dto.UpdateLogConfigRequest) error {
			return nil
		},
	}
	r := newTestEngine()
	ctrl := NewSystemController(&testutil.MockWechatConfigService{}, &testutil.MockAuditLogService{}, logConfigSvc, logrus.New())
	r.PUT("/admin/log-config", ctrl.UpdateLogConfig)

	w := performRequest(r, "PUT", "/admin/log-config", `{"retention_days":60}`)
	assert.Equal(t, 200, w.Code)
}

func TestSystemController_UpdateLogConfig_InvalidJSON(t *testing.T) {
	r := newTestEngine()
	ctrl := NewSystemController(&testutil.MockWechatConfigService{}, &testutil.MockAuditLogService{}, &testutil.MockLogConfigService{}, logrus.New())
	r.PUT("/admin/log-config", ctrl.UpdateLogConfig)

	w := performRequest(r, "PUT", "/admin/log-config", `not-json`)
	assert.Equal(t, 400, w.Code)
}

func TestSystemController_UpdateLogConfig_ValidationFail(t *testing.T) {
	r := newTestEngine()
	ctrl := NewSystemController(&testutil.MockWechatConfigService{}, &testutil.MockAuditLogService{}, &testutil.MockLogConfigService{}, logrus.New())
	r.PUT("/admin/log-config", ctrl.UpdateLogConfig)

	w := performRequest(r, "PUT", "/admin/log-config", `{"retention_days":0}`)
	assert.Equal(t, 422, w.Code)
}

func TestSystemController_UpdateLogConfig_ServiceError(t *testing.T) {
	logConfigSvc := &testutil.MockLogConfigService{
		UpdateFn: func(_ context.Context, req *dto.UpdateLogConfigRequest) error {
			return apperrors.NewInternal("db error", nil)
		},
	}
	r := newTestEngine()
	ctrl := NewSystemController(&testutil.MockWechatConfigService{}, &testutil.MockAuditLogService{}, logConfigSvc, logrus.New())
	r.PUT("/admin/log-config", ctrl.UpdateLogConfig)

	w := performRequest(r, "PUT", "/admin/log-config", `{"retention_days":60}`)
	assert.Equal(t, 500, w.Code)
}

// ==================== NotificationController ====================

func TestNotificationController_List_Success(t *testing.T) {
	svc := &testutil.MockNotificationService{
		ListFn: func(_ context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, int64, error) {
			return []*entity.Notification{{ID: 1}}, 1, 0, nil
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewNotificationController(svc, logrus.New())
	r.GET("/notifications", ctrl.List)

	w := performRequest(r, "GET", "/notifications?is_read=true", "")
	assert.Equal(t, 200, w.Code)
}

func TestNotificationController_List_WithIsReadFalse(t *testing.T) {
	svc := &testutil.MockNotificationService{
		ListFn: func(_ context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, int64, error) {
			return []*entity.Notification{}, 0, 0, nil
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewNotificationController(svc, logrus.New())
	r.GET("/notifications", ctrl.List)

	w := performRequest(r, "GET", "/notifications?is_read=false", "")
	assert.Equal(t, 200, w.Code)
}

func TestNotificationController_List_Unauthorized(t *testing.T) {
	r := newTestEngine()
	ctrl := NewNotificationController(&testutil.MockNotificationService{}, logrus.New())
	r.GET("/notifications", ctrl.List)

	w := performRequest(r, "GET", "/notifications", "")
	assert.Equal(t, 401, w.Code)
}

func TestNotificationController_List_ServiceError(t *testing.T) {
	svc := &testutil.MockNotificationService{
		ListFn: func(_ context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, int64, error) {
			return nil, 0, 0, apperrors.NewInternal("db error", nil)
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewNotificationController(svc, logrus.New())
	r.GET("/notifications", ctrl.List)

	w := performRequest(r, "GET", "/notifications", "")
	assert.Equal(t, 500, w.Code)
}

func TestNotificationController_MarkRead_Success(t *testing.T) {
	svc := &testutil.MockNotificationService{
		MarkReadFn: func(_ context.Context, id uint64) error {
			return nil
		},
	}
	r := newTestEngine()
	ctrl := NewNotificationController(svc, logrus.New())
	r.PUT("/notifications/:id/read", ctrl.MarkRead)

	w := performRequest(r, "PUT", "/notifications/1/read", "")
	assert.Equal(t, 200, w.Code)
}

func TestNotificationController_MarkRead_InvalidID(t *testing.T) {
	r := newTestEngine()
	ctrl := NewNotificationController(&testutil.MockNotificationService{}, logrus.New())
	r.PUT("/notifications/:id/read", ctrl.MarkRead)

	w := performRequest(r, "PUT", "/notifications/abc/read", "")
	assert.Equal(t, 400, w.Code)
}

func TestNotificationController_MarkRead_ServiceError(t *testing.T) {
	svc := &testutil.MockNotificationService{
		MarkReadFn: func(_ context.Context, id uint64) error {
			return apperrors.NewInternal("db error", nil)
		},
	}
	r := newTestEngine()
	ctrl := NewNotificationController(svc, logrus.New())
	r.PUT("/notifications/:id/read", ctrl.MarkRead)

	w := performRequest(r, "PUT", "/notifications/1/read", "")
	assert.Equal(t, 500, w.Code)
}

func TestNotificationController_MarkAllRead_Success(t *testing.T) {
	svc := &testutil.MockNotificationService{
		MarkAllReadFn: func(_ context.Context, userID uint64) error {
			return nil
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewNotificationController(svc, logrus.New())
	r.PUT("/notifications/read-all", ctrl.MarkAllRead)

	w := performRequest(r, "PUT", "/notifications/read-all", "")
	assert.Equal(t, 200, w.Code)
}

func TestNotificationController_MarkAllRead_Unauthorized(t *testing.T) {
	r := newTestEngine()
	ctrl := NewNotificationController(&testutil.MockNotificationService{}, logrus.New())
	r.PUT("/notifications/read-all", ctrl.MarkAllRead)

	w := performRequest(r, "PUT", "/notifications/read-all", "")
	assert.Equal(t, 401, w.Code)
}

func TestNotificationController_MarkAllRead_ServiceError(t *testing.T) {
	svc := &testutil.MockNotificationService{
		MarkAllReadFn: func(_ context.Context, userID uint64) error {
			return apperrors.NewInternal("db error", nil)
		},
	}
	r := newTestEngineWithAuth(1, 1)
	ctrl := NewNotificationController(svc, logrus.New())
	r.PUT("/notifications/read-all", ctrl.MarkAllRead)

	w := performRequest(r, "PUT", "/notifications/read-all", "")
	assert.Equal(t, 500, w.Code)
}

// ==================== UploadController ====================

func TestUploadController_UploadImage_NoFile(t *testing.T) {
	r := newTestEngine()
	ctrl := NewUploadController("/tmp/uploads", "http://localhost", logrus.New())
	r.POST("/upload/image", ctrl.UploadImage)

	req, _ := http.NewRequest("POST", "/upload/image", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestUploadController_UploadVideo_NoFile(t *testing.T) {
	r := newTestEngine()
	ctrl := NewUploadController("/tmp/uploads", "http://localhost", logrus.New())
	r.POST("/upload/video", ctrl.UploadVideo)

	req, _ := http.NewRequest("POST", "/upload/video", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestUploadController_UploadImage_UnsupportedType(t *testing.T) {
	r := newTestEngine()
	ctrl := NewUploadController("/tmp/uploads", "http://localhost", logrus.New())
	r.POST("/upload/image", ctrl.UploadImage)

	body, contentType := testutil.CreateMultipartFile(t, "file", "test.pdf", []byte("pdf content"))
	req, _ := http.NewRequest("POST", "/upload/image", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestUploadController_UploadVideo_UnsupportedType(t *testing.T) {
	r := newTestEngine()
	ctrl := NewUploadController("/tmp/uploads", "http://localhost", logrus.New())
	r.POST("/upload/video", ctrl.UploadVideo)

	body, contentType := testutil.CreateMultipartFile(t, "file", "test.avi", []byte("avi content"))
	req, _ := http.NewRequest("POST", "/upload/video", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}

// ==================== Ensure compilation of user fields ====================

var _ require.TestingT = (*testing.T)(nil)
