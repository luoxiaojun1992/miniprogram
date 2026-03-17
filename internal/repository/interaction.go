package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

// ==================== ContentPermission Repository ====================

type contentPermissionRepository struct {
	db *gorm.DB
}

// NewContentPermissionRepository creates a new ContentPermissionRepository.
func NewContentPermissionRepository(db *gorm.DB) ContentPermissionRepository {
	return &contentPermissionRepository{db: db}
}

func (r *contentPermissionRepository) GetByContent(ctx context.Context, contentType int8, contentID uint64) ([]*entity.ContentPermission, error) {
	var perms []*entity.ContentPermission
	if err := r.db.WithContext(ctx).Where("content_type = ? AND content_id = ?", contentType, contentID).Find(&perms).Error; err != nil {
		return nil, errors.NewInternal("查询内容权限失败", err)
	}
	return perms, nil
}

func (r *contentPermissionRepository) SetContentPermissions(ctx context.Context, contentType int8, contentID uint64, roleIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM content_permissions WHERE content_type = ? AND content_id = ?", contentType, contentID).Error; err != nil {
			return errors.NewInternal("清除内容权限失败", err)
		}
		for _, roleID := range roleIDs {
			rid := roleID
			cp := &entity.ContentPermission{
				ContentType: contentType,
				ContentID:   contentID,
				RoleID:      &rid,
			}
			if err := tx.Create(cp).Error; err != nil {
				return errors.NewInternal("设置内容权限失败", err)
			}
		}
		return nil
	})
}

// ==================== StudyRecord Repository ====================

type studyRecordRepository struct {
	db *gorm.DB
}

// NewStudyRecordRepository creates a new StudyRecordRepository.
func NewStudyRecordRepository(db *gorm.DB) StudyRecordRepository {
	return &studyRecordRepository{db: db}
}

func (r *studyRecordRepository) GetByUserAndUnit(ctx context.Context, userID, unitID uint64) (*entity.UserStudyRecord, error) {
	var rec entity.UserStudyRecord
	res := r.db.WithContext(ctx).Where("user_id = ? AND unit_id = ?", userID, unitID).First(&rec)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询学习记录失败", res.Error)
	}
	return &rec, nil
}

func (r *studyRecordRepository) ListByUser(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.UserStudyRecord, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.UserStudyRecord{}).Where("user_id = ?", userID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternal("查询学习记录失败", err)
	}
	var records []*entity.UserStudyRecord
	if err := db.Order("updated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, errors.NewInternal("查询学习记录失败", err)
	}
	return records, total, nil
}

func (r *studyRecordRepository) Upsert(ctx context.Context, record *entity.UserStudyRecord) error {
	now := time.Now()
	record.LastStudyAt = &now
	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "unit_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"progress", "status", "last_study_at", "updated_at"}),
		}).Create(record)
	if res.Error != nil {
		return errors.NewInternal("更新学习记录失败", res.Error)
	}
	return nil
}

// ==================== Collection Repository ====================

type collectionRepository struct {
	db *gorm.DB
}

// NewCollectionRepository creates a new CollectionRepository.
func NewCollectionRepository(db *gorm.DB) CollectionRepository {
	return &collectionRepository{db: db}
}

func (r *collectionRepository) Get(ctx context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Collection, error) {
	var c entity.Collection
	res := r.db.WithContext(ctx).Where("user_id = ? AND content_type = ? AND content_id = ?", userID, contentType, contentID).First(&c)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询收藏失败", res.Error)
	}
	return &c, nil
}

func (r *collectionRepository) List(ctx context.Context, userID uint64, page, pageSize int, contentType *int8) ([]*entity.Collection, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.Collection{}).Where("user_id = ?", userID)
	if contentType != nil {
		db = db.Where("content_type = ?", *contentType)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternal("查询收藏列表失败", err)
	}
	var collections []*entity.Collection
	if err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&collections).Error; err != nil {
		return nil, 0, errors.NewInternal("查询收藏列表失败", err)
	}
	return collections, total, nil
}

