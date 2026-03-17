package service

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/cosutil"
)

type inMemoryFileRepo struct {
	seq   uint64
	store map[uint64]*entity.File
}

func (r *inMemoryFileRepo) GetByID(_ context.Context, id uint64) (*entity.File, error) {
	if r.store == nil {
		return nil, nil
	}
	return r.store[id], nil
}

func (r *inMemoryFileRepo) Create(_ context.Context, file *entity.File) error {
	r.seq++
	file.ID = r.seq
	if r.store == nil {
		r.store = map[uint64]*entity.File{}
	}
	r.store[file.ID] = file
	return nil
}

func TestUploadFileService_GenerateProtectedBusinessPresign_OnlyMedia(t *testing.T) {
	cosClient, err := cosutil.NewClient("http://cos:9000", "http://cos:9000", "miniapp-test", "", "")
	assert.NoError(t, err)
	repo := &inMemoryFileRepo{}
	svc := NewUploadFileService(repo, cosClient, logrus.New())

	_, err = svc.GenerateProtectedBusinessPresign(context.Background(), 1, "x.pdf", "banner_media", "600", []string{"image", "video"})
	assert.Error(t, err)

	resp, err := svc.GenerateProtectedBusinessPresign(context.Background(), 1, "x.png", "banner_media", "600", []string{"image", "video"})
	assert.NoError(t, err)
	assert.NotZero(t, resp.FileID)
	assert.Contains(t, resp.Key, "protected-image/")
}

func TestUploadFileService_GenerateBusinessDownload_AllowsMultipleCategories(t *testing.T) {
	cosClient, err := cosutil.NewClient("http://cos:9000", "http://cos:9000", "miniapp-test", "", "")
	assert.NoError(t, err)
	repo := &inMemoryFileRepo{
		store: map[uint64]*entity.File{
			7: {ID: 7, Key: "protected-video/20260317/a.mp4", Usage: "protected", Category: "video"},
		},
	}
	svc := NewUploadFileService(repo, cosClient, logrus.New())
	resp, err := svc.GenerateBusinessDownload(context.Background(), 7, []string{"image", "video"}, "")
	assert.NoError(t, err)
	assert.Equal(t, uint64(7), resp.FileID)
	assert.Contains(t, resp.Download, "protected-video/20260317/a.mp4")
}
