package entity

import "time"

const (
	// AttributeTypeString means attribute value is stored in string column.
	AttributeTypeString int8 = 1
	// AttributeTypeBigInt means attribute value is stored in bigint column.
	AttributeTypeBigInt int8 = 2
)

// Attribute represents the attributes table.
type Attribute struct {
	ID        uint      `gorm:"primarykey;comment:属性ID" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:64;not null;comment:属性名称" json:"name"`
	Type      int8      `gorm:"default:1;comment:1字符串 2BigInt" json:"type"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
