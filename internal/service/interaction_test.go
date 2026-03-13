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

// ==================== StudyRecordService ====================

func TestStudyRecordService_List(t *testing.T) {
	records := []*entity.UserStudyRecord{{ID: 1, UserID: 1}}
	repo := &testutil.MockStudyRecordRepository{
		ListByUserFn: func(_ context.Context, userID uint64, page, pageSize int) ([]*entity.UserStudyRecord, int64, error) {
			return records, 1, nil
		},
	}
	svc := NewStudyRecordService(repo, logrus.New())
	got, total, err := svc.List(context.Background(), 1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, records, got)
}

func TestStudyRecordService_Update(t *testing.T) {
	repo := &testutil.MockStudyRecordRepository{
		UpsertFn: func(_ context.Context, record *entity.UserStudyRecord) error {
			return nil
		},
	}
	svc := NewStudyRecordService(repo, logrus.New())
	err := svc.Update(context.Background(), 1, &dto.UpdateStudyRecordRequest{
		UnitID: 1, Progress: 50, Status: 1,
	})
	require.NoError(t, err)
}

// ==================== CollectionService ====================

func TestCollectionService_List(t *testing.T) {
	cols := []*entity.Collection{{ID: 1, UserID: 1}}
	repo := &testutil.MockCollectionRepository{
		ListFn: func(_ context.Context, userID uint64, page, pageSize int, contentType *int8) ([]*entity.Collection, int64, error) {
			return cols, 1, nil
		},
	}
	svc := NewCollectionService(repo, logrus.New())
	got, total, err := svc.List(context.Background(), 1, 1, 10, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, cols, got)
}

func TestCollectionService_Add_Success(t *testing.T) {
	repo := &testutil.MockCollectionRepository{
		GetFn: func(_ context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Collection, error) {
			return nil, nil
		},
		CreateFn: func(_ context.Context, c *entity.Collection) error {
			return nil
		},
	}
	svc := NewCollectionService(repo, logrus.New())
	err := svc.Add(context.Background(), 1, 1, 1)
	require.NoError(t, err)
}

func TestCollectionService_Add_AlreadyExists(t *testing.T) {
	repo := &testutil.MockCollectionRepository{
		GetFn: func(_ context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Collection, error) {
			return &entity.Collection{ID: 1}, nil
		},
	}
	svc := NewCollectionService(repo, logrus.New())
	err := svc.Add(context.Background(), 1, 1, 1)
	require.Error(t, err)
	appErr := err.(*apperrors.AppError)
	assert.Equal(t, 409001, appErr.Code)
}

