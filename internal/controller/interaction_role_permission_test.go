package controller

import (
	"context"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	apperrors "github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/testutil"
)

// ── StudyRecordController ─────────────────────────────────────────────────────

func TestStudyRecordCtrl_List_OK(t *testing.T) {
	svc := &testutil.MockStudyRecordService{
		ListFn: func(_ context.Context, userID uint64, p, ps int) ([]*entity.UserStudyRecord, int64, error) {
			return []*entity.UserStudyRecord{{ID: 1}}, 1, nil
		},
	}
	r := newTestRouterWithAuth(1, 1)
	r.GET("/study-records", NewStudyRecordController(svc, logrus.New()).List)
	assert.Equal(t, 200, doRequest(r, "GET", "/study-records", "").Code)
}

func TestStudyRecordCtrl_List_Unauthorized(t *testing.T) {
	r := newTestRouter()
	r.GET("/study-records", NewStudyRecordController(&testutil.MockStudyRecordService{}, logrus.New()).List)
	assert.Equal(t, 401, doRequest(r, "GET", "/study-records", "").Code)
}

func TestStudyRecordCtrl_List_SvcErr(t *testing.T) {
	svc := &testutil.MockStudyRecordService{
		ListFn: func(_ context.Context, userID uint64, p, ps int) ([]*entity.UserStudyRecord, int64, error) {
			return nil, 0, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouterWithAuth(1, 1)
	r.GET("/study-records", NewStudyRecordController(svc, logrus.New()).List)
	assert.Equal(t, 500, doRequest(r, "GET", "/study-records", "").Code)
}

func TestStudyRecordCtrl_Update_OK(t *testing.T) {
	svc := &testutil.MockStudyRecordService{
		UpdateFn: func(_ context.Context, userID uint64, req *dto.UpdateStudyRecordRequest) error { return nil },
	}
	r := newTestRouterWithAuth(1, 1)
	r.POST("/study-records", NewStudyRecordController(svc, logrus.New()).Update)
	assert.Equal(t, 200, doRequest(r, "POST", "/study-records", `{"unit_id":1,"progress":50}`).Code)
}

func TestStudyRecordCtrl_Update_Unauthorized(t *testing.T) {
	r := newTestRouter()
	r.POST("/study-records", NewStudyRecordController(&testutil.MockStudyRecordService{}, logrus.New()).Update)
	assert.Equal(t, 401, doRequest(r, "POST", "/study-records", `{"unit_id":1,"progress":50}`).Code)
}

func TestStudyRecordCtrl_Update_BadJSON(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.POST("/study-records", NewStudyRecordController(&testutil.MockStudyRecordService{}, logrus.New()).Update)
	assert.Equal(t, 400, doRequest(r, "POST", "/study-records", `bad`).Code)
}

func TestStudyRecordCtrl_Update_Validation(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.POST("/study-records", NewStudyRecordController(&testutil.MockStudyRecordService{}, logrus.New()).Update)
	assert.Equal(t, 400, doRequest(r, "POST", "/study-records", `{"unit_id":0,"progress":0}`).Code)
}

func TestStudyRecordCtrl_Update_SvcErr(t *testing.T) {
	svc := &testutil.MockStudyRecordService{
		UpdateFn: func(_ context.Context, userID uint64, req *dto.UpdateStudyRecordRequest) error {
			return apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouterWithAuth(1, 1)
	r.POST("/study-records", NewStudyRecordController(svc, logrus.New()).Update)
	assert.Equal(t, 500, doRequest(r, "POST", "/study-records", `{"unit_id":1,"progress":50}`).Code)
}

// ── CollectionController ──────────────────────────────────────────────────────

func TestCollectionCtrl_List_OK(t *testing.T) {
	svc := &testutil.MockCollectionService{
		ListFn: func(_ context.Context, uid uint64, p, ps int, ct *int8) ([]*entity.Collection, int64, error) {
			return []*entity.Collection{{ID: 1}}, 1, nil
		},
	}
	r := newTestRouterWithAuth(1, 1)
	r.GET("/collections", NewCollectionController(svc, logrus.New()).List)
	assert.Equal(t, 200, doRequest(r, "GET", "/collections?content_type=1", "").Code)
}

func TestCollectionCtrl_List_Unauthorized(t *testing.T) {
	r := newTestRouter()
	r.GET("/collections", NewCollectionController(&testutil.MockCollectionService{}, logrus.New()).List)
	assert.Equal(t, 401, doRequest(r, "GET", "/collections", "").Code)
}

func TestCollectionCtrl_List_SvcErr(t *testing.T) {
	svc := &testutil.MockCollectionService{
		ListFn: func(_ context.Context, uid uint64, p, ps int, ct *int8) ([]*entity.Collection, int64, error) {
			return nil, 0, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouterWithAuth(1, 1)
	r.GET("/collections", NewCollectionController(svc, logrus.New()).List)
	assert.Equal(t, 500, doRequest(r, "GET", "/collections", "").Code)
}

func TestCollectionCtrl_Add_OK(t *testing.T) {
	svc := &testutil.MockCollectionService{
		AddFn: func(_ context.Context, uid uint64, ct int8, cid uint64) error { return nil },
	}
	r := newTestRouterWithAuth(1, 1)
	r.POST("/collections/:content_type/:content_id", NewCollectionController(svc, logrus.New()).Add)
	assert.Equal(t, 201, doRequest(r, "POST", "/collections/1/10", "").Code)
}

func TestCollectionCtrl_Add_Unauthorized(t *testing.T) {
	r := newTestRouter()
	r.POST("/collections/:content_type/:content_id", NewCollectionController(&testutil.MockCollectionService{}, logrus.New()).Add)
	assert.Equal(t, 401, doRequest(r, "POST", "/collections/1/10", "").Code)
}

func TestCollectionCtrl_Add_BadContentType(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.POST("/collections/:content_type/:content_id", NewCollectionController(&testutil.MockCollectionService{}, logrus.New()).Add)
	assert.Equal(t, 400, doRequest(r, "POST", "/collections/abc/10", "").Code)
}

func TestCollectionCtrl_Add_InvalidContentTypeValue(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.POST("/collections/:content_type/:content_id", NewCollectionController(&testutil.MockCollectionService{}, logrus.New()).Add)
	assert.Equal(t, 400, doRequest(r, "POST", "/collections/3/10", "").Code)
}

func TestCollectionCtrl_Add_BadContentID(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.POST("/collections/:content_type/:content_id", NewCollectionController(&testutil.MockCollectionService{}, logrus.New()).Add)
	assert.Equal(t, 400, doRequest(r, "POST", "/collections/1/abc", "").Code)
}

func TestCollectionCtrl_Add_SvcErr(t *testing.T) {
	svc := &testutil.MockCollectionService{
		AddFn: func(_ context.Context, uid uint64, ct int8, cid uint64) error {
			return apperrors.NewConflict("already", nil)
		},
	}
	r := newTestRouterWithAuth(1, 1)
	r.POST("/collections/:content_type/:content_id", NewCollectionController(svc, logrus.New()).Add)
	assert.Equal(t, 409, doRequest(r, "POST", "/collections/1/10", "").Code)
}

func TestCollectionCtrl_Remove_OK(t *testing.T) {
	svc := &testutil.MockCollectionService{
		RemoveFn: func(_ context.Context, uid uint64, ct int8, cid uint64) error { return nil },
	}
	r := newTestRouterWithAuth(1, 1)
	r.DELETE("/collections/:content_type/:content_id", NewCollectionController(svc, logrus.New()).Remove)
	assert.Equal(t, http.StatusNoContent, doRequest(r, "DELETE", "/collections/1/10", "").Code)
}

func TestCollectionCtrl_Remove_Unauthorized(t *testing.T) {
	r := newTestRouter()
	r.DELETE("/collections/:content_type/:content_id", NewCollectionController(&testutil.MockCollectionService{}, logrus.New()).Remove)
	assert.Equal(t, 401, doRequest(r, "DELETE", "/collections/1/10", "").Code)
}

func TestCollectionCtrl_Remove_BadContentType(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.DELETE("/collections/:content_type/:content_id", NewCollectionController(&testutil.MockCollectionService{}, logrus.New()).Remove)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/collections/abc/10", "").Code)
}

func TestCollectionCtrl_Remove_InvalidContentTypeValue(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.DELETE("/collections/:content_type/:content_id", NewCollectionController(&testutil.MockCollectionService{}, logrus.New()).Remove)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/collections/3/10", "").Code)
}

func TestCollectionCtrl_Remove_BadContentID(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.DELETE("/collections/:content_type/:content_id", NewCollectionController(&testutil.MockCollectionService{}, logrus.New()).Remove)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/collections/1/abc", "").Code)
}

func TestCollectionCtrl_Remove_SvcErr(t *testing.T) {
	svc := &testutil.MockCollectionService{
		RemoveFn: func(_ context.Context, uid uint64, ct int8, cid uint64) error {
			return apperrors.NewNotFound("nf", nil)
		},
	}
	r := newTestRouterWithAuth(1, 1)
	r.DELETE("/collections/:content_type/:content_id", NewCollectionController(svc, logrus.New()).Remove)
	assert.Equal(t, 404, doRequest(r, "DELETE", "/collections/1/10", "").Code)
}

// ── LikeController ────────────────────────────────────────────────────────────

func TestLikeCtrl_Add_OK(t *testing.T) {
	svc := &testutil.MockLikeService{
		AddFn: func(_ context.Context, uid uint64, ct int8, cid uint64) error { return nil },
	}
	r := newTestRouterWithAuth(1, 1)
	r.POST("/likes/:content_type/:content_id", NewLikeController(svc, logrus.New()).Add)
	assert.Equal(t, 201, doRequest(r, "POST", "/likes/1/10", "").Code)
}

func TestLikeCtrl_Add_Unauthorized(t *testing.T) {
	r := newTestRouter()
	r.POST("/likes/:content_type/:content_id", NewLikeController(&testutil.MockLikeService{}, logrus.New()).Add)
	assert.Equal(t, 401, doRequest(r, "POST", "/likes/1/10", "").Code)
}

func TestLikeCtrl_Add_BadContentType(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.POST("/likes/:content_type/:content_id", NewLikeController(&testutil.MockLikeService{}, logrus.New()).Add)
	assert.Equal(t, 400, doRequest(r, "POST", "/likes/abc/10", "").Code)
}

func TestLikeCtrl_Add_InvalidContentTypeValue(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.POST("/likes/:content_type/:content_id", NewLikeController(&testutil.MockLikeService{}, logrus.New()).Add)
	assert.Equal(t, 400, doRequest(r, "POST", "/likes/3/10", "").Code)
}

func TestLikeCtrl_Add_BadContentID(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.POST("/likes/:content_type/:content_id", NewLikeController(&testutil.MockLikeService{}, logrus.New()).Add)
	assert.Equal(t, 400, doRequest(r, "POST", "/likes/1/abc", "").Code)
}

func TestLikeCtrl_Add_SvcErr(t *testing.T) {
	svc := &testutil.MockLikeService{
		AddFn: func(_ context.Context, uid uint64, ct int8, cid uint64) error {
			return apperrors.NewConflict("already", nil)
		},
	}
	r := newTestRouterWithAuth(1, 1)
	r.POST("/likes/:content_type/:content_id", NewLikeController(svc, logrus.New()).Add)
	assert.Equal(t, 409, doRequest(r, "POST", "/likes/1/10", "").Code)
}

func TestLikeCtrl_Remove_OK(t *testing.T) {
	svc := &testutil.MockLikeService{
		RemoveFn: func(_ context.Context, uid uint64, ct int8, cid uint64) error { return nil },
	}
	r := newTestRouterWithAuth(1, 1)
	r.DELETE("/likes/:content_type/:content_id", NewLikeController(svc, logrus.New()).Remove)
	assert.Equal(t, 200, doRequest(r, "DELETE", "/likes/1/10", "").Code)
}

func TestLikeCtrl_Remove_Unauthorized(t *testing.T) {
	r := newTestRouter()
	r.DELETE("/likes/:content_type/:content_id", NewLikeController(&testutil.MockLikeService{}, logrus.New()).Remove)
	assert.Equal(t, 401, doRequest(r, "DELETE", "/likes/1/10", "").Code)
}

func TestLikeCtrl_Remove_BadContentType(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.DELETE("/likes/:content_type/:content_id", NewLikeController(&testutil.MockLikeService{}, logrus.New()).Remove)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/likes/abc/10", "").Code)
}

func TestLikeCtrl_Remove_InvalidContentTypeValue(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.DELETE("/likes/:content_type/:content_id", NewLikeController(&testutil.MockLikeService{}, logrus.New()).Remove)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/likes/3/10", "").Code)
}

func TestLikeCtrl_Remove_BadContentID(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.DELETE("/likes/:content_type/:content_id", NewLikeController(&testutil.MockLikeService{}, logrus.New()).Remove)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/likes/1/abc", "").Code)
}

func TestLikeCtrl_Remove_SvcErr(t *testing.T) {
	svc := &testutil.MockLikeService{
		RemoveFn: func(_ context.Context, uid uint64, ct int8, cid uint64) error {
			return apperrors.NewNotFound("nf", nil)
		},
	}
	r := newTestRouterWithAuth(1, 1)
	r.DELETE("/likes/:content_type/:content_id", NewLikeController(svc, logrus.New()).Remove)
	assert.Equal(t, 404, doRequest(r, "DELETE", "/likes/1/10", "").Code)
}

// ── FollowController ──────────────────────────────────────────────────────────

func TestFollowCtrl_Add_OK(t *testing.T) {
	svc := &testutil.MockFollowService{
		AddFn: func(_ context.Context, followerID, followedID uint64) error { return nil },
	}
	r := newTestRouterWithAuth(1, 1)
	r.POST("/follows/:user_id", NewFollowController(svc, logrus.New()).Add)
	assert.Equal(t, 201, doRequest(r, "POST", "/follows/10", "").Code)
}

func TestFollowCtrl_Add_Unauthorized(t *testing.T) {
	r := newTestRouter()
	r.POST("/follows/:user_id", NewFollowController(&testutil.MockFollowService{}, logrus.New()).Add)
	assert.Equal(t, 401, doRequest(r, "POST", "/follows/10", "").Code)
}

func TestFollowCtrl_Add_BadUserID(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.POST("/follows/:user_id", NewFollowController(&testutil.MockFollowService{}, logrus.New()).Add)
	assert.Equal(t, 400, doRequest(r, "POST", "/follows/abc", "").Code)
}

func TestFollowCtrl_Remove_OK(t *testing.T) {
	svc := &testutil.MockFollowService{
		RemoveFn: func(_ context.Context, followerID, followedID uint64) error { return nil },
	}
	r := newTestRouterWithAuth(1, 1)
	r.DELETE("/follows/:user_id", NewFollowController(svc, logrus.New()).Remove)
	assert.Equal(t, 200, doRequest(r, "DELETE", "/follows/10", "").Code)
}

func TestFollowCtrl_Remove_Unauthorized(t *testing.T) {
	r := newTestRouter()
	r.DELETE("/follows/:user_id", NewFollowController(&testutil.MockFollowService{}, logrus.New()).Remove)
	assert.Equal(t, 401, doRequest(r, "DELETE", "/follows/10", "").Code)
}

func TestFollowCtrl_Remove_BadUserID(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.DELETE("/follows/:user_id", NewFollowController(&testutil.MockFollowService{}, logrus.New()).Remove)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/follows/abc", "").Code)
}

// ── CommentController ─────────────────────────────────────────────────────────

func TestCommentCtrl_List_OK(t *testing.T) {
	svc := &testutil.MockCommentService{
		ListFn: func(_ context.Context, ct int8, cid uint64, p, ps int) ([]*entity.Comment, int64, error) {
			return []*entity.Comment{{ID: 1}}, 1, nil
		},
	}
	r := newTestRouter()
	r.GET("/comments/:content_type/:content_id", NewCommentController(svc, logrus.New()).List)
	assert.Equal(t, 200, doRequest(r, "GET", "/comments/1/10", "").Code)
}

func TestCommentCtrl_List_BadContentType(t *testing.T) {
	r := newTestRouter()
	r.GET("/comments/:content_type/:content_id", NewCommentController(&testutil.MockCommentService{}, logrus.New()).List)
	assert.Equal(t, 400, doRequest(r, "GET", "/comments/abc/10", "").Code)
}

func TestCommentCtrl_List_InvalidContentTypeValue(t *testing.T) {
	r := newTestRouter()
	r.GET("/comments/:content_type/:content_id", NewCommentController(&testutil.MockCommentService{}, logrus.New()).List)
	assert.Equal(t, 400, doRequest(r, "GET", "/comments/3/10", "").Code)
}

func TestCommentCtrl_List_BadContentID(t *testing.T) {
	r := newTestRouter()
	r.GET("/comments/:content_type/:content_id", NewCommentController(&testutil.MockCommentService{}, logrus.New()).List)
	assert.Equal(t, 400, doRequest(r, "GET", "/comments/1/abc", "").Code)
}

func TestCommentCtrl_List_SvcErr(t *testing.T) {
	svc := &testutil.MockCommentService{
		ListFn: func(_ context.Context, ct int8, cid uint64, p, ps int) ([]*entity.Comment, int64, error) {
			return nil, 0, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouter()
	r.GET("/comments/:content_type/:content_id", NewCommentController(svc, logrus.New()).List)
	assert.Equal(t, 500, doRequest(r, "GET", "/comments/1/10", "").Code)
}

func TestCommentCtrl_Create_OK(t *testing.T) {
	svc := &testutil.MockCommentService{
		CreateFn: func(_ context.Context, uid uint64, ct int8, cid uint64, req *dto.CreateCommentRequest) (*entity.Comment, error) {
			return &entity.Comment{ID: 1}, nil
		},
	}
	r := newTestRouterWithAuth(1, 1)
	r.POST("/comments/:content_type/:content_id", NewCommentController(svc, logrus.New()).Create)
	assert.Equal(t, 201, doRequest(r, "POST", "/comments/1/10", `{"content":"Hello"}`).Code)
}

func TestCommentCtrl_Create_Unauthorized(t *testing.T) {
	r := newTestRouter()
	r.POST("/comments/:content_type/:content_id", NewCommentController(&testutil.MockCommentService{}, logrus.New()).Create)
	assert.Equal(t, 401, doRequest(r, "POST", "/comments/1/10", `{"content":"Hello"}`).Code)
}

func TestCommentCtrl_Create_BadContentType(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.POST("/comments/:content_type/:content_id", NewCommentController(&testutil.MockCommentService{}, logrus.New()).Create)
	assert.Equal(t, 400, doRequest(r, "POST", "/comments/abc/10", `{"content":"Hello"}`).Code)
}

func TestCommentCtrl_Create_InvalidContentTypeValue(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.POST("/comments/:content_type/:content_id", NewCommentController(&testutil.MockCommentService{}, logrus.New()).Create)
	assert.Equal(t, 400, doRequest(r, "POST", "/comments/3/10", `{"content":"Hello"}`).Code)
}

func TestCommentCtrl_Create_BadContentID(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.POST("/comments/:content_type/:content_id", NewCommentController(&testutil.MockCommentService{}, logrus.New()).Create)
	assert.Equal(t, 400, doRequest(r, "POST", "/comments/1/abc", `{"content":"Hello"}`).Code)
}

func TestCommentCtrl_Create_BadJSON(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.POST("/comments/:content_type/:content_id", NewCommentController(&testutil.MockCommentService{}, logrus.New()).Create)
	assert.Equal(t, 400, doRequest(r, "POST", "/comments/1/10", `bad`).Code)
}

func TestCommentCtrl_Create_Validation(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.POST("/comments/:content_type/:content_id", NewCommentController(&testutil.MockCommentService{}, logrus.New()).Create)
	assert.Equal(t, 400, doRequest(r, "POST", "/comments/1/10", `{"content":""}`).Code)
}

func TestCommentCtrl_Create_SvcErr(t *testing.T) {
	svc := &testutil.MockCommentService{
		CreateFn: func(_ context.Context, uid uint64, ct int8, cid uint64, req *dto.CreateCommentRequest) (*entity.Comment, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouterWithAuth(1, 1)
	r.POST("/comments/:content_type/:content_id", NewCommentController(svc, logrus.New()).Create)
	assert.Equal(t, 500, doRequest(r, "POST", "/comments/1/10", `{"content":"Hello"}`).Code)
}

func TestCommentCtrl_AdminList_OK(t *testing.T) {
	svc := &testutil.MockCommentService{
		AdminListFn: func(_ context.Context, p, ps int, st *int8) ([]*entity.Comment, int64, error) {
			return []*entity.Comment{{ID: 1}}, 1, nil
		},
	}
	r := newTestRouter()
	r.GET("/admin/comments", NewCommentController(svc, logrus.New()).AdminList)
	assert.Equal(t, 200, doRequest(r, "GET", "/admin/comments?status=1", "").Code)
}

func TestCommentCtrl_AdminList_SvcErr(t *testing.T) {
	svc := &testutil.MockCommentService{
		AdminListFn: func(_ context.Context, p, ps int, st *int8) ([]*entity.Comment, int64, error) {
			return nil, 0, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouter()
	r.GET("/admin/comments", NewCommentController(svc, logrus.New()).AdminList)
	assert.Equal(t, 500, doRequest(r, "GET", "/admin/comments", "").Code)
}

func TestCommentCtrl_AdminAudit_OK(t *testing.T) {
	svc := &testutil.MockCommentService{
		AuditFn: func(_ context.Context, id uint64, req *dto.AuditCommentRequest) error { return nil },
	}
	r := newTestRouter()
	r.PUT("/admin/comments/:id/audit", NewCommentController(svc, logrus.New()).AdminAudit)
	assert.Equal(t, 200, doRequest(r, "PUT", "/admin/comments/1/audit", `{"status":1}`).Code)
}

func TestCommentCtrl_AdminAudit_BadID(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/comments/:id/audit", NewCommentController(&testutil.MockCommentService{}, logrus.New()).AdminAudit)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/comments/abc/audit", `{"status":1}`).Code)
}

func TestCommentCtrl_AdminAudit_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/comments/:id/audit", NewCommentController(&testutil.MockCommentService{}, logrus.New()).AdminAudit)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/comments/1/audit", `bad`).Code)
}

func TestCommentCtrl_AdminAudit_Validation(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/comments/:id/audit", NewCommentController(&testutil.MockCommentService{}, logrus.New()).AdminAudit)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/comments/1/audit", `{"status":99}`).Code)
}

func TestCommentCtrl_AdminAudit_SvcErr(t *testing.T) {
	svc := &testutil.MockCommentService{
		AuditFn: func(_ context.Context, id uint64, req *dto.AuditCommentRequest) error {
			return apperrors.NewNotFound("nf", nil)
		},
	}
	r := newTestRouter()
	r.PUT("/admin/comments/:id/audit", NewCommentController(svc, logrus.New()).AdminAudit)
	assert.Equal(t, 404, doRequest(r, "PUT", "/admin/comments/1/audit", `{"status":1}`).Code)
}

func TestCommentCtrl_AdminDelete_OK(t *testing.T) {
	svc := &testutil.MockCommentService{
		DeleteFn: func(_ context.Context, id uint64) error { return nil },
	}
	r := newTestRouter()
	r.DELETE("/admin/comments/:id", NewCommentController(svc, logrus.New()).AdminDelete)
	assert.Equal(t, http.StatusNoContent, doRequest(r, "DELETE", "/admin/comments/1", "").Code)
}

func TestCommentCtrl_AdminDelete_BadID(t *testing.T) {
	r := newTestRouter()
	r.DELETE("/admin/comments/:id", NewCommentController(&testutil.MockCommentService{}, logrus.New()).AdminDelete)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/admin/comments/abc", "").Code)
}

func TestCommentCtrl_AdminDelete_SvcErr(t *testing.T) {
	svc := &testutil.MockCommentService{
		DeleteFn: func(_ context.Context, id uint64) error { return apperrors.NewNotFound("nf", nil) },
	}
	r := newTestRouter()
	r.DELETE("/admin/comments/:id", NewCommentController(svc, logrus.New()).AdminDelete)
	assert.Equal(t, 404, doRequest(r, "DELETE", "/admin/comments/1", "").Code)
}

// ── RoleController ────────────────────────────────────────────────────────────

func TestRoleCtrl_List_OK(t *testing.T) {
	svc := &testutil.MockRoleService{
		ListFn: func(_ context.Context) ([]*entity.Role, error) { return []*entity.Role{{ID: 1}}, nil },
	}
	r := newTestRouter()
	r.GET("/admin/roles", NewRoleController(svc, logrus.New()).List)
	assert.Equal(t, 200, doRequest(r, "GET", "/admin/roles", "").Code)
}

func TestRoleCtrl_List_SvcErr(t *testing.T) {
	svc := &testutil.MockRoleService{
		ListFn: func(_ context.Context) ([]*entity.Role, error) { return nil, apperrors.NewInternal("db", nil) },
	}
	r := newTestRouter()
	r.GET("/admin/roles", NewRoleController(svc, logrus.New()).List)
	assert.Equal(t, 500, doRequest(r, "GET", "/admin/roles", "").Code)
}

func TestRoleCtrl_GetByID_OK(t *testing.T) {
	svc := &testutil.MockRoleService{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Role, error) { return &entity.Role{ID: id}, nil },
	}
	r := newTestRouter()
	r.GET("/admin/roles/:id", NewRoleController(svc, logrus.New()).GetByID)
	assert.Equal(t, 200, doRequest(r, "GET", "/admin/roles/1", "").Code)
}

func TestRoleCtrl_GetByID_BadID(t *testing.T) {
	r := newTestRouter()
	r.GET("/admin/roles/:id", NewRoleController(&testutil.MockRoleService{}, logrus.New()).GetByID)
	assert.Equal(t, 400, doRequest(r, "GET", "/admin/roles/abc", "").Code)
}

func TestRoleCtrl_GetByID_SvcErr(t *testing.T) {
	svc := &testutil.MockRoleService{
		GetByIDFn: func(_ context.Context, id uint) (*entity.Role, error) {
			return nil, apperrors.NewNotFound("nf", nil)
		},
	}
	r := newTestRouter()
	r.GET("/admin/roles/:id", NewRoleController(svc, logrus.New()).GetByID)
	assert.Equal(t, 404, doRequest(r, "GET", "/admin/roles/1", "").Code)
}

func TestRoleCtrl_Create_OK(t *testing.T) {
	svc := &testutil.MockRoleService{
		CreateFn: func(_ context.Context, req *dto.CreateRoleRequest) (uint, error) { return 1, nil },
	}
	r := newTestRouter()
	r.POST("/admin/roles", NewRoleController(svc, logrus.New()).Create)
	assert.Equal(t, 201, doRequest(r, "POST", "/admin/roles", `{"name":"Admin","code":"admin"}`).Code)
}

func TestRoleCtrl_Create_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/roles", NewRoleController(&testutil.MockRoleService{}, logrus.New()).Create)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/roles", `bad`).Code)
}

func TestRoleCtrl_Create_Validation(t *testing.T) {
	r := newTestRouter()
	r.POST("/admin/roles", NewRoleController(&testutil.MockRoleService{}, logrus.New()).Create)
	assert.Equal(t, 400, doRequest(r, "POST", "/admin/roles", `{"name":"","code":""}`).Code)
}

func TestRoleCtrl_Create_SvcErr(t *testing.T) {
	svc := &testutil.MockRoleService{
		CreateFn: func(_ context.Context, req *dto.CreateRoleRequest) (uint, error) {
			return 0, apperrors.NewConflict("dup", nil)
		},
	}
	r := newTestRouter()
	r.POST("/admin/roles", NewRoleController(svc, logrus.New()).Create)
	assert.Equal(t, 409, doRequest(r, "POST", "/admin/roles", `{"name":"Admin","code":"admin"}`).Code)
}

func TestRoleCtrl_Update_OK(t *testing.T) {
	svc := &testutil.MockRoleService{
		UpdateFn: func(_ context.Context, id uint, req *dto.UpdateRoleRequest) error { return nil },
	}
	r := newTestRouter()
	r.PUT("/admin/roles/:id", NewRoleController(svc, logrus.New()).Update)
	assert.Equal(t, 200, doRequest(r, "PUT", "/admin/roles/1", `{"name":"Admin","code":"admin","permission_ids":[1]}`).Code)
}

func TestRoleCtrl_Update_BadID(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/roles/:id", NewRoleController(&testutil.MockRoleService{}, logrus.New()).Update)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/roles/abc", `{}`).Code)
}

