package entity

import "time"

// Like represents the likes table.
type Like struct {
	ID          uint64    `gorm:"primarykey" json:"id"`
	UserID      uint64    `gorm:"comment:用户ID" json:"user_id"`
	ContentType int8      `gorm:"comment:1文章 2课程" json:"content_type"`
	ContentID   uint64    `gorm:"comment:内容ID" json:"content_id"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"created_at"`
}
