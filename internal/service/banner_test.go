package service

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/testutil"
)

func TestBannerService_UserAndAdminList(t *testing.T) {
	repo := &testutil.MockBannerRepository{
		ListFn: func(_ context.Context, status *int8) ([]*entity.Banner, error) {
			return []*entity.Banner{{ID: 1}}, nil
		},
	}
	svc := NewBannerService(repo, logrus.New())
	_, err := svc.List(context.Background())
	require.NoError(t, err)
	_, err = svc.AdminList(context.Background(), nil)
	require.NoError(t, err)
}

func TestBannerService_CreateUpdateDelete(t *testing.T) {
	repo := &testutil.MockBannerRepository{
		CreateFn: func(_ context.Context, b *entity.Banner) error {
			b.ID = 1
			return nil
		},
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Banner, error) {
			return &entity.Banner{ID: id}, nil
		},
		UpdateFn: func(_ context.Context, b *entity.Banner) error { return nil },
		DeleteFn: func(_ context.Context, id uint64) error { return nil },
	}
	fileRepo := &testutil.MockFileRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.File, error) {
			return &entity.File{ID: id, Usage: "protected", Category: "image"}, nil
		},
	}
	svc := NewBannerService(repo, logrus.New(), fileRepo)
	_, err := svc.Create(context.Background(), &dto.CreateBannerRequest{Title: "t", ImageFileID: 10, Status: 1})
	require.NoError(t, err)
	err = svc.Update(context.Background(), 1, &dto.CreateBannerRequest{Title: "n", ImageFileID: 10, Status: 1})
	require.NoError(t, err)
	err = svc.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestBannerService_ValidateMediaErrors(t *testing.T) {
	repo := &testutil.MockBannerRepository{}
	fileRepo := &testutil.MockFileRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.File, error) {
			return &entity.File{ID: id, Usage: "static", Category: "text"}, nil
		},
	}
	svc := NewBannerService(repo, logrus.New(), fileRepo)
	_, err := svc.Create(context.Background(), &dto.CreateBannerRequest{Title: "t", ImageFileID: 0})
	require.Error(t, err)
	_, err = svc.Create(context.Background(), &dto.CreateBannerRequest{Title: "t", ImageFileID: 10})
	require.Error(t, err)
}

func TestBannerService_UpdateDeleteNotFound(t *testing.T) {
	repo := &testutil.MockBannerRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Banner, error) { return nil, nil },
	}
	svc := NewBannerService(repo, logrus.New())
	err := svc.Update(context.Background(), 1, &dto.CreateBannerRequest{})
	require.Error(t, err)
	err = svc.Delete(context.Background(), 1)
	require.Error(t, err)
}

func TestBannerService_CreateErrorPropagates(t *testing.T) {
	repo := &testutil.MockBannerRepository{
		CreateFn: func(_ context.Context, b *entity.Banner) error { return apperrors.NewInternal("db", nil) },
	}
	svc := NewBannerService(repo, logrus.New())
	_, err := svc.Create(context.Background(), &dto.CreateBannerRequest{Title: "t", ImageFileID: 1})
	require.Error(t, err)
}