func TestRoleCtrl_Update_BadJSON(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/roles/:id", NewRoleController(&testutil.MockRoleService{}, logrus.New()).Update)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/roles/1", `bad`).Code)
}

func TestRoleCtrl_Update_Validation(t *testing.T) {
	r := newTestRouter()
	r.PUT("/admin/roles/:id", NewRoleController(&testutil.MockRoleService{}, logrus.New()).Update)
	assert.Equal(t, 400, doRequest(r, "PUT", "/admin/roles/1", `{"name":"","code":""}`).Code)
}

func TestRoleCtrl_Update_SvcErr(t *testing.T) {
	svc := &testutil.MockRoleService{
		UpdateFn: func(_ context.Context, id uint, req *dto.UpdateRoleRequest) error {
			return apperrors.NewNotFound("nf", nil)
		},
	}
	r := newTestRouter()
	r.PUT("/admin/roles/:id", NewRoleController(svc, logrus.New()).Update)
	assert.Equal(t, 404, doRequest(r, "PUT", "/admin/roles/1", `{"name":"A","code":"a","permission_ids":[1]}`).Code)
}

func TestRoleCtrl_Delete_OK(t *testing.T) {
	svc := &testutil.MockRoleService{
		DeleteFn: func(_ context.Context, id uint) error { return nil },
	}
	r := newTestRouter()
	r.DELETE("/admin/roles/:id", NewRoleController(svc, logrus.New()).Delete)
	assert.Equal(t, http.StatusNoContent, doRequest(r, "DELETE", "/admin/roles/1", "").Code)
}

