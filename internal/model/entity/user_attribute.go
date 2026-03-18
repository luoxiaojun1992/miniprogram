package entity

import "time"

// Attribute represents the attributes table.
// UserAttribute represents the user_attributes table.
type UserAttribute struct {
	ID          uint64     `gorm:"primarykey;comment:ID" json:"id"`
	UserID      uint64     `gorm:"not null;comment:用户ID" json:"user_id"`
	AttributeID uint       `gorm:"not null;comment:属性ID" json:"attribute_id"`
	Value       string     `gorm:"size:255;not null;default:'';comment:属性值" json:"value"`
	CreatedAt   time.Time  `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"comment:更新时间" json:"updated_at"`
	Attribute   *Attribute `gorm:"foreignKey:AttributeID" json:"attribute,omitempty"`
}
