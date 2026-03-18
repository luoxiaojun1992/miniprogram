package entity

import "time"

// Role represents the roles table.
// Permission represents the permissions table.
type Permission struct {
	ID        uint          `gorm:"primarykey" json:"id"`
	Name      string        `gorm:"size:64;not null;comment:权限名称" json:"name"`
	Code      string        `gorm:"uniqueIndex;size:128;not null;comment:权限编码" json:"code"`
	Type      int8          `gorm:"default:1;comment:1菜单 2按钮 3接口" json:"type"`
	ParentID  uint          `gorm:"default:0;comment:父权限ID" json:"parent_id"`
	Level     int8          `gorm:"default:1;comment:层级" json:"level"`
	IsBuiltin int8          `gorm:"default:0;comment:是否内置" json:"is_builtin"`
	CreatedAt time.Time     `gorm:"comment:创建时间" json:"created_at"`
	Children  []*Permission `gorm:"-" json:"children,omitempty"`
}
