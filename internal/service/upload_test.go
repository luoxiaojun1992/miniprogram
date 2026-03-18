package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	var gotHeadPath string
	mockCOS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			gotHeadPath = r.URL.Path
			w.Header().Set("Content-Type", "video/mp4")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockCOS.Close()
	cosClient, err := cosutil.NewClient(mockCOS.URL, mockCOS.URL, "miniapp-test", "", "")
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
	assert.Contains(t, gotHeadPath, "/protected-video/20260317/a.mp4")
}

func TestUploadFileService_GenerateBusinessDownload_MimeMismatchDenied(t *testing.T) {
	mockCOS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Content-Type", "image/png")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockCOS.Close()
	cosClient, err := cosutil.NewClient(mockCOS.URL, mockCOS.URL, "miniapp-test", "", "")
	assert.NoError(t, err)
	repo := &inMemoryFileRepo{
		store: map[uint64]*entity.File{
			7: {ID: 7, Key: "protected-video/20260317/a.mp4", Filename: "a.mp4", Usage: "protected", Category: "video"},
		},
	}
	svc := NewUploadFileService(repo, cosClient, logrus.New())
	resp, err := svc.GenerateBusinessDownload(context.Background(), 7, []string{"video"}, "")
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestUploadFileService_GenerateAdminPresign_Success(t *testing.T) {
	cosClient, err := cosutil.NewClient("http://cos:9000", "http://cos:9000", "miniapp-test", "", "")
	require.NoError(t, err)
	repo := &inMemoryFileRepo{}
	svc := NewUploadFileService(repo, cosClient, logrus.New())

	resp, err := svc.GenerateAdminPresign(context.Background(), 1, "cover.png", "embedded", "600")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotZero(t, resp.FileID)
	assert.Contains(t, resp.Key, "embedded-image/")
	assert.Contains(t, resp.StaticURL, "/miniapp-test/")
}

func TestUploadFileService_GenerateAdminPresign_InvalidUsage(t *testing.T) {
	cosClient, err := cosutil.NewClient("http://cos:9000", "http://cos:9000", "miniapp-test", "", "")
	require.NoError(t, err)
	repo := &inMemoryFileRepo{}
	svc := NewUploadFileService(repo, cosClient, logrus.New())

	_, err = svc.GenerateAdminPresign(context.Background(), 1, "cover.png", "cover", "600")
	assert.Error(t, err)
}

func TestUploadFileService_GenerateAdminPresign_UnsupportedExtension(t *testing.T) {
	cosClient, err := cosutil.NewClient("http://cos:9000", "http://cos:9000", "miniapp-test", "", "")
	require.NoError(t, err)
	repo := &inMemoryFileRepo{}
	svc := NewUploadFileService(repo, cosClient, logrus.New())

	_, err = svc.GenerateAdminPresign(context.Background(), 1, "cover.exe", "embedded", "600")
	assert.Error(t, err)
}

func TestUploadFileService_GenerateStaticURL(t *testing.T) {
	t.Run("file not found", func(t *testing.T) {
		cosClient, err := cosutil.NewClient("http://cos:9000", "http://cos:9000", "miniapp-test", "", "")
		require.NoError(t, err)
		svc := NewUploadFileService(&inMemoryFileRepo{}, cosClient, logrus.New())

		_, err = svc.GenerateStaticURL(context.Background(), 9)
		assert.Error(t, err)
	})

	t.Run("forbidden usage", func(t *testing.T) {
		cosClient, err := cosutil.NewClient("http://cos:9000", "http://cos:9000", "miniapp-test", "", "")
		require.NoError(t, err)
		svc := NewUploadFileService(&inMemoryFileRepo{
			store: map[uint64]*entity.File{
				1: {ID: 1, Key: "protected-image/20260318/x.png", Usage: "protected", Category: "image"},
			},
		}, cosClient, logrus.New())

		_, err = svc.GenerateStaticURL(context.Background(), 1)
		assert.Error(t, err)
	})
}

func TestUploadHelpers(t *testing.T) {
	t.Run("classifyFileCategory", func(t *testing.T) {
		assert.Equal(t, "image", classifyFileCategory(".png"))
		assert.Equal(t, "video", classifyFileCategory(".mp4"))
		assert.Equal(t, "attachment", classifyFileCategory(".pdf"))
		assert.Equal(t, "", classifyFileCategory(".exe"))
	})

	t.Run("parseExpiresIn", func(t *testing.T) {
		v, err := parseExpiresIn("", 900)
		require.NoError(t, err)
		assert.Equal(t, 900, v)
		_, err = parseExpiresIn("59", 900)
		assert.Error(t, err)
		v, err = parseExpiresIn("600", 900)
		require.NoError(t, err)
		assert.Equal(t, 600, v)
	})

	t.Run("containsCategory", func(t *testing.T) {
		assert.True(t, containsCategory([]string{" image ", "video"}, "image"))
		assert.False(t, containsCategory([]string{"video"}, "attachment"))
	})

	t.Run("resolveFileExtension", func(t *testing.T) {
		assert.Equal(t, ".pdf", resolveFileExtension("a.pdf", "x/y/z.txt"))
		assert.Equal(t, ".txt", resolveFileExtension("", "x/y/z.txt"))
	})

	t.Run("attachmentContentTypeAllowed", func(t *testing.T) {
		assert.True(t, attachmentContentTypeAllowed(".pdf", "application/pdf"))
		assert.True(t, attachmentContentTypeAllowed(".pdf", "application/octet-stream"))
		assert.False(t, attachmentContentTypeAllowed(".pdf", "image/png"))
	})

	t.Run("contentTypeMatchesCategory", func(t *testing.T) {
		assert.True(t, contentTypeMatchesCategory("image", "image/png", "", ""))
		assert.True(t, contentTypeMatchesCategory("video", "video/mp4", "", ""))
		assert.True(t, contentTypeMatchesCategory("attachment", "application/pdf", "a.pdf", ""))
		assert.False(t, contentTypeMatchesCategory("attachment", "image/png", "a.pdf", ""))
		assert.False(t, contentTypeMatchesCategory("unknown", "image/png", "", ""))
	})
}
