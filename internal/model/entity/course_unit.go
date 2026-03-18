package entity

import "time"

// CourseUnit represents the course_units table.
type CourseUnit struct {
	ID          uint64    `gorm:"primarykey" json:"id"`
	CourseID    uint64    `gorm:"comment:课程ID" json:"course_id"`
	Title       string    `gorm:"size:200;comment:标题" json:"title"`
	VideoFileID *uint64   `gorm:"comment:视频文件ID" json:"video_file_id,omitempty"`
	Duration    uint      `gorm:"comment:课时分钟" json:"duration"`
	SortOrder   int       `gorm:"default:0;comment:排序" json:"sort_order"`
	Status      int8      `gorm:"default:1;comment:状态" json:"status"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"created_at"`
}
