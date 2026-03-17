package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type fileRepository struct {
	db *gorm.DB
}

// NewFileRepository creates a new FileRepository.
func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) GetByID(ctx context.Context, id uint64) (*entity.File, error) {
	var file entity.File
	res := r.db.WithContext(ctx).First(&file, id)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, errors.NewInternal("查询文件失败", res.Error)
	}
	return &file, nil
}

func (r *fileRepository) Create(ctx context.Context, file *entity.File) error {
	if err := r.db.WithContext(ctx).Create(file).Error; err != nil {
		return errors.NewInternal("创建文件记录失败", err)
	}
	return nil
}

type articleAttachmentRepository struct {
	db *gorm.DB
}

// NewArticleAttachmentRepository creates a new ArticleAttachmentRepository.
func NewArticleAttachmentRepository(db *gorm.DB) ArticleAttachmentRepository {
	return &articleAttachmentRepository{db: db}
}

func (r *articleAttachmentRepository) ListFileIDs(ctx context.Context, articleID uint64) ([]uint64, error) {
	var rows []*entity.ArticleAttachment
	if err := r.db.WithContext(ctx).Where("article_id = ?", articleID).Order("sort_order ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, errors.NewInternal("查询文章附件失败", err)
	}
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.FileID)
	}
	return ids, nil
}

func (r *articleAttachmentRepository) Replace(ctx context.Context, articleID uint64, fileIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("article_id = ?", articleID).Delete(&entity.ArticleAttachment{}).Error; err != nil {
			return errors.NewInternal("清理文章附件失败", err)
		}
		rows := make([]*entity.ArticleAttachment, 0, len(fileIDs))
		seen := make(map[uint64]struct{}, len(fileIDs))
		for i, fileID := range fileIDs {
			if fileID == 0 {
				continue
			}
			if _, ok := seen[fileID]; ok {
				continue
			}
			seen[fileID] = struct{}{}
			rows = append(rows, &entity.ArticleAttachment{
				ArticleID: articleID,
				FileID:    fileID,
				SortOrder: i,
			})
		}
		if len(rows) == 0 {
			return nil
		}
		if err := tx.Create(&rows).Error; err != nil {
			return errors.NewInternal("写入文章附件失败", err)
		}
		return nil
	})
}

type courseAttachmentRepository struct {
	db *gorm.DB
}

// NewCourseAttachmentRepository creates a new CourseAttachmentRepository.
func NewCourseAttachmentRepository(db *gorm.DB) CourseAttachmentRepository {
	return &courseAttachmentRepository{db: db}
}

func (r *courseAttachmentRepository) ListFileIDs(ctx context.Context, courseID uint64) ([]uint64, error) {
	var rows []*entity.CourseAttachment
	if err := r.db.WithContext(ctx).Where("course_id = ?", courseID).Order("sort_order ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, errors.NewInternal("查询课程附件失败", err)
	}
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.FileID)
	}
	return ids, nil
}

func (r *courseAttachmentRepository) Replace(ctx context.Context, courseID uint64, fileIDs []uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("course_id = ?", courseID).Delete(&entity.CourseAttachment{}).Error; err != nil {
			return errors.NewInternal("清理课程附件失败", err)
		}
		rows := make([]*entity.CourseAttachment, 0, len(fileIDs))
		seen := make(map[uint64]struct{}, len(fileIDs))
		for i, fileID := range fileIDs {
			if fileID == 0 {
				continue
			}
			if _, ok := seen[fileID]; ok {
				continue
			}
			seen[fileID] = struct{}{}
			rows = append(rows, &entity.CourseAttachment{
				CourseID:  courseID,
				FileID:    fileID,
				SortOrder: i,
			})
		}
		if len(rows) == 0 {
			return nil
		}
		if err := tx.Create(&rows).Error; err != nil {
			return errors.NewInternal("写入课程附件失败", err)
		}
		return nil
	})
}
