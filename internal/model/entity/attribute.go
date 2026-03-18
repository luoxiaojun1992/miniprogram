package entity

import "time"

// Attribute represents the attributes table.
type Attribute struct {
	ID        uint      `gorm:"primarykey;comment:属性ID" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:64;not null;comment:属性名称" json:"name"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
