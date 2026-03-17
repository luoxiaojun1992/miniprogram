package entity

import "time"

// Banner represents the banners table.
type Banner struct {
	ID          uint64    `gorm:"primarykey" json:"id"`
	Title       string    `gorm:"size:128;comment:标题" json:"title"`
	ImageFileID *uint64   `gorm:"comment:图片文件ID" json:"image_file_id,omitempty"`
	LinkURL     string    `gorm:"size:255;comment:跳转链接" json:"link_url"`
	SortOrder   int       `gorm:"default:0;comment:排序" json:"sort_order"`
	Status      int8      `gorm:"default:1;comment:状态" json:"status"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
