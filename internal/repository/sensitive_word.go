package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

type sensitiveWordRepository struct {
	db *gorm.DB
}

// NewSensitiveWordRepository creates a new SensitiveWordRepository.
func NewSensitiveWordRepository(db *gorm.DB) SensitiveWordRepository {
	return &sensitiveWordRepository{db: db}
}

func (r *sensitiveWordRepository) ListEnabledWords(ctx context.Context) ([]string, error) {
	var rows []*entity.SensitiveWord
	if err := r.db.WithContext(ctx).
		Where("status = ?", 1).
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return nil, errors.NewInternal("查询敏感词失败", err)
	}
	words := make([]string, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.Word == "" {
			continue
		}
		words = append(words, row.Word)
	}
	return words, nil
}