func TestRoleCtrl_Delete_BadID(t *testing.T) {
	r := newTestRouter()
	r.DELETE("/admin/roles/:id", NewRoleController(&testutil.MockRoleService{}, logrus.New()).Delete)
	assert.Equal(t, 400, doRequest(r, "DELETE", "/admin/roles/abc", "").Code)
}

func TestRoleCtrl_Delete_SvcErr(t *testing.T) {
	svc := &testutil.MockRoleService{
		DeleteFn: func(_ context.Context, id uint) error { return apperrors.NewNotFound("nf", nil) },
	}
	r := newTestRouter()
	r.DELETE("/admin/roles/:id", NewRoleController(svc, logrus.New()).Delete)
	assert.Equal(t, 404, doRequest(r, "DELETE", "/admin/roles/1", "").Code)
}

// ── PermissionController ──────────────────────────────────────────────────────

func TestPermissionCtrl_GetTree_OK(t *testing.T) {
	svc := &testutil.MockPermissionService{
		GetTreeFn: func(_ context.Context) ([]*entity.Permission, error) {
			return []*entity.Permission{{ID: 1}}, nil
		},
	}
	r := newTestRouter()
	r.GET("/admin/permissions", NewPermissionController(svc, logrus.New()).GetTree)
	assert.Equal(t, 200, doRequest(r, "GET", "/admin/permissions", "").Code)
}

