package entity

import "time"

// WechatConfig represents the wechat_configs table.
type WechatConfig struct {
	ID              uint       `gorm:"primarykey" json:"id"`
	AppID           string     `gorm:"size:32;not null;comment:AppID" json:"app_id"`
	AppSecret       string     `gorm:"size:64;not null;comment:AppSecret" json:"-"`
	APIToken        string     `gorm:"size:255;comment:微信API Token" json:"api_token"`
	JSAPITicket     string     `gorm:"size:512;comment:JSAPITicket" json:"jsapi_ticket,omitempty"`
	TicketExpiresAt *time.Time `gorm:"comment:Ticket过期时间" json:"ticket_expires_at,omitempty"`
	AccessToken     string     `gorm:"size:512;comment:AccessToken" json:"-"`
	TokenExpiresAt  *time.Time `gorm:"comment:Token过期时间" json:"token_expires_at,omitempty"`
	Status          int8       `gorm:"default:1;comment:状态" json:"status"`
	CreatedAt       time.Time  `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"comment:更新时间" json:"updated_at"`
}

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

// LogConfig represents the log_configs table.
type LogConfig struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	RetentionDays int       `gorm:"default:90;comment:日志保留天数" json:"retention_days"`
	UpdatedAt     time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
