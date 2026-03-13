package entity

import (
	"time"

	"gorm.io/gorm"
)

// User represents the users table.
type User struct {
	ID            uint64         `gorm:"primarykey;comment:用户ID" json:"id"`
	OpenID        string         `gorm:"uniqueIndex;size:64;comment:微信OpenID" json:"openid,omitempty"`
	UnionID       string         `gorm:"size:64;comment:微信UnionID" json:"unionid,omitempty"`
	Nickname      string         `gorm:"size:64;comment:用户昵称" json:"nickname"`
	AvatarURL     string         `gorm:"size:255;comment:头像URL" json:"avatar_url"`
	UserType      int8           `gorm:"default:1;comment:1前台用户 2普通管理员 3系统管理员" json:"user_type"`
	Status        int8           `gorm:"default:1;comment:0冻结 1正常" json:"status"`
	FreezeEndTime *time.Time     `gorm:"comment:冻结结束时间" json:"freeze_end_time,omitempty"`
	CreatedAt     time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
	Tags          []UserTag      `gorm:"foreignKey:UserID" json:"tags,omitempty"`
	AdminInfo     *AdminUser     `gorm:"foreignKey:UserID" json:"admin_info,omitempty"`
}

// AdminUser represents the admin_users table.
type AdminUser struct {
	ID           uint64     `gorm:"primarykey" json:"id"`
	UserID       uint64     `gorm:"not null;comment:关联用户ID" json:"user_id"`
	Email        string     `gorm:"uniqueIndex;size:128;comment:邮箱" json:"email"`
	PasswordHash string     `gorm:"size:255;comment:密码哈希" json:"-"`
	LastLoginAt  *time.Time `gorm:"comment:最后登录时间" json:"last_login_at,omitempty"`
}

// UserTag represents the user_tags table.
type UserTag struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint64    `gorm:"comment:用户ID" json:"user_id"`
	TagName   string    `gorm:"size:32;comment:标签名" json:"tag_name"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
}
