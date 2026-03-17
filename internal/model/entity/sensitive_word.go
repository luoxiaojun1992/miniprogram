package entity

import "time"

// SensitiveWord represents the sensitive_words table.
type SensitiveWord struct {
	ID        uint64    `gorm:"primarykey" json:"id"`
	Word      string    `gorm:"size:128;uniqueIndex;not null;comment:敏感词" json:"word"`
	Status    int8      `gorm:"default:1;comment:0禁用 1启用" json:"status"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
