package entity

import "time"

// Role represents the roles table.
type Role struct {
	ID          uint         `gorm:"primarykey" json:"id"`
	Name        string       `gorm:"size:64;not null;comment:角色名称" json:"name"`
	Description string       `gorm:"size:255;comment:描述" json:"description"`
	ParentID    uint         `gorm:"default:0;comment:父角色ID" json:"parent_id"`
	Level       int8         `gorm:"default:1;comment:层级" json:"level"`
	IsBuiltin   int8         `gorm:"default:0;comment:是否内置" json:"is_builtin"`
	CreatedAt   time.Time    `gorm:"comment:创建时间" json:"created_at"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}