func TestPermissionCtrl_GetTree_Err(t *testing.T) {
	svc := &testutil.MockPermissionService{
		GetTreeFn: func(_ context.Context) ([]*entity.Permission, error) {
			return nil, apperrors.NewInternal("db", nil)
		},
	}
	r := newTestRouter()
	r.GET("/admin/permissions", NewPermissionController(svc, logrus.New()).GetTree)
	assert.Equal(t, 500, doRequest(r, "GET", "/admin/permissions", "").Code)
}

// ── Additional coverage tests ─────────────────────────────────────────────────

func TestStudyRecordCtrl_List_BindQueryErr(t *testing.T) {
r := newTestRouterWithAuth(1, 1)
r.GET("/study-records", NewStudyRecordController(&testutil.MockStudyRecordService{}, logrus.New()).List)
assert.Equal(t, 400, doRequest(r, "GET", "/study-records?page=abc", "").Code)
}

func TestCollectionCtrl_List_BindQueryErr(t *testing.T) {
r := newTestRouterWithAuth(1, 1)
r.GET("/collections", NewCollectionController(&testutil.MockCollectionService{}, logrus.New()).List)
assert.Equal(t, 400, doRequest(r, "GET", "/collections?page=abc", "").Code)
}

func TestCollectionCtrl_List_BadContentTypeQuery(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.GET("/collections", NewCollectionController(&testutil.MockCollectionService{}, logrus.New()).List)
	assert.Equal(t, 400, doRequest(r, "GET", "/collections?content_type=abc", "").Code)
}

