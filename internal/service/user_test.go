package service

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/testutil"
)

func newUserService(
	userRepo *testutil.MockUserRepository,
	adminRepo *testutil.MockAdminUserRepository,
	tagRepo *testutil.MockUserTagRepository,
	roleRepo *testutil.MockRoleRepository,
	permRepo *testutil.MockPermissionRepository,
) UserService {
	return NewUserService(userRepo, adminRepo, tagRepo, roleRepo, permRepo, logrus.New())
}

func newUserServiceWithAttr(
	userRepo *testutil.MockUserRepository,
	adminRepo *testutil.MockAdminUserRepository,
	tagRepo *testutil.MockUserTagRepository,
	roleRepo *testutil.MockRoleRepository,
	permRepo *testutil.MockPermissionRepository,
	attrRepo *testutil.MockAttributeRepository,
	uaRepo *testutil.MockUserAttributeRepository,
) UserService {
	return NewUserService(userRepo, adminRepo, tagRepo, roleRepo, permRepo, logrus.New(), attrRepo, uaRepo)
}

// ==================== GetProfile ====================

func TestUserService_GetProfile_Found(t *testing.T) {
	user := &entity.User{ID: 1, Nickname: "Alice"}
	userRepo := &testutil.MockUserRepository{
		GetWithTagsFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return user, nil
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	got, err := svc.GetProfile(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, user, got)
}

func TestUserService_GetProfile_WithAvatarFileIDAttribute(t *testing.T) {
	user := &entity.User{ID: 1, Nickname: "Alice"}
	userRepo := &testutil.MockUserRepository{
		GetWithTagsFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return user, nil
		},
	}
	avatar := int64(99)
	uaRepo := &testutil.MockUserAttributeRepository{
		ListByUserIDFn: func(_ context.Context, userID uint64) ([]*entity.UserAttribute, error) {
			return []*entity.UserAttribute{
				{
					UserID:      userID,
					AttributeID: 1,
					ValueBigint: &avatar,
					Attribute:   &entity.Attribute{Name: "avatar_file_id", Type: entity.AttributeTypeBigInt},
				},
			}, nil
		},
	}
	svc := newUserServiceWithAttr(userRepo, nil, nil, nil, nil, nil, uaRepo)
	got, err := svc.GetProfile(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, got.AvatarFileID)
	assert.Equal(t, uint64(99), *got.AvatarFileID)
}

func TestUserService_GetProfile_NotFound(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetWithTagsFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return nil, nil
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	_, err := svc.GetProfile(context.Background(), 1)
	require.Error(t, err)
	assert.IsType(t, &apperrors.AppError{}, err)
}

func TestUserService_GetProfile_DBError(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetWithTagsFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	_, err := svc.GetProfile(context.Background(), 1)
	require.Error(t, err)
}

// ==================== UpdateProfile ====================

func TestUserService_UpdateProfile_Success(t *testing.T) {
	user := &entity.User{ID: 1, Nickname: "Alice"}
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return user, nil
		},
		UpdateFn: func(_ context.Context, u *entity.User) error {
			return nil
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	err := svc.UpdateProfile(context.Background(), 1, &dto.UserProfileUpdateRequest{
		Nickname: "Bob", AvatarURL: "http://example.com/a.jpg",
	})
	require.NoError(t, err)
	assert.Equal(t, "Bob", user.Nickname)
}

