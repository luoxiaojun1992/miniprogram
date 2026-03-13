package service

import (
	"context"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/testutil"
)

func newAuthService(
	userRepo *testutil.MockUserRepository,
	adminRepo *testutil.MockAdminUserRepository,
	wechat *testutil.MockWechatClient,
) AuthService {
	return NewAuthService(userRepo, adminRepo, wechat, "secret", 3600, logrus.New())
}

// ==================== WechatLogin ====================

func TestAuthService_WechatLogin_NewUser(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByOpenIDFn: func(_ context.Context, openID string) (*entity.User, error) {
			return nil, nil // user not found
		},
		CreateFn: func(_ context.Context, u *entity.User) error {
			u.ID = 1
			return nil
		},
	}
	adminRepo := &testutil.MockAdminUserRepository{}
	wechatClient := &testutil.MockWechatClient{
		Code2SessionFn: func(_ context.Context, code string) (string, error) {
			return "open_id_123", nil
		},
	}

	svc := newAuthService(userRepo, adminRepo, wechatClient)
	resp, err := svc.WechatLogin(context.Background(), &dto.WechatLoginRequest{Code: "code"})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.Equal(t, "Bearer", resp.TokenType)
}

func TestAuthService_WechatLogin_ExistingUser(t *testing.T) {
	existing := &entity.User{ID: 5, OpenID: "open_id_123", UserType: 1, Status: 1}
	userRepo := &testutil.MockUserRepository{
		GetByOpenIDFn: func(_ context.Context, openID string) (*entity.User, error) {
			return existing, nil
		},
	}
	adminRepo := &testutil.MockAdminUserRepository{}
	wechatClient := &testutil.MockWechatClient{
		Code2SessionFn: func(_ context.Context, code string) (string, error) {
			return "open_id_123", nil
		},
	}

	svc := newAuthService(userRepo, adminRepo, wechatClient)
	resp, err := svc.WechatLogin(context.Background(), &dto.WechatLoginRequest{Code: "code"})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
}

func TestAuthService_WechatLogin_WechatFails(t *testing.T) {
	userRepo := &testutil.MockUserRepository{}
	adminRepo := &testutil.MockAdminUserRepository{}
	wechatClient := &testutil.MockWechatClient{
		Code2SessionFn: func(_ context.Context, code string) (string, error) {
			return "", errors.New("wechat error")
		},
	}

	svc := newAuthService(userRepo, adminRepo, wechatClient)
	_, err := svc.WechatLogin(context.Background(), &dto.WechatLoginRequest{Code: "code"})
	require.Error(t, err)
	assert.IsType(t, &apperrors.AppError{}, err)
}

func TestAuthService_WechatLogin_GetOpenIDFails(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByOpenIDFn: func(_ context.Context, openID string) (*entity.User, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	adminRepo := &testutil.MockAdminUserRepository{}
	wechatClient := &testutil.MockWechatClient{
		Code2SessionFn: func(_ context.Context, code string) (string, error) {
			return "open_id", nil
		},
	}

	svc := newAuthService(userRepo, adminRepo, wechatClient)
	_, err := svc.WechatLogin(context.Background(), &dto.WechatLoginRequest{Code: "code"})
	require.Error(t, err)
}

func TestAuthService_WechatLogin_CreateFails(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByOpenIDFn: func(_ context.Context, openID string) (*entity.User, error) {
			return nil, nil
		},
		CreateFn: func(_ context.Context, u *entity.User) error {
			return apperrors.NewInternal("create failed", nil)
		},
	}
	adminRepo := &testutil.MockAdminUserRepository{}
	wechatClient := &testutil.MockWechatClient{
		Code2SessionFn: func(_ context.Context, code string) (string, error) {
			return "open_id", nil
		},
	}

	svc := newAuthService(userRepo, adminRepo, wechatClient)
	_, err := svc.WechatLogin(context.Background(), &dto.WechatLoginRequest{Code: "code"})
	require.Error(t, err)
}

// ==================== AdminLogin ====================

func TestAuthService_AdminLogin_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	admin := &entity.AdminUser{ID: 1, UserID: 10, Email: "admin@test.com", PasswordHash: string(hash)}
	user := &entity.User{ID: 10, UserType: 2, Status: 1}

	adminRepo := &testutil.MockAdminUserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*entity.AdminUser, error) {
			return admin, nil
		},
		UpdateLastLoginFn: func(_ context.Context, id uint64) error {
			return nil
		},
	}
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return user, nil
		},
	}
	wechatClient := &testutil.MockWechatClient{}

	svc := newAuthService(userRepo, adminRepo, wechatClient)
	resp, err := svc.AdminLogin(context.Background(), &dto.AdminLoginRequest{
		Email:    "admin@test.com",
		Password: "password123",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
}

func TestAuthService_AdminLogin_AdminNotFound(t *testing.T) {
	adminRepo := &testutil.MockAdminUserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*entity.AdminUser, error) {
			return nil, nil
		},
	}
	userRepo := &testutil.MockUserRepository{}
	wechatClient := &testutil.MockWechatClient{}

	svc := newAuthService(userRepo, adminRepo, wechatClient)
	_, err := svc.AdminLogin(context.Background(), &dto.AdminLoginRequest{
		Email: "nope@test.com", Password: "pw",
	})
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	assert.Equal(t, 401001, appErr.Code)
}

