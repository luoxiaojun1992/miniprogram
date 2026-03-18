package entity

import "time"

// ArticleAttachment represents article and attachment relation.
type ArticleAttachment struct {
	ID        uint64    `gorm:"primarykey" json:"id"`
	ArticleID uint64    `gorm:"not null;index:idx_article_sort,priority:1;comment:文章ID" json:"article_id"`
	FileID    uint64    `gorm:"not null;index;comment:文件ID" json:"file_id"`
	SortOrder int       `gorm:"default:0;index:idx_article_sort,priority:2;comment:排序" json:"sort_order"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
}