func (r *collectionRepository) Create(ctx context.Context, collection *entity.Collection) error {
	if err := r.db.WithContext(ctx).Create(collection).Error; err != nil {
		return errors.NewInternal("创建收藏失败", err)
	}
	return nil
}

func (r *collectionRepository) Delete(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	if err := r.db.WithContext(ctx).Where("user_id = ? AND content_type = ? AND content_id = ?", userID, contentType, contentID).Delete(&entity.Collection{}).Error; err != nil {
		return errors.NewInternal("删除收藏失败", err)
	}
	return nil
}

// ==================== Like Repository ====================

type likeRepository struct {
	db *gorm.DB
}

// NewLikeRepository creates a new LikeRepository.
func NewLikeRepository(db *gorm.DB) LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) Get(ctx context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Like, error) {
	var l entity.Like
	res := r.db.WithContext(ctx).Where("user_id = ? AND content_type = ? AND content_id = ?", userID, contentType, contentID).First(&l)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询点赞失败", res.Error)
	}
	return &l, nil
}

func (r *likeRepository) Create(ctx context.Context, like *entity.Like) error {
	if err := r.db.WithContext(ctx).Create(like).Error; err != nil {
		return errors.NewInternal("创建点赞失败", err)
	}
	return nil
}

func (r *likeRepository) Delete(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	if err := r.db.WithContext(ctx).Where("user_id = ? AND content_type = ? AND content_id = ?", userID, contentType, contentID).Delete(&entity.Like{}).Error; err != nil {
		return errors.NewInternal("删除点赞失败", err)
	}
	return nil
}

// ==================== Comment Repository ====================

type commentRepository struct {
	db *gorm.DB
}

// NewCommentRepository creates a new CommentRepository.
func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) GetByID(ctx context.Context, id uint64) (*entity.Comment, error) {
	var c entity.Comment
	res := r.db.WithContext(ctx).Preload("User").First(&c, id)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询评论失败", res.Error)
	}
	return &c, nil
}

func (r *commentRepository) List(ctx context.Context, contentType int8, contentID uint64, page, pageSize int) ([]*entity.Comment, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.Comment{}).
		Where("content_type = ? AND content_id = ? AND status = 1 AND parent_id = 0", contentType, contentID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternal("查询评论列表失败", err)
	}
	var comments []*entity.Comment
	if err := db.Preload("User").Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&comments).Error; err != nil {
		return nil, 0, errors.NewInternal("查询评论列表失败", err)
	}
	return comments, total, nil
}

func (r *commentRepository) ListAdmin(ctx context.Context, page, pageSize int, status *int8) ([]*entity.Comment, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.Comment{})
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternal("查询评论列表失败", err)
	}
	var comments []*entity.Comment
	if err := db.Preload("User").Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&comments).Error; err != nil {
		return nil, 0, errors.NewInternal("查询评论列表失败", err)
	}
	return comments, total, nil
}

func (r *commentRepository) Create(ctx context.Context, comment *entity.Comment) error {
	if err := r.db.WithContext(ctx).Create(comment).Error; err != nil {
		return errors.NewInternal("创建评论失败", err)
	}
	return nil
}

func (r *commentRepository) UpdateStatus(ctx context.Context, id uint64, status int8) error {
	if err := r.db.WithContext(ctx).Model(&entity.Comment{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return errors.NewInternal("更新评论状态失败", err)
	}
	return nil
}

func (r *commentRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&entity.Comment{}, id).Error; err != nil {
		return errors.NewInternal("删除评论失败", err)
	}
	return nil
}

func (r *commentRepository) HasReplies(ctx context.Context, id uint64) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.Comment{}).Where("parent_id = ?", id).Count(&count).Error; err != nil {
		return false, errors.NewInternal("查询评论回复失败", err)
	}
	return count > 0, nil
}
