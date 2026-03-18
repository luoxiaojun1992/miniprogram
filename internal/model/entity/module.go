package entity

import "time"

// Module represents the modules table.
type Module struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Title       string    `gorm:"size:128;not null;comment:标题" json:"title"`
	Description string    `gorm:"type:text;comment:描述" json:"description"`
	SortOrder   int       `gorm:"default:0;comment:排序" json:"sort_order"`
	Status      int8      `gorm:"default:1;comment:0禁用 1启用" json:"status"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
