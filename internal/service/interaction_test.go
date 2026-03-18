package service

import (
	"context"
	"errors"
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
	svc := NewStudyRecordService(repo, nil, nil, logrus.New())
	got, total, err := svc.List(context.Background(), 1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, records, got)
}

func TestStudyRecordService_Update(t *testing.T) {
	unitRepo := &testutil.MockCourseUnitRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.CourseUnit, error) {
			return &entity.CourseUnit{ID: id, CourseID: 10}, nil
		},
	}
	repo := &testutil.MockStudyRecordRepository{
		UpsertFn: func(_ context.Context, record *entity.UserStudyRecord) error {
			return nil
		},
	}
	svc := NewStudyRecordService(repo, unitRepo, nil, logrus.New())
	err := svc.Update(context.Background(), 1, &dto.UpdateStudyRecordRequest{
		UnitID: 1, Progress: 50, Status: 1,
	})
	require.NoError(t, err)
}

func TestStudyRecordService_Update_FirstStudyIncrementsCourseCount(t *testing.T) {
	unitRepo := &testutil.MockCourseUnitRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.CourseUnit, error) {
			return &entity.CourseUnit{ID: id, CourseID: 10}, nil
		},
	}
	studyRepo := &testutil.MockStudyRecordRepository{
		GetByUserAndUnitFn: func(_ context.Context, userID, unitID uint64) (*entity.UserStudyRecord, error) {
			return nil, nil
		},
		UpsertFn: func(_ context.Context, record *entity.UserStudyRecord) error {
			return nil
		},
	}
	courseRepo := &testutil.MockCourseRepository{
		IncrStudyCountFn: func(_ context.Context, id uint64) error {
			assert.Equal(t, uint64(10), id)
			return nil
		},
	}
	svc := NewStudyRecordService(studyRepo, unitRepo, courseRepo, logrus.New())
	err := svc.Update(context.Background(), 1, &dto.UpdateStudyRecordRequest{
		UnitID: 1, Progress: 30, Status: 1,
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
	svc := NewCollectionService(repo, nil, nil, logrus.New())
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
	svc := NewCollectionService(repo, nil, nil, logrus.New())
	err := svc.Add(context.Background(), 1, 1, 1)
	require.NoError(t, err)
}

func TestCollectionService_Add_AlreadyExists(t *testing.T) {
	repo := &testutil.MockCollectionRepository{
		GetFn: func(_ context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Collection, error) {
			return &entity.Collection{ID: 1}, nil
		},
	}
	svc := NewCollectionService(repo, nil, nil, logrus.New())
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
	svc := NewCollectionService(repo, nil, nil, logrus.New())
	err := svc.Add(context.Background(), 1, 1, 1)
	require.Error(t, err)
}

func TestCollectionService_Remove(t *testing.T) {
	repo := &testutil.MockCollectionRepository{
		DeleteFn: func(_ context.Context, userID uint64, contentType int8, contentID uint64) error {
			return nil
		},
	}
	svc := NewCollectionService(repo, nil, nil, logrus.New())
	err := svc.Remove(context.Background(), 1, 1, 1)
	require.NoError(t, err)
}

func TestCollectionService_Add_IncrArticleCollectCount(t *testing.T) {
	repo := &testutil.MockCollectionRepository{
		GetFn: func(_ context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Collection, error) {
			return nil, nil
		},
		CreateFn: func(_ context.Context, c *entity.Collection) error { return nil },
	}
	articleRepo := &testutil.MockArticleRepository{
		IncrCollectCountFn: func(_ context.Context, id uint64) error {
			assert.Equal(t, uint64(8), id)
			return nil
		},
	}
	svc := NewCollectionService(repo, articleRepo, nil, logrus.New())
	require.NoError(t, svc.Add(context.Background(), 1, 1, 8))
}

func TestCollectionService_Remove_DecrCourseCollectCount(t *testing.T) {
	repo := &testutil.MockCollectionRepository{
		DeleteFn: func(_ context.Context, userID uint64, contentType int8, contentID uint64) error { return nil },
	}
	courseRepo := &testutil.MockCourseRepository{
		DecrCollectCountFn: func(_ context.Context, id uint64) error {
			assert.Equal(t, uint64(11), id)
			return nil
		},
	}
	svc := NewCollectionService(repo, nil, courseRepo, logrus.New())
	require.NoError(t, svc.Remove(context.Background(), 1, 2, 11))
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
	svc := NewLikeService(likeRepo, nil, nil, nil, logrus.New())
	err := svc.Add(context.Background(), 1, 1, 1)
	require.NoError(t, err)
}

func TestLikeService_Add_AlreadyExists(t *testing.T) {
	likeRepo := &testutil.MockLikeRepository{
		GetFn: func(_ context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Like, error) {
			return &entity.Like{ID: 1}, nil
		},
	}
	svc := NewLikeService(likeRepo, nil, nil, nil, logrus.New())
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
	svc := NewLikeService(likeRepo, nil, nil, nil, logrus.New())
	err := svc.Add(context.Background(), 1, 1, 1)
	require.Error(t, err)
}

func TestLikeService_Remove(t *testing.T) {
	likeRepo := &testutil.MockLikeRepository{
		DeleteFn: func(_ context.Context, userID uint64, contentType int8, contentID uint64) error {
			return nil
		},
	}
	svc := NewLikeService(likeRepo, nil, nil, nil, logrus.New())
	err := svc.Remove(context.Background(), 1, 1, 1)
	require.NoError(t, err)
}

func TestLikeService_Add_SendsNotificationAndIncrArticleLikeCount(t *testing.T) {
	likeRepo := &testutil.MockLikeRepository{
		GetFn: func(_ context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Like, error) {
			return nil, nil
		},
		CreateFn: func(_ context.Context, like *entity.Like) error { return nil },
	}
	articleRepo := &testutil.MockArticleRepository{
		IncrLikeCountFn: func(_ context.Context, id uint64) error {
			assert.Equal(t, uint64(9), id)
			return nil
		},
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return &entity.Article{ID: id, AuthorID: 2}, nil
		},
	}
	notifRepo := &testutil.MockNotificationRepository{
		CreateFn: func(_ context.Context, n *entity.Notification) error {
			require.NotNil(t, n.UserID)
			assert.Equal(t, uint64(2), *n.UserID)
			assert.Equal(t, int8(4), n.Type)
			return nil
		},
	}
	svc := NewLikeService(likeRepo, articleRepo, nil, notifRepo, logrus.New())
	require.NoError(t, svc.Add(context.Background(), 1, 1, 9))
}