func TestAuthService_AdminLogin_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	admin := &entity.AdminUser{ID: 1, UserID: 10, Email: "a@b.com", PasswordHash: string(hash)}
	adminRepo := &testutil.MockAdminUserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*entity.AdminUser, error) {
			return admin, nil
		},
	}
	userRepo := &testutil.MockUserRepository{}
	wechatClient := &testutil.MockWechatClient{}

	svc := newAuthService(userRepo, adminRepo, wechatClient)
	_, err := svc.AdminLogin(context.Background(), &dto.AdminLoginRequest{
		Email: "a@b.com", Password: "wrong",
	})
	require.Error(t, err)
}

func TestAuthService_AdminLogin_GetEmailFails(t *testing.T) {
	adminRepo := &testutil.MockAdminUserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*entity.AdminUser, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	userRepo := &testutil.MockUserRepository{}
	wechatClient := &testutil.MockWechatClient{}

	svc := newAuthService(userRepo, adminRepo, wechatClient)
	_, err := svc.AdminLogin(context.Background(), &dto.AdminLoginRequest{
		Email: "a@b.com", Password: "pw",
	})
	require.Error(t, err)
}

func TestAuthService_AdminLogin_UserFrozen(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	admin := &entity.AdminUser{ID: 1, UserID: 10, Email: "a@b.com", PasswordHash: string(hash)}
	frozen := &entity.User{ID: 10, Status: 0}

	adminRepo := &testutil.MockAdminUserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*entity.AdminUser, error) {
			return admin, nil
		},
	}
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return frozen, nil
		},
	}
	wechatClient := &testutil.MockWechatClient{}

	svc := newAuthService(userRepo, adminRepo, wechatClient)
	_, err := svc.AdminLogin(context.Background(), &dto.AdminLoginRequest{
		Email: "a@b.com", Password: "pw",
	})
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	assert.Equal(t, 401001, appErr.Code)
}

func TestAuthService_AdminLogin_UserNil(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	admin := &entity.AdminUser{ID: 1, UserID: 10, Email: "a@b.com", PasswordHash: string(hash)}

	adminRepo := &testutil.MockAdminUserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*entity.AdminUser, error) {
			return admin, nil
		},
	}
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return nil, nil
		},
	}
	wechatClient := &testutil.MockWechatClient{}

	svc := newAuthService(userRepo, adminRepo, wechatClient)
	_, err := svc.AdminLogin(context.Background(), &dto.AdminLoginRequest{
		Email: "a@b.com", Password: "pw",
	})
	require.Error(t, err)
}

func TestAuthService_AdminLogin_GetUserFails(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	admin := &entity.AdminUser{ID: 1, UserID: 10, Email: "a@b.com", PasswordHash: string(hash)}

	adminRepo := &testutil.MockAdminUserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*entity.AdminUser, error) {
			return admin, nil
		},
	}
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	wechatClient := &testutil.MockWechatClient{}

	svc := newAuthService(userRepo, adminRepo, wechatClient)
	_, err := svc.AdminLogin(context.Background(), &dto.AdminLoginRequest{
		Email: "a@b.com", Password: "pw",
	})
	require.Error(t, err)
}

// ==================== RefreshToken ====================

func TestAuthService_RefreshToken_Success(t *testing.T) {
	user := &entity.User{ID: 1, UserType: 1, Status: 1}
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return user, nil
		},
	}
	svc := newAuthService(userRepo, &testutil.MockAdminUserRepository{}, &testutil.MockWechatClient{})
	resp, err := svc.RefreshToken(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
}

func TestAuthService_RefreshToken_UserNotFound(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return nil, nil
		},
	}
	svc := newAuthService(userRepo, &testutil.MockAdminUserRepository{}, &testutil.MockWechatClient{})
	_, err := svc.RefreshToken(context.Background(), 1, 1)
	require.Error(t, err)
}

func TestAuthService_RefreshToken_DBError(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := newAuthService(userRepo, &testutil.MockAdminUserRepository{}, &testutil.MockWechatClient{})
	_, err := svc.RefreshToken(context.Background(), 1, 1)
	require.Error(t, err)
}