func TestCollectionService_Add_GetError(t *testing.T) {
	repo := &testutil.MockCollectionRepository{
		GetFn: func(_ context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Collection, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := NewCollectionService(repo, logrus.New())
	err := svc.Add(context.Background(), 1, 1, 1)
	require.Error(t, err)
}

func TestCollectionService_Remove(t *testing.T) {
	repo := &testutil.MockCollectionRepository{
		DeleteFn: func(_ context.Context, userID uint64, contentType int8, contentID uint64) error {
			return nil
		},
	}
	svc := NewCollectionService(repo, logrus.New())
	err := svc.Remove(context.Background(), 1, 1, 1)
	require.NoError(t, err)
}

// ==================== LikeService ====================

func TestLikeService_Add_Success(t *testing.T) {
	likeRepo := &testutil.MockLikeRepository{
		GetFn: func(_ context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Like, error) {
			return nil, nil
		},
		CreateFn: func(_ context.Context, l *entity.Like) error {
			return nil
		},
	}
	svc := NewLikeService(likeRepo, nil, nil, logrus.New())
	err := svc.Add(context.Background(), 1, 1, 1)
	require.NoError(t, err)
}

func TestLikeService_Add_AlreadyExists(t *testing.T) {
	likeRepo := &testutil.MockLikeRepository{
		GetFn: func(_ context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Like, error) {
			return &entity.Like{ID: 1}, nil
		},
	}
	svc := NewLikeService(likeRepo, nil, nil, logrus.New())
	err := svc.Add(context.Background(), 1, 1, 1)
	require.Error(t, err)
	appErr := err.(*apperrors.AppError)
	assert.Equal(t, 409001, appErr.Code)
}

func TestLikeService_Add_GetError(t *testing.T) {
	likeRepo := &testutil.MockLikeRepository{
		GetFn: func(_ context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Like, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := NewLikeService(likeRepo, nil, nil, logrus.New())
	err := svc.Add(context.Background(), 1, 1, 1)
	require.Error(t, err)
}

func TestLikeService_Remove(t *testing.T) {
	likeRepo := &testutil.MockLikeRepository{
		DeleteFn: func(_ context.Context, userID uint64, contentType int8, contentID uint64) error {
			return nil
		},
	}
	svc := NewLikeService(likeRepo, nil, nil, logrus.New())
	err := svc.Remove(context.Background(), 1, 1, 1)
	require.NoError(t, err)
}

// ==================== NotificationService ====================

func TestNotificationService_List_Success(t *testing.T) {
	notifs := []*entity.Notification{{ID: 1}}
	repo := &testutil.MockNotificationRepository{
		ListFn: func(_ context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, error) {
			return notifs, 1, nil
		},
		UnreadCountFn: func(_ context.Context, userID uint64) (int64, error) {
			return 3, nil
		},
	}
	svc := NewNotificationService(repo, logrus.New())
	got, total, unread, err := svc.List(context.Background(), 1, 1, 10, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, int64(3), unread)
	assert.Equal(t, notifs, got)
}

func TestNotificationService_List_ListError(t *testing.T) {
	repo := &testutil.MockNotificationRepository{
		ListFn: func(_ context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, error) {
			return nil, 0, apperrors.NewInternal("db error", nil)
		},
	}
	svc := NewNotificationService(repo, logrus.New())
	_, _, _, err := svc.List(context.Background(), 1, 1, 10, nil)
	require.Error(t, err)
}

func TestNotificationService_List_UnreadError(t *testing.T) {
	repo := &testutil.MockNotificationRepository{
		ListFn: func(_ context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, error) {
			return []*entity.Notification{}, 0, nil
		},
		UnreadCountFn: func(_ context.Context, userID uint64) (int64, error) {
			return 0, apperrors.NewInternal("db error", nil)
		},
	}
	svc := NewNotificationService(repo, logrus.New())
	_, _, _, err := svc.List(context.Background(), 1, 1, 10, nil)
	require.Error(t, err)
}

func TestNotificationService_MarkRead(t *testing.T) {
	repo := &testutil.MockNotificationRepository{
		MarkReadFn: func(_ context.Context, id uint64) error {
			return nil
		},
	}
	svc := NewNotificationService(repo, logrus.New())
	err := svc.MarkRead(context.Background(), 1)
	require.NoError(t, err)
}

func TestNotificationService_MarkAllRead(t *testing.T) {
	repo := &testutil.MockNotificationRepository{
		MarkAllReadFn: func(_ context.Context, userID uint64) error {
			return nil
		},
	}
	svc := NewNotificationService(repo, logrus.New())
	err := svc.MarkAllRead(context.Background(), 1)
	require.NoError(t, err)
}

// ==================== WechatConfigService ====================

func TestWechatConfigService_Get_Found(t *testing.T) {
	cfg := &entity.WechatConfig{AppID: "wx123"}
	repo := &testutil.MockWechatConfigRepository{
		GetFn: func(_ context.Context) (*entity.WechatConfig, error) {
			return cfg, nil
		},
	}
	svc := NewWechatConfigService(repo, logrus.New())
	got, err := svc.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, cfg, got)
}

func TestWechatConfigService_Get_NotFound(t *testing.T) {
	repo := &testutil.MockWechatConfigRepository{
		GetFn: func(_ context.Context) (*entity.WechatConfig, error) {
			return nil, nil
		},
	}
	svc := NewWechatConfigService(repo, logrus.New())
	got, err := svc.Get(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, got)
}

func TestWechatConfigService_Get_Error(t *testing.T) {
	repo := &testutil.MockWechatConfigRepository{
		GetFn: func(_ context.Context) (*entity.WechatConfig, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := NewWechatConfigService(repo, logrus.New())
	_, err := svc.Get(context.Background())
	require.Error(t, err)
}

func TestWechatConfigService_Update_Found(t *testing.T) {
	cfg := &entity.WechatConfig{AppID: "wx123"}
	repo := &testutil.MockWechatConfigRepository{
		GetFn: func(_ context.Context) (*entity.WechatConfig, error) {
			return cfg, nil
		},
		UpdateFn: func(_ context.Context, c *entity.WechatConfig) error {
			return nil
		},
	}
	svc := NewWechatConfigService(repo, logrus.New())
	err := svc.Update(context.Background(), &dto.UpdateWechatConfigRequest{
		AppID: "new_app_id", AppSecret: "secret", APIToken: "token",
	})
	require.NoError(t, err)
}

func TestWechatConfigService_Update_NotFound(t *testing.T) {
	repo := &testutil.MockWechatConfigRepository{
		GetFn: func(_ context.Context) (*entity.WechatConfig, error) {
			return nil, nil
		},
		UpdateFn: func(_ context.Context, c *entity.WechatConfig) error {
			return nil
		},
	}
	svc := NewWechatConfigService(repo, logrus.New())
	err := svc.Update(context.Background(), &dto.UpdateWechatConfigRequest{
		AppID: "new_app_id", AppSecret: "secret",
	})
	require.NoError(t, err)
}

func TestWechatConfigService_Update_GetError(t *testing.T) {
	repo := &testutil.MockWechatConfigRepository{
		GetFn: func(_ context.Context) (*entity.WechatConfig, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := NewWechatConfigService(repo, logrus.New())
	err := svc.Update(context.Background(), &dto.UpdateWechatConfigRequest{})
	require.Error(t, err)
}

// ==================== AuditLogService ====================

func TestAuditLogService_List(t *testing.T) {
	logs := []*entity.AuditLog{{ID: 1}}
	repo := &testutil.MockAuditLogRepository{
		ListFn: func(_ context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error) {
			return logs, 1, nil
		},
	}
	svc := NewAuditLogService(repo, logrus.New())
	got, total, err := svc.List(context.Background(), 1, 10, "", "", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, logs, got)
}

func TestAuditLogService_Log_Success(t *testing.T) {
	repo := &testutil.MockAuditLogRepository{
		CreateFn: func(_ context.Context, l *entity.AuditLog) error {
			return nil
		},
	}
	svc := NewAuditLogService(repo, logrus.New())
	svc.Log(context.Background(), &entity.AuditLog{Module: "user", Action: "create"})
}

func TestAuditLogService_Log_Error(t *testing.T) {
	repo := &testutil.MockAuditLogRepository{
		CreateFn: func(_ context.Context, l *entity.AuditLog) error {
			return apperrors.NewInternal("db error", nil)
		},
	}
	svc := NewAuditLogService(repo, logrus.New())
	// Should not panic, just logs warning
	svc.Log(context.Background(), &entity.AuditLog{Module: "user", Action: "create"})
}

// ==================== LogConfigService ====================

func TestLogConfigService_Get_Found(t *testing.T) {
	cfg := &entity.LogConfig{RetentionDays: 30}
	repo := &testutil.MockLogConfigRepository{
		GetFn: func(_ context.Context) (*entity.LogConfig, error) {
			return cfg, nil
		},
	}
	svc := NewLogConfigService(repo, logrus.New())
	got, err := svc.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, cfg, got)
}

func TestLogConfigService_Get_NotFound(t *testing.T) {
	repo := &testutil.MockLogConfigRepository{
		GetFn: func(_ context.Context) (*entity.LogConfig, error) {
			return nil, nil
		},
	}
	svc := NewLogConfigService(repo, logrus.New())
	got, err := svc.Get(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 90, got.RetentionDays)
}

func TestLogConfigService_Get_Error(t *testing.T) {
	repo := &testutil.MockLogConfigRepository{
		GetFn: func(_ context.Context) (*entity.LogConfig, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := NewLogConfigService(repo, logrus.New())
	_, err := svc.Get(context.Background())
	require.Error(t, err)
}

func TestLogConfigService_Update_Found(t *testing.T) {
	cfg := &entity.LogConfig{RetentionDays: 30}
	repo := &testutil.MockLogConfigRepository{
		GetFn: func(_ context.Context) (*entity.LogConfig, error) {
			return cfg, nil
		},
		UpdateFn: func(_ context.Context, c *entity.LogConfig) error {
			return nil
		},
	}
	svc := NewLogConfigService(repo, logrus.New())
	err := svc.Update(context.Background(), &dto.UpdateLogConfigRequest{RetentionDays: 60})
	require.NoError(t, err)
}

func TestLogConfigService_Update_NotFound(t *testing.T) {
	repo := &testutil.MockLogConfigRepository{
		GetFn: func(_ context.Context) (*entity.LogConfig, error) {
			return nil, nil
		},
		UpdateFn: func(_ context.Context, c *entity.LogConfig) error {
			return nil
		},
	}
	svc := NewLogConfigService(repo, logrus.New())
	err := svc.Update(context.Background(), &dto.UpdateLogConfigRequest{RetentionDays: 60})
	require.NoError(t, err)
}

func TestLogConfigService_Update_GetError(t *testing.T) {
	repo := &testutil.MockLogConfigRepository{
		GetFn: func(_ context.Context) (*entity.LogConfig, error) {
			return nil, apperrors.NewInternal("db error", nil)
		},
	}
	svc := NewLogConfigService(repo, logrus.New())
	err := svc.Update(context.Background(), &dto.UpdateLogConfigRequest{})
	require.Error(t, err)
}
