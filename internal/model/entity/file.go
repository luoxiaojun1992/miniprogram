package entity

import "time"

// File represents uploaded file metadata and bound COS object information.
type File struct {
	ID        uint64    `gorm:"primarykey" json:"id"`
	Key       string    `gorm:"size:255;not null;comment:COS对象Key" json:"key"`
	Filename  string    `gorm:"size:255;not null;comment:原始文件名" json:"filename"`
	Usage     string    `gorm:"size:32;not null;comment:文件用途" json:"usage"`
	Category  string    `gorm:"size:32;not null;comment:文件分类" json:"category"`
	Business  string    `gorm:"size:64;comment:业务类型" json:"business,omitempty"`
	StaticURL string    `gorm:"size:512;comment:静态访问地址" json:"static_url,omitempty"`
	CreatedBy uint64    `gorm:"comment:创建人ID" json:"created_by"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
}

// ArticleAttachment represents article and attachment relation.
type ArticleAttachment struct {
	ID        uint64    `gorm:"primarykey" json:"id"`
	ArticleID uint64    `gorm:"not null;index:idx_article_sort,priority:1;comment:文章ID" json:"article_id"`
	FileID    uint64    `gorm:"not null;index;comment:文件ID" json:"file_id"`
	SortOrder int       `gorm:"default:0;index:idx_article_sort,priority:2;comment:排序" json:"sort_order"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
}

// CourseAttachment represents course and attachment relation.
type CourseAttachment struct {
	ID        uint64    `gorm:"primarykey" json:"id"`
	CourseID  uint64    `gorm:"not null;index:idx_course_sort,priority:1;comment:课程ID" json:"course_id"`
	FileID    uint64    `gorm:"not null;index;comment:文件ID" json:"file_id"`
	SortOrder int       `gorm:"default:0;index:idx_course_sort,priority:2;comment:排序" json:"sort_order"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
}
