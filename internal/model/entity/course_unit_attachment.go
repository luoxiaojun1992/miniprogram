package entity

import "time"

// CourseUnitAttachment represents course unit and attachment relation.
type CourseUnitAttachment struct {
	ID           uint64    `gorm:"primarykey" json:"id"`
	UnitID       uint64    `gorm:"not null;index:idx_unit_sort,priority:1;comment:单元ID" json:"unit_id"`
	FileID       uint64    `gorm:"not null;index;comment:文件ID" json:"file_id"`
	PermissionID *uint     `gorm:"comment:关联权限ID" json:"permission_id,omitempty"`
	SortOrder    int       `gorm:"default:0;index:idx_unit_sort,priority:2;comment:排序" json:"sort_order"`
	CreatedAt    time.Time `gorm:"comment:创建时间" json:"created_at"`
}
