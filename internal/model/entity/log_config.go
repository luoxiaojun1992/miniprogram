package entity

import "time"

// WechatConfig represents the wechat_configs table.
// LogConfig represents the log_configs table.
type LogConfig struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	RetentionDays int       `gorm:"default:90;comment:日志保留天数" json:"retention_days"`
	UpdatedAt     time.Time `gorm:"comment:更新时间" json:"updated_at"`
}