func TestUserService_UpdateProfile_WithAvatarFileIDAttribute(t *testing.T) {
	user := &entity.User{ID: 1, Nickname: "Alice"}
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return user, nil
		},
		UpdateFn: func(_ context.Context, u *entity.User) error {
			return nil
		},
	}
	attrRepo := &testutil.MockAttributeRepository{
		ListFn: func(_ context.Context) ([]*entity.Attribute, error) {
			return []*entity.Attribute{{ID: 7, Name: "avatar_file_id", Type: entity.AttributeTypeBigInt}}, nil
		},
	}
	called := false
	uaRepo := &testutil.MockUserAttributeRepository{
		UpsertFn: func(_ context.Context, ua *entity.UserAttribute) error {
			called = true
			assert.Equal(t, uint64(1), ua.UserID)
			assert.Equal(t, uint(7), ua.AttributeID)
			require.NotNil(t, ua.ValueBigint)
			assert.Equal(t, int64(123), *ua.ValueBigint)
			return nil
		},
	}
	svc := newUserServiceWithAttr(userRepo, nil, nil, nil, nil, attrRepo, uaRepo)
	err := svc.UpdateProfile(context.Background(), 1, &dto.UserProfileUpdateRequest{AvatarFileID: 123})
	require.NoError(t, err)
	assert.True(t, called)
}

func TestUserService_UpdateProfile_NoChange(t *testing.T) {
	user := &entity.User{ID: 1, Nickname: "Alice", AvatarURL: "old"}
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return user, nil
		},
		UpdateFn: func(_ context.Context, u *entity.User) error {
			return nil
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	err := svc.UpdateProfile(context.Background(), 1, &dto.UserProfileUpdateRequest{})
	require.NoError(t, err)
}

func TestUserService_UpdateProfile_NotFound(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return nil, nil
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	err := svc.UpdateProfile(context.Background(), 1, &dto.UserProfileUpdateRequest{Nickname: "Bob"})
	require.Error(t, err)
}

func TestUserService_UpdateProfile_DBError(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	err := svc.UpdateProfile(context.Background(), 1, &dto.UserProfileUpdateRequest{Nickname: "Bob"})
	require.Error(t, err)
}

// ==================== GetPermissions ====================

func TestUserService_GetPermissions_Success(t *testing.T) {
	roleRepo := &testutil.MockRoleRepository{
		GetUserRolesFn: func(_ context.Context, userID uint64) ([]*entity.Role, error) {
			return []*entity.Role{{Name: "admin"}}, nil
		},
	}
	permRepo := &testutil.MockPermissionRepository{
		GetUserPermissionsFn: func(_ context.Context, userID uint64) ([]*entity.Permission, error) {
			return []*entity.Permission{{Code: "user:list"}}, nil
		},
	}
	svc := newUserService(nil, nil, nil, roleRepo, permRepo)
	roles, perms, err := svc.GetPermissions(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, []string{"admin"}, roles)
	assert.Equal(t, []string{"user:list"}, perms)
}

