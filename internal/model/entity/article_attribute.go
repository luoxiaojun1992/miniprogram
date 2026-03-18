package entity

import "time"

// ArticleAttribute represents article_attributes table.
type ArticleAttribute struct {
	ID          uint64     `gorm:"primarykey;comment:ID" json:"id"`
	ArticleID   uint64     `gorm:"not null;comment:文章ID" json:"article_id"`
	AttributeID uint       `gorm:"not null;comment:属性ID" json:"attribute_id"`
	ValueString string     `gorm:"size:255;not null;default:'';comment:字符串属性值" json:"value_string,omitempty"`
	ValueBigint *int64     `gorm:"comment:BigInt属性值" json:"value_bigint,omitempty"`
	CreatedAt   time.Time  `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"comment:更新时间" json:"updated_at"`
	Attribute   *Attribute `gorm:"foreignKey:AttributeID" json:"attribute,omitempty"`
}
