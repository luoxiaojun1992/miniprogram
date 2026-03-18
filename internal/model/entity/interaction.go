package entity

import "time"

// ContentPermission represents the content_permissions table.
type ContentPermission struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	ContentType    int8      `gorm:"comment:1文章 2课程" json:"content_type"`
	ContentID      uint64    `gorm:"comment:内容ID" json:"content_id"`
	RoleID         *uint     `gorm:"comment:角色ID null表示公开" json:"role_id,omitempty"`
	PermissionType int8      `gorm:"default:1;comment:1查看 2编辑" json:"permission_type"`
	CreatedAt      time.Time `gorm:"comment:创建时间" json:"created_at"`
	RoleName       string    `gorm:"-" json:"role_name,omitempty"`
}
