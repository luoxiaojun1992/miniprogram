package entity

import "time"

// Comment represents the comments table.
type Comment struct {
	ID          uint64    `gorm:"primarykey" json:"id"`
	UserID      uint64    `gorm:"comment:用户ID" json:"user_id"`
	ContentType int8      `gorm:"comment:1文章 2课程" json:"content_type"`
	ContentID   uint64    `gorm:"comment:内容ID" json:"content_id"`
	ParentID    uint64    `gorm:"default:0;comment:回复评论ID" json:"parent_id"`
	Content     string    `gorm:"type:text;not null;comment:内容" json:"content"`
	Status      int8      `gorm:"default:1;comment:0待审核 1通过 2拒绝" json:"status"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"created_at"`
	User        *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Replies     []Comment `gorm:"-" json:"replies,omitempty"`
}
