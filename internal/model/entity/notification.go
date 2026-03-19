package entity

import "time"

// Notification represents the notifications table.
type Notification struct {
	ID        uint64    `gorm:"primarykey" json:"id"`
	UserID    *uint64   `gorm:"comment:用户ID null为全站广播" json:"user_id,omitempty"`
	Type      int8      `gorm:"default:1;comment:1系统通知 2评论回复 3学习提醒 4点赞通知 5关注通知" json:"type"`
	Title     string    `gorm:"size:128;comment:标题" json:"title"`
	Content   string    `gorm:"type:text;comment:内容" json:"content"`
	IsRead    int8      `gorm:"default:0;comment:是否已读" json:"is_read"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
}
