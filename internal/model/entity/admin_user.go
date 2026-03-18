package entity

import (
	"time"
)

// AdminUser represents the admin_users table.
type AdminUser struct {
	ID           uint64     `gorm:"primarykey" json:"id"`
	UserID       uint64     `gorm:"not null;comment:关联用户ID" json:"user_id"`
	Email        string     `gorm:"uniqueIndex;size:128;comment:邮箱" json:"email"`
	PasswordHash string     `gorm:"size:255;comment:密码哈希" json:"-"`
	LastLoginAt  *time.Time `gorm:"comment:最后登录时间" json:"last_login_at,omitempty"`
}