func TestUserService_GetPermissions_RoleError(t *testing.T) {
	roleRepo := &testutil.MockRoleRepository{
		GetUserRolesFn: func(_ context.Context, userID uint64) ([]*entity.Role, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := newUserService(nil, nil, nil, roleRepo, nil)
	_, _, err := svc.GetPermissions(context.Background(), 1)
	require.Error(t, err)
}

func TestUserService_GetPermissions_PermError(t *testing.T) {
	roleRepo := &testutil.MockRoleRepository{
		GetUserRolesFn: func(_ context.Context, userID uint64) ([]*entity.Role, error) {
			return nil, nil
		},
	}
	permRepo := &testutil.MockPermissionRepository{
		GetUserPermissionsFn: func(_ context.Context, userID uint64) ([]*entity.Permission, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := newUserService(nil, nil, nil, roleRepo, permRepo)
	_, _, err := svc.GetPermissions(context.Background(), 1)
	require.Error(t, err)
}

// ==================== List ====================

func TestUserService_List(t *testing.T) {
	users := []*entity.User{{ID: 1}, {ID: 2}}
	userRepo := &testutil.MockUserRepository{
		ListFn: func(_ context.Context, page, pageSize int, keyword string, userType, status *int8) ([]*entity.User, int64, error) {
			return users, 2, nil
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	got, total, err := svc.List(context.Background(), 1, 10, "", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, got, 2)
}

// ==================== GetByID ====================

func TestUserService_GetByID_Found(t *testing.T) {
	user := &entity.User{ID: 1}
	userRepo := &testutil.MockUserRepository{
		GetWithTagsFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return user, nil
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	got, err := svc.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, user, got)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetWithTagsFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return nil, nil
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	_, err := svc.GetByID(context.Background(), 1)
	require.Error(t, err)
}

// ==================== CreateAdminUser ====================

func TestUserService_CreateAdminUser_Success(t *testing.T) {
	adminRepo := &testutil.MockAdminUserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*entity.AdminUser, error) {
			return nil, nil
		},
		CreateFn: func(_ context.Context, a *entity.AdminUser) error {
			a.ID = 1
			return nil
		},
	}
	userRepo := &testutil.MockUserRepository{
		CreateFn: func(_ context.Context, u *entity.User) error {
			u.ID = 10
			return nil
		},
	}
	svc := newUserService(userRepo, adminRepo, nil, nil, nil)
	id, err := svc.CreateAdminUser(context.Background(), &dto.CreateAdminUserRequest{
		Email:    "a@b.com",
		Password: "password123",
		Nickname: "Admin",
		UserType: 2,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(10), id)
}

func TestUserService_CreateAdminUser_EmailExists(t *testing.T) {
	adminRepo := &testutil.MockAdminUserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*entity.AdminUser, error) {
			return &entity.AdminUser{ID: 1}, nil
		},
	}
	svc := newUserService(nil, adminRepo, nil, nil, nil)
	_, err := svc.CreateAdminUser(context.Background(), &dto.CreateAdminUserRequest{
		Email: "a@b.com", Password: "pw", UserType: 2,
	})
	require.Error(t, err)
	appErr := err.(*apperrors.AppError)
	assert.Equal(t, 409001, appErr.Code)
}

func TestUserService_CreateAdminUser_GetEmailError(t *testing.T) {
	adminRepo := &testutil.MockAdminUserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*entity.AdminUser, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := newUserService(nil, adminRepo, nil, nil, nil)
	_, err := svc.CreateAdminUser(context.Background(), &dto.CreateAdminUserRequest{
		Email: "a@b.com", Password: "pw", UserType: 2,
	})
	require.Error(t, err)
}

func TestUserService_CreateAdminUser_CreateUserError(t *testing.T) {
	adminRepo := &testutil.MockAdminUserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*entity.AdminUser, error) {
			return nil, nil
		},
	}
	userRepo := &testutil.MockUserRepository{
		CreateFn: func(_ context.Context, u *entity.User) error {
			return apperrors.NewInternal("db error", nil)
		},
	}
	svc := newUserService(userRepo, adminRepo, nil, nil, nil)
	_, err := svc.CreateAdminUser(context.Background(), &dto.CreateAdminUserRequest{
		Email: "a@b.com", Password: "pw", UserType: 2,
	})
	require.Error(t, err)
}

func TestUserService_CreateAdminUser_CreateAdminError(t *testing.T) {
	adminRepo := &testutil.MockAdminUserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*entity.AdminUser, error) {
			return nil, nil
		},
		CreateFn: func(_ context.Context, a *entity.AdminUser) error {
			return apperrors.NewInternal("db error", nil)
		},
	}
	userRepo := &testutil.MockUserRepository{
		CreateFn: func(_ context.Context, u *entity.User) error {
			u.ID = 10
			return nil
		},
	}
	svc := newUserService(userRepo, adminRepo, nil, nil, nil)
	_, err := svc.CreateAdminUser(context.Background(), &dto.CreateAdminUserRequest{
		Email: "a@b.com", Password: "pw", UserType: 2,
	})
	require.Error(t, err)
}

// ==================== UpdateUser ====================

func TestUserService_UpdateUser_Success(t *testing.T) {
	user := &entity.User{ID: 1, UserType: 2}
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return user, nil
		},
		UpdateFn: func(_ context.Context, u *entity.User) error {
			return nil
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	err := svc.UpdateUser(context.Background(), 1, &dto.UpdateUserRequest{Nickname: "NewName", Status: 1}, 99)
	require.NoError(t, err)
}

