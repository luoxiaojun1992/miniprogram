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

func newAttrService(
	attrRepo *testutil.MockAttributeRepository,
	uaRepo *testutil.MockUserAttributeRepository,
	userRepo *testutil.MockUserRepository,
) AttributeService {
	return NewAttributeService(attrRepo, uaRepo, userRepo, logrus.New())
}

// ==================== List ====================

func TestAttrService_List(t *testing.T) {
	repo := &testutil.MockAttributeRepository{
		ListFn: func(_ context.Context) ([]*entity.Attribute, error) {
			return []*entity.Attribute{{ID: 1, Name: "性别"}}, nil
		},
	}
	svc := newAttrService(repo, nil, nil)
	attrs, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, attrs, 1)
}

func TestAttrService_List_Error(t *testing.T) {
	repo := &testutil.MockAttributeRepository{
		ListFn: func(_ context.Context) ([]*entity.Attribute, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := newAttrService(repo, nil, nil)
	_, err := svc.List(context.Background())
	require.Error(t, err)
}

// ==================== Create ====================

func TestAttrService_Create_OK(t *testing.T) {
	repo := &testutil.MockAttributeRepository{
		CreateFn: func(_ context.Context, attr *entity.Attribute) error {
			attr.ID = 1
			return nil
		},
	}
	svc := newAttrService(repo, nil, nil)
	id, err := svc.Create(context.Background(), &dto.CreateAttributeRequest{Name: "性别", Type: entity.AttributeTypeString})
	require.NoError(t, err)
	assert.Equal(t, uint(1), id)
}

func TestAttrService_Create_Error(t *testing.T) {
	repo := &testutil.MockAttributeRepository{
		CreateFn: func(_ context.Context, attr *entity.Attribute) error {
			return apperrors.NewInternal("db", nil)
		},
	}
	svc := newAttrService(repo, nil, nil)
	_, err := svc.Create(context.Background(), &dto.CreateAttributeRequest{Name: "性别", Type: entity.AttributeTypeString})
	require.Error(t, err)
}

// ==================== Update ====================

func TestAttrService_Update_OK(t *testing.T) {
	repo := &testutil.MockAttributeRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Attribute, error) {
			return &entity.Attribute{ID: id, Name: "旧名"}, nil
		},
		UpdateFn: func(_ context.Context, attr *entity.Attribute) error { return nil },
	}
	svc := newAttrService(repo, nil, nil)
	err := svc.Update(context.Background(), 1, &dto.UpdateAttributeRequest{Name: "新名", Type: entity.AttributeTypeString})
	require.NoError(t, err)
}

func TestAttrService_Update_NotFound(t *testing.T) {
	repo := &testutil.MockAttributeRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Attribute, error) { return nil, nil },
	}
	svc := newAttrService(repo, nil, nil)
	err := svc.Update(context.Background(), 999, &dto.UpdateAttributeRequest{Name: "新名", Type: entity.AttributeTypeString})
	require.Error(t, err)
	assert.IsType(t, &apperrors.AppError{}, err)
}

// ==================== Delete ====================

func TestAttrService_Delete_OK(t *testing.T) {
	repo := &testutil.MockAttributeRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Attribute, error) {
			return &entity.Attribute{ID: id, Name: "性别"}, nil
		},
		DeleteFn: func(_ context.Context, id uint) error { return nil },
	}
	svc := newAttrService(repo, nil, nil)
	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestAttrService_Delete_NotFound(t *testing.T) {
	repo := &testutil.MockAttributeRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Attribute, error) { return nil, nil },
	}
	svc := newAttrService(repo, nil, nil)
	err := svc.Delete(context.Background(), 999)
	require.Error(t, err)
}

func TestAttrService_Delete_WithUserAssociations(t *testing.T) {
	repo := &testutil.MockAttributeRepository{
		GetByIDFn:             func(_ context.Context, id uint) (*entity.Attribute, error) { return &entity.Attribute{ID: id}, nil },
		HasUserAssociationsFn: func(_ context.Context, id uint) (bool, error) { return true, nil },
	}
	svc := newAttrService(repo, nil, nil)
	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
}

// ==================== ListUserAttributes ====================

func TestAttrService_ListUserAttributes_OK(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return &entity.User{ID: id}, nil
		},
	}
	uaRepo := &testutil.MockUserAttributeRepository{
		ListByUserIDFn: func(_ context.Context, userID uint64) ([]*entity.UserAttribute, error) {
			return []*entity.UserAttribute{{ID: 1, UserID: userID, AttributeID: 1, ValueString: "男"}}, nil
		},
	}
	svc := newAttrService(nil, uaRepo, userRepo)
	uas, err := svc.ListUserAttributes(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, uas, 1)
}

func TestAttrService_ListUserAttributes_UserNotFound(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) { return nil, nil },
	}
	svc := newAttrService(nil, nil, userRepo)
	_, err := svc.ListUserAttributes(context.Background(), 999)
	require.Error(t, err)
}

// ==================== SetUserAttribute ====================

func TestAttrService_SetUserAttribute_OK(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return &entity.User{ID: id}, nil
		},
	}
	attrRepo := &testutil.MockAttributeRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Attribute, error) {
			return &entity.Attribute{ID: id, Name: "性别", Type: entity.AttributeTypeString}, nil
		},
	}
	uaRepo := &testutil.MockUserAttributeRepository{
		UpsertFn: func(_ context.Context, ua *entity.UserAttribute) error { return nil },
	}
	svc := newAttrService(attrRepo, uaRepo, userRepo)
	err := svc.SetUserAttribute(context.Background(), 1, &dto.SetUserAttributeRequest{AttributeID: 1, ValueString: "男"})
	require.NoError(t, err)
}

func TestAttrService_SetUserAttribute_UserNotFound(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) { return nil, nil },
	}
	svc := newAttrService(nil, nil, userRepo)
	err := svc.SetUserAttribute(context.Background(), 999, &dto.SetUserAttributeRequest{AttributeID: 1, ValueString: "男"})
	require.Error(t, err)
}

func TestAttrService_SetUserAttribute_AttrNotFound(t *testing.T) {
	userRepo := &testutil.MockUserRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.User, error) {
			return &entity.User{ID: id}, nil
		},
	}
	attrRepo := &testutil.MockAttributeRepository{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Attribute, error) { return nil, nil },
	}
	svc := newAttrService(attrRepo, nil, userRepo)
	err := svc.SetUserAttribute(context.Background(), 1, &dto.SetUserAttributeRequest{AttributeID: 999, ValueString: "男"})
	require.Error(t, err)
}

// ==================== DeleteUserAttribute ====================

func TestAttrService_DeleteUserAttribute_OK(t *testing.T) {
	uaRepo := &testutil.MockUserAttributeRepository{
		DeleteFn: func(_ context.Context, userID uint64, attributeID uint) error { return nil },
	}
	svc := newAttrService(nil, uaRepo, nil)
	err := svc.DeleteUserAttribute(context.Background(), 1, 1)
	require.NoError(t, err)
}

func TestAttrService_DeleteUserAttribute_Error(t *testing.T) {
	uaRepo := &testutil.MockUserAttributeRepository{
		DeleteFn: func(_ context.Context, userID uint64, attributeID uint) error {
			return apperrors.NewInternal("db", nil)
		},
	}
	svc := newAttrService(nil, uaRepo, nil)
	err := svc.DeleteUserAttribute(context.Background(), 1, 1)
	require.Error(t, err)
}
