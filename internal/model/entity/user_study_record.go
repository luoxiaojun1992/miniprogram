package entity

import "time"

// ContentPermission represents the content_permissions table.
// UserStudyRecord represents the user_study_records table.
type UserStudyRecord struct {
	ID          uint64     `gorm:"primarykey" json:"id"`
	UserID      uint64     `gorm:"comment:用户ID" json:"user_id"`
	CourseID    uint64     `gorm:"comment:课程ID" json:"course_id"`
	UnitID      uint64     `gorm:"comment:单元ID" json:"unit_id"`
	Progress    uint       `gorm:"default:0;comment:学习进度秒" json:"progress"`
	Status      int8       `gorm:"default:0;comment:0未开始 1学习中 2已完成" json:"status"`
	LastStudyAt *time.Time `gorm:"comment:最后学习时间" json:"last_study_at,omitempty"`
	CreatedAt   time.Time  `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"comment:更新时间" json:"updated_at"`
}