func TestCollectionCtrl_List_InvalidContentTypeQueryValue(t *testing.T) {
	r := newTestRouterWithAuth(1, 1)
	r.GET("/collections", NewCollectionController(&testutil.MockCollectionService{}, logrus.New()).List)
	assert.Equal(t, 400, doRequest(r, "GET", "/collections?content_type=3", "").Code)
}

func TestCollectionCtrl_List_NoContentType(t *testing.T) {
svc := &testutil.MockCollectionService{
ListFn: func(_ context.Context, uid uint64, p, ps int, ct *int8) ([]*entity.Collection, int64, error) {
return []*entity.Collection{{ID: 1}}, 1, nil
},
}
r := newTestRouterWithAuth(1, 1)
r.GET("/collections", NewCollectionController(svc, logrus.New()).List)
assert.Equal(t, 200, doRequest(r, "GET", "/collections", "").Code)
}

func TestCommentCtrl_List_BindQueryErr(t *testing.T) {
r := newTestRouter()
r.GET("/comments/:content_type/:content_id", NewCommentController(&testutil.MockCommentService{}, logrus.New()).List)
assert.Equal(t, 400, doRequest(r, "GET", "/comments/1/10?page=abc", "").Code)
}

func TestCommentCtrl_AdminList_BindQueryErr(t *testing.T) {
r := newTestRouter()
r.GET("/admin/comments", NewCommentController(&testutil.MockCommentService{}, logrus.New()).AdminList)
assert.Equal(t, 400, doRequest(r, "GET", "/admin/comments?page=abc", "").Code)
}
