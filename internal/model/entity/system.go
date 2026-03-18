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
