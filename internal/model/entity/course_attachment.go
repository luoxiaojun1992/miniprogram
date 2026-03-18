package entity

import "time"

// File represents uploaded file metadata and bound COS object information.
// CourseAttachment represents course and attachment relation.
type CourseAttachment struct {
	ID        uint64    `gorm:"primarykey" json:"id"`
	CourseID  uint64    `gorm:"not null;index:idx_course_sort,priority:1;comment:课程ID" json:"course_id"`
	FileID    uint64    `gorm:"not null;index;comment:文件ID" json:"file_id"`
	SortOrder int       `gorm:"default:0;index:idx_course_sort,priority:2;comment:排序" json:"sort_order"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
}
