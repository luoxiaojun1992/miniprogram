package entity

import "time"

// AuditLog represents the audit_logs table.
type AuditLog struct {
	ID          uint64    `gorm:"primarykey" json:"id"`
	UserID      uint64    `gorm:"comment:操作用户ID" json:"user_id"`
	Username    string    `gorm:"size:64;comment:操作人昵称" json:"username"`
	Action      string    `gorm:"size:64;comment:操作类型" json:"action"`
	Module      string    `gorm:"size:64;comment:操作模块" json:"module"`
	Description string    `gorm:"type:text;comment:描述" json:"description"`
	IPAddress   string    `gorm:"size:45;comment:IP地址" json:"ip_address"`
	UserAgent   string    `gorm:"size:255;comment:UserAgent" json:"user_agent"`
	RequestData string    `gorm:"type:json;comment:请求数据" json:"request_data"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"created_at"`
}