func TestLikeService_Remove_DecrCourseLikeCount(t *testing.T) {
	likeRepo := &testutil.MockLikeRepository{
		DeleteFn: func(_ context.Context, userID uint64, contentType int8, contentID uint64) error { return nil },
	}
	courseRepo := &testutil.MockCourseRepository{
		DecrLikeCountFn: func(_ context.Context, id uint64) error {
			assert.Equal(t, uint64(7), id)
			return nil
		},
	}
	svc := NewLikeService(likeRepo, nil, courseRepo, nil, logrus.New())
	require.NoError(t, svc.Remove(context.Background(), 1, 2, 7))
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

func TestNotificationService_Send(t *testing.T) {
	repo := &testutil.MockNotificationRepository{
		CreateFn: func(_ context.Context, n *entity.Notification) error {
			assert.Equal(t, int8(2), n.Type)
			return nil
		},
	}
	svc := NewNotificationService(repo, logrus.New())
	target := uint64(2)
	err := svc.Send(context.Background(), &entity.Notification{
		UserID:  &target,
		Type:    2,
		Title:   "t",
		Content: "c",
	})
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

func TestCollectionService_Add_CreateError_Extra(t *testing.T) {
	repo := &testutil.MockCollectionRepository{
		GetFn: func(_ context.Context, _ uint64, _ int8, _ uint64) (*entity.Collection, error) { return nil, nil },
		CreateFn: func(_ context.Context, _ *entity.Collection) error {
			return errors.New("create failed")
		},
	}
	svc := NewCollectionService(repo, nil, nil, logrus.New())
	require.Error(t, svc.Add(context.Background(), 1, 1, 1))
}

func TestCollectionService_Remove_DeleteError_Extra(t *testing.T) {
	repo := &testutil.MockCollectionRepository{
		DeleteFn: func(_ context.Context, _ uint64, _ int8, _ uint64) error {
			return errors.New("delete failed")
		},
	}
	svc := NewCollectionService(repo, nil, nil, logrus.New())
	require.Error(t, svc.Remove(context.Background(), 1, 1, 1))
}

func TestLikeService_Add_CreateError_Extra(t *testing.T) {
	likeRepo := &testutil.MockLikeRepository{
		GetFn: func(_ context.Context, _ uint64, _ int8, _ uint64) (*entity.Like, error) { return nil, nil },
		CreateFn: func(_ context.Context, _ *entity.Like) error {
			return errors.New("create failed")
		},
	}
	svc := NewLikeService(likeRepo, nil, nil, nil, logrus.New())
	require.Error(t, svc.Add(context.Background(), 1, 1, 1))
}

func TestLikeService_Add_CourseNotificationPath_Extra(t *testing.T) {
	likeRepo := &testutil.MockLikeRepository{
		GetFn:    func(_ context.Context, _ uint64, _ int8, _ uint64) (*entity.Like, error) { return nil, nil },
		CreateFn: func(_ context.Context, _ *entity.Like) error { return nil },
	}
	courseRepo := &testutil.MockCourseRepository{
		IncrLikeCountFn: func(_ context.Context, id uint64) error {
			assert.Equal(t, uint64(5), id)
			return nil
		},
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Course, error) {
			return &entity.Course{ID: id, AuthorID: 3}, nil
		},
	}
	notifRepo := &testutil.MockNotificationRepository{
		CreateFn: func(_ context.Context, n *entity.Notification) error {
			require.NotNil(t, n.UserID)
			assert.Equal(t, uint64(3), *n.UserID)
			return nil
		},
	}
	svc := NewLikeService(likeRepo, nil, courseRepo, notifRepo, logrus.New())
	require.NoError(t, svc.Add(context.Background(), 1, 2, 5))
}

func TestLikeService_Remove_DeleteError_Extra(t *testing.T) {
	likeRepo := &testutil.MockLikeRepository{
		DeleteFn: func(_ context.Context, _ uint64, _ int8, _ uint64) error {
			return errors.New("delete failed")
		},
	}
	svc := NewLikeService(likeRepo, nil, nil, nil, logrus.New())
	require.Error(t, svc.Remove(context.Background(), 1, 1, 1))
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

// ==================== CommentService ====================

func TestCommentService_List(t *testing.T) {
	repo := &testutil.MockCommentRepository{
		ListFn: func(_ context.Context, ct int8, cid uint64, p, ps int) ([]*entity.Comment, int64, error) {
			return []*entity.Comment{{ID: 1}}, 1, nil
		},
	}
	svc := NewCommentService(repo, nil, nil, nil, logrus.New())
	list, total, err := svc.List(context.Background(), 1, 10, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
}

func TestCommentService_List_Err(t *testing.T) {
	repo := &testutil.MockCommentRepository{
		ListFn: func(_ context.Context, ct int8, cid uint64, p, ps int) ([]*entity.Comment, int64, error) {
			return nil, 0, apperrors.NewInternal("db", nil)
		},
	}
	svc := NewCommentService(repo, nil, nil, nil, logrus.New())
	_, _, err := svc.List(context.Background(), 1, 10, 1, 20)
	require.Error(t, err)
}

func TestCommentService_Create_OK(t *testing.T) {
	repo := &testutil.MockCommentRepository{
		CreateFn: func(_ context.Context, c *entity.Comment) error { return nil },
	}
	svc := NewCommentService(repo, nil, nil, nil, logrus.New())
	c, err := svc.Create(context.Background(), 1, 1, 10, &dto.CreateCommentRequest{Content: "hello"})
	require.NoError(t, err)
	assert.Equal(t, "hello", c.Content)
}

func TestCommentService_Create_IncrCountAndNotify(t *testing.T) {
	repo := &testutil.MockCommentRepository{
		CreateFn: func(_ context.Context, c *entity.Comment) error { return nil },
	}
	articleRepo := &testutil.MockArticleRepository{
		IncrCommentCountFn: func(_ context.Context, id uint64) error {
			assert.Equal(t, uint64(10), id)
			return nil
		},
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Article, error) {
			return &entity.Article{ID: id, AuthorID: 2}, nil
		},
	}
	notifRepo := &testutil.MockNotificationRepository{
		CreateFn: func(_ context.Context, n *entity.Notification) error {
			require.NotNil(t, n.UserID)
			assert.Equal(t, uint64(2), *n.UserID)
			assert.Equal(t, int8(2), n.Type)
			return nil
		},
	}
	svc := NewCommentService(repo, articleRepo, nil, notifRepo, logrus.New())
	_, err := svc.Create(context.Background(), 1, 1, 10, &dto.CreateCommentRequest{Content: "hello"})
	require.NoError(t, err)
}

func TestCommentService_Create_ReplyNotifyParent(t *testing.T) {
	repo := &testutil.MockCommentRepository{
		CreateFn: func(_ context.Context, c *entity.Comment) error { return nil },
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Comment, error) {
			return &entity.Comment{ID: id, UserID: 3}, nil
		},
	}
	notifRepo := &testutil.MockNotificationRepository{
		CreateFn: func(_ context.Context, n *entity.Notification) error {
			require.NotNil(t, n.UserID)
			assert.Equal(t, uint64(3), *n.UserID)
			return nil
		},
	}
	svc := NewCommentService(repo, nil, nil, notifRepo, logrus.New())
	_, err := svc.Create(context.Background(), 1, 1, 10, &dto.CreateCommentRequest{Content: "reply", ParentID: 99})
	require.NoError(t, err)
}

func TestCommentService_Create_Err(t *testing.T) {
	repo := &testutil.MockCommentRepository{
		CreateFn: func(_ context.Context, c *entity.Comment) error { return apperrors.NewInternal("db", nil) },
	}
	svc := NewCommentService(repo, nil, nil, nil, logrus.New())
	_, err := svc.Create(context.Background(), 1, 1, 10, &dto.CreateCommentRequest{Content: "hello"})
	require.Error(t, err)
}

func TestCommentService_Create_MaskSensitiveWords(t *testing.T) {
	repo := &testutil.MockCommentRepository{
		CreateFn: func(_ context.Context, c *entity.Comment) error { return nil },
	}
	wordsRepo := &testutil.MockSensitiveWordRepository{
		ListEnabledWordsFn: func(_ context.Context) ([]string, error) {
			return []string{"bad", "词"}, nil
		},
	}
	svc := NewCommentService(repo, nil, nil, nil, logrus.New(), wordsRepo)
	c, err := svc.Create(context.Background(), 1, 1, 10, &dto.CreateCommentRequest{Content: "bad词"})
	require.NoError(t, err)
	assert.Equal(t, "****", c.Content)
}

func TestCommentService_AdminList_OK(t *testing.T) {
	repo := &testutil.MockCommentRepository{
		ListAdminFn: func(_ context.Context, p, ps int, st *int8) ([]*entity.Comment, int64, error) {
			return []*entity.Comment{{ID: 1}}, 1, nil
		},
	}
	svc := NewCommentService(repo, nil, nil, nil, logrus.New())
	list, total, err := svc.AdminList(context.Background(), 1, 20, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
}

func TestCommentService_Audit_OK(t *testing.T) {
	st := int8(1)
	repo := &testutil.MockCommentRepository{
		GetByIDFn:      func(_ context.Context, id uint64) (*entity.Comment, error) { return &entity.Comment{ID: id}, nil },
		UpdateStatusFn: func(_ context.Context, id uint64, status int8) error { return nil },
	}
	svc := NewCommentService(repo, nil, nil, nil, logrus.New())
	err := svc.Audit(context.Background(), 1, &dto.AuditCommentRequest{Status: st})
	require.NoError(t, err)
}

func TestCommentService_Audit_GetByIDErr(t *testing.T) {
	repo := &testutil.MockCommentRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Comment, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := NewCommentService(repo, nil, nil, nil, logrus.New())
	err := svc.Audit(context.Background(), 1, &dto.AuditCommentRequest{Status: 1})
	require.Error(t, err)
}

func TestCommentService_Audit_NotFound(t *testing.T) {
	repo := &testutil.MockCommentRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Comment, error) { return nil, nil },
	}
	svc := NewCommentService(repo, nil, nil, nil, logrus.New())
	err := svc.Audit(context.Background(), 1, &dto.AuditCommentRequest{Status: 1})
	require.Error(t, err)
}

