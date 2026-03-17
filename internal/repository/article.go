package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type articleRepository struct {
	db *gorm.DB
}

// NewArticleRepository creates a new ArticleRepository.
func NewArticleRepository(db *gorm.DB) ArticleRepository {
	return &articleRepository{db: db}
}

func (r *articleRepository) GetByID(ctx context.Context, id uint64) (*entity.Article, error) {
	var a entity.Article
	res := r.db.WithContext(ctx).Preload("Author").First(&a, id)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询文章失败", res.Error)
	}
	return &a, nil
}

func (r *articleRepository) List(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8, sort string) ([]*entity.Article, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.Article{}).Preload("Author")
	if keyword != "" {
		db = db.Where("title LIKE ? OR summary LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if moduleID != nil {
		db = db.Where("module_id = ?", *moduleID)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternal("查询文章列表失败", err)
	}
	orderClause := "sort_order DESC, created_at DESC"
	switch sort {
	case "created_at":
		orderClause = "created_at ASC"
	case "-view_count":
		orderClause = "view_count DESC"
	case "-like_count":
		orderClause = "like_count DESC"
	}
	var articles []*entity.Article
	if err := db.Order(orderClause).Offset((page - 1) * pageSize).Limit(pageSize).Find(&articles).Error; err != nil {
		return nil, 0, errors.NewInternal("查询文章列表失败", err)
	}
	return articles, total, nil
}

func (r *articleRepository) Create(ctx context.Context, article *entity.Article) error {
	if err := r.db.WithContext(ctx).Create(article).Error; err != nil {
		return errors.NewInternal("创建文章失败", err)
	}
	return nil
}

func (r *articleRepository) Update(ctx context.Context, article *entity.Article) error {
	if err := r.db.WithContext(ctx).Save(article).Error; err != nil {
		return errors.NewInternal("更新文章失败", err)
	}
	return nil
}

func (r *articleRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&entity.Article{}, id).Error; err != nil {
		return errors.NewInternal("删除文章失败", err)
	}
	return nil
}

func (r *articleRepository) IncrViewCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec("UPDATE articles SET view_count = view_count + 1 WHERE id = ?", id).Error; err != nil {
		return errors.NewInternal("更新浏览量失败", err)
	}
	return nil
}

func (r *articleRepository) IncrLikeCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec("UPDATE articles SET like_count = like_count + 1 WHERE id = ?", id).Error; err != nil {
		return errors.NewInternal("更新点赞数失败", err)
	}
	return nil
}

func (r *articleRepository) DecrLikeCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec("UPDATE articles SET like_count = CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END WHERE id = ?", id).Error; err != nil {
		return errors.NewInternal("更新点赞数失败", err)
	}
	return nil
}

func (r *articleRepository) IncrCollectCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec("UPDATE articles SET collect_count = collect_count + 1 WHERE id = ?", id).Error; err != nil {
		return errors.NewInternal("更新收藏数失败", err)
	}
	return nil
}

func (r *articleRepository) DecrCollectCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec("UPDATE articles SET collect_count = CASE WHEN collect_count > 0 THEN collect_count - 1 ELSE 0 END WHERE id = ?", id).Error; err != nil {
		return errors.NewInternal("更新收藏数失败", err)
	}
	return nil
}

func (r *articleRepository) IncrCommentCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec("UPDATE articles SET comment_count = comment_count + 1 WHERE id = ?", id).Error; err != nil {
		return errors.NewInternal("更新评论数失败", err)
	}
	return nil
}

func (r *articleRepository) DecrCommentCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec("UPDATE articles SET comment_count = CASE WHEN comment_count > 0 THEN comment_count - 1 ELSE 0 END WHERE id = ?", id).Error; err != nil {
		return errors.NewInternal("更新评论数失败", err)
	}
	return nil
}

func (r *articleRepository) IncrShareCount(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Exec("UPDATE articles SET share_count = share_count + 1 WHERE id = ?", id).Error; err != nil {
		return errors.NewInternal("更新分享数失败", err)
	}
	return nil
}

func (r *articleRepository) HasAssociations(ctx context.Context, id uint64) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT (
			(SELECT COUNT(1) FROM likes WHERE content_type = 1 AND content_id = ?) +
			(SELECT COUNT(1) FROM collections WHERE content_type = 1 AND content_id = ?) +
			(SELECT COUNT(1) FROM comments WHERE content_type = 1 AND content_id = ?)
		) AS cnt
	`, id, id, id).Scan(&count).Error; err != nil {
		return false, errors.NewInternal("查询文章关联失败", err)
	}
	return count > 0, nil
}
