package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

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