func TestCommentService_Delete_OK(t *testing.T) {
	repo := &testutil.MockCommentRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Comment, error) { return &entity.Comment{ID: id}, nil },
		DeleteFn:  func(_ context.Context, id uint64) error { return nil },
	}
	svc := NewCommentService(repo, nil, nil, nil, logrus.New())
	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestCommentService_Delete_GetByIDErr(t *testing.T) {
	repo := &testutil.MockCommentRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Comment, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	svc := NewCommentService(repo, nil, nil, nil, logrus.New())
	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
}

func TestCommentService_Delete_NotFound(t *testing.T) {
	repo := &testutil.MockCommentRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Comment, error) { return nil, nil },
	}
	svc := NewCommentService(repo, nil, nil, nil, logrus.New())
	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
}

func TestCommentService_Delete_HasReplies(t *testing.T) {
	repo := &testutil.MockCommentRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Comment, error) {
			return &entity.Comment{ID: id, ContentType: 1, ContentID: 9}, nil
		},
		HasRepliesFn: func(_ context.Context, id uint64) (bool, error) { return true, nil },
	}
	svc := NewCommentService(repo, nil, nil, nil, logrus.New())
	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
}

func TestCommentService_Delete_DecrCourseCommentCount(t *testing.T) {
	repo := &testutil.MockCommentRepository{
		GetByIDFn: func(_ context.Context, id uint64) (*entity.Comment, error) {
			return &entity.Comment{ID: id, ContentType: 2, ContentID: 5}, nil
		},
		HasRepliesFn: func(_ context.Context, id uint64) (bool, error) { return false, nil },
		DeleteFn:     func(_ context.Context, id uint64) error { return nil },
	}
	courseRepo := &testutil.MockCourseRepository{
		DecrCommentCountFn: func(_ context.Context, id uint64) error {
			assert.Equal(t, uint64(5), id)
			return nil
		},
	}
	svc := NewCommentService(repo, nil, courseRepo, nil, logrus.New())
	err := svc.Delete(context.Background(), 1)
	require.NoError(t, err)
}
