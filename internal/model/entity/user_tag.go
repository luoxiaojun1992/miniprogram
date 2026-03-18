package entity

import (
	"time"
)

// User represents the users table.
// UserTag represents the user_tags table.
type UserTag struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint64    `gorm:"comment:用户ID" json:"user_id"`
	TagName   string    `gorm:"size:32;comment:标签名" json:"tag_name"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
}