func TestUserService_UpdateUser_SelfTypeChange(t *testing.T) {
	user := &entity.User{ID: 1, UserType: 2}
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return user, nil
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	err := svc.UpdateUser(context.Background(), 1, &dto.UpdateUserRequest{UserType: 3}, 1)
	require.Error(t, err)
	appErr := err.(*apperrors.AppError)
	assert.Equal(t, 403001, appErr.Code)
}

func TestUserService_UpdateUser_NotFound(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return nil, nil
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	err := svc.UpdateUser(context.Background(), 1, &dto.UpdateUserRequest{}, 99)
	require.Error(t, err)
}

func TestUserService_UpdateUser_WithFreezeTime(t *testing.T) {
	user := &entity.User{ID: 1}
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return user, nil
		},
		UpdateFn: func(_ context.Context, u *entity.User) error {
			return nil
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	err := svc.UpdateUser(context.Background(), 1, &dto.UpdateUserRequest{}, 99)
	require.NoError(t, err)
}

// ==================== DeleteUser ====================

func TestUserService_DeleteUser_Success(t *testing.T) {
	user := &entity.User{ID: 1}
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return user, nil
		},
		DeleteFn: func(_ context.Context, id uint64) error {
			return nil
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	err := svc.DeleteUser(context.Background(), 1)
	require.NoError(t, err)
}

func TestUserService_DeleteUser_NotFound(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return nil, nil
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	err := svc.DeleteUser(context.Background(), 1)
	require.Error(t, err)
}

func TestUserService_DeleteUser_WithAssociations(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByIDFn:         func(_ context.Context, id uint64) (*entity.User, error) { return &entity.User{ID: id}, nil },
		HasAssociationsFn: func(_ context.Context, id uint64) (bool, error) { return true, nil },
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	err := svc.DeleteUser(context.Background(), 1)
	require.Error(t, err)
}

// ==================== AssignRoles ====================

func TestUserService_AssignRoles_Success(t *testing.T) {
	user := &entity.User{ID: 1}
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return user, nil
		},
	}
	roleRepo := &testutil.MockRoleRepository{
		AssignUserRolesFn: func(_ context.Context, userID uint64, roleIDs []uint) error {
			return nil
		},
	}
	svc := newUserService(userRepo, nil, nil, roleRepo, nil)
	err := svc.AssignRoles(context.Background(), 1, &dto.AssignRolesRequest{RoleIDs: []uint{1, 2}})
	require.NoError(t, err)
}

func TestUserService_AssignRoles_UserNotFound(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return nil, nil
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	err := svc.AssignRoles(context.Background(), 1, &dto.AssignRolesRequest{RoleIDs: []uint{1}})
	require.Error(t, err)
}

// ==================== AddTag ====================

func TestUserService_AddTag_Success(t *testing.T) {
	tagRepo := &testutil.MockUserTagRepository{
		GetByUserIDFn: func(_ context.Context, userID uint64) ([]*entity.UserTag, error) {
			return []*entity.UserTag{}, nil
		},
		CreateFn: func(_ context.Context, tag *entity.UserTag) error {
			tag.ID = 1
			return nil
		},
	}
	svc := newUserService(nil, nil, tagRepo, nil, nil)
	id, err := svc.AddTag(context.Background(), 1, &dto.AddTagRequest{TagName: "go"})
	require.NoError(t, err)
	assert.Equal(t, uint(1), id)
}

func TestUserService_AddTag_LimitExceeded(t *testing.T) {
	tags := make([]*entity.UserTag, 10)
	for i := range tags {
		tags[i] = &entity.UserTag{ID: uint(i + 1)}
	}
	tagRepo := &testutil.MockUserTagRepository{
		GetByUserIDFn: func(_ context.Context, userID uint64) ([]*entity.UserTag, error) {
			return tags, nil
		},
	}
	svc := newUserService(nil, nil, tagRepo, nil, nil)
	_, err := svc.AddTag(context.Background(), 1, &dto.AddTagRequest{TagName: "go"})
	require.Error(t, err)
	appErr := err.(*apperrors.AppError)
	assert.Equal(t, 400001, appErr.Code)
}

