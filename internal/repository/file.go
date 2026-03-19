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

func (r *fileRepository) DeleteByIDs(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	uniq := make(map[uint64]struct{}, len(ids))
	normalized := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := uniq[id]; ok {
			continue
		}
		uniq[id] = struct{}{}
		normalized = append(normalized, id)
	}
	if len(normalized) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Where("id IN ?", normalized).Delete(&entity.File{}).Error; err != nil {
		return errors.NewInternal("删除文件记录失败", err)
	}
	return nil
}
