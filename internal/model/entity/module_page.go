package entity

import "time"

// ModulePage represents the module_pages table.
type ModulePage struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	ModuleID    uint      `gorm:"comment:模块ID" json:"module_id"`
	Title       string    `gorm:"size:128;comment:标题" json:"title"`
	Content     string    `gorm:"type:longtext;comment:富文本内容" json:"content"`
	ContentType int8      `gorm:"default:1;comment:1富文本 2HTML" json:"content_type"`
	SortOrder   int       `gorm:"default:0;comment:排序" json:"sort_order"`
	Status      int8      `gorm:"default:1;comment:状态" json:"status"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