func TestUserService_AddTag_GetTagsError(t *testing.T) {
	tagRepo := &testutil.MockUserTagRepository{
		GetByUserIDFn: func(_ context.Context, userID uint64) ([]*entity.UserTag, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := newUserService(nil, nil, tagRepo, nil, nil)
	_, err := svc.AddTag(context.Background(), 1, &dto.AddTagRequest{TagName: "go"})
	require.Error(t, err)
}

// ==================== DeleteTag ====================

func TestUserService_DeleteTag_Success(t *testing.T) {
	tagRepo := &testutil.MockUserTagRepository{
		DeleteFn: func(_ context.Context, id uint) error {
			return nil
		},
	}
	svc := newUserService(nil, nil, tagRepo, nil, nil)
	err := svc.DeleteTag(context.Background(), 1, 5)
	require.NoError(t, err)
}

// ==================== Missing user service error paths ====================

func TestUserService_GetByID_DBError(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	_, err := svc.GetByID(context.Background(), 1)
	require.Error(t, err)
}

func TestUserService_CreateAdminUser_PasswordTooLong(t *testing.T) {
	// A 73-byte password triggers bcrypt.ErrPasswordTooLong
	longPass := string(make([]byte, 73))
	adminRepo := &testutil.MockAdminUserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*entity.AdminUser, error) {
			return nil, nil
		},
	}
	svc := newUserService(nil, adminRepo, nil, nil, nil)
	_, err := svc.CreateAdminUser(context.Background(), &dto.CreateAdminUserRequest{
		Email: "a@b.com", Password: longPass, Nickname: "N",
	})
	require.Error(t, err)
}

func TestUserService_UpdateUser_ChangeUserType(t *testing.T) {
	// operator != user, req.UserType != 0 => user.UserType should be updated
	user := &entity.User{ID: 1, UserType: 1}
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) { return user, nil },
		UpdateFn:  func(_ context.Context, u *entity.User) error { return nil },
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	err := svc.UpdateUser(context.Background(), 1, &dto.UpdateUserRequest{UserType: 2}, 99)
	require.NoError(t, err)
	assert.Equal(t, int8(2), user.UserType)
}

func TestUserService_UpdateUser_SetFreezeEndTime(t *testing.T) {
	now := time.Now()
	user := &entity.User{ID: 1}
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) { return user, nil },
		UpdateFn:  func(_ context.Context, u *entity.User) error { return nil },
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	err := svc.UpdateUser(context.Background(), 1, &dto.UpdateUserRequest{FreezeEndTime: &now}, 99)
	require.NoError(t, err)
	assert.Equal(t, &now, user.FreezeEndTime)
}

func TestUserService_DeleteUser_DeleteError(t *testing.T) {
	user := &entity.User{ID: 1}
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) { return user, nil },
		DeleteFn:  func(_ context.Context, id uint64) error { return apperrors.NewInternal("db", nil) },
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	err := svc.DeleteUser(context.Background(), 1)
	require.Error(t, err)
}

func TestUserService_AssignRoles_GetByIDError(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := newUserService(userRepo, nil, nil, nil, nil)
	err := svc.AssignRoles(context.Background(), 1, &dto.AssignRolesRequest{RoleIDs: []uint{1}})
	require.Error(t, err)
}

func TestUserService_AddTag_TagError(t *testing.T) {
	tagRepo := &testutil.MockUserTagRepository{
		GetByUserIDFn: func(_ context.Context, uid uint64) ([]*entity.UserTag, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := newUserService(nil, nil, tagRepo, nil, nil)
	_, err := svc.AddTag(context.Background(), 1, &dto.AddTagRequest{TagName: "vip"})
	require.Error(t, err)
}
