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

// UserStudyRecord represents the user_study_records table.
type UserStudyRecord struct {
	ID          uint64     `gorm:"primarykey" json:"id"`
	UserID      uint64     `gorm:"comment:用户ID" json:"user_id"`
	CourseID    uint64     `gorm:"comment:课程ID" json:"course_id"`
	UnitID      uint64     `gorm:"comment:单元ID" json:"unit_id"`
	Progress    uint       `gorm:"default:0;comment:学习进度秒" json:"progress"`
	Status      int8       `gorm:"default:0;comment:0未开始 1学习中 2已完成" json:"status"`
	LastStudyAt *time.Time `gorm:"comment:最后学习时间" json:"last_study_at,omitempty"`
	CreatedAt   time.Time  `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"comment:更新时间" json:"updated_at"`
}

// Collection represents the collections table.
type Collection struct {
	ID           uint64    `gorm:"primarykey" json:"id"`
	UserID       uint64    `gorm:"comment:用户ID" json:"user_id"`
	ContentType  int8      `gorm:"comment:1文章 2课程" json:"content_type"`
	ContentID    uint64    `gorm:"comment:内容ID" json:"content_id"`
	CreatedAt    time.Time `gorm:"comment:创建时间" json:"created_at"`
	ContentTitle string    `gorm:"-" json:"content_title,omitempty"`
	CoverImage   string    `gorm:"-" json:"cover_image,omitempty"`
}

// Like represents the likes table.
type Like struct {
	ID          uint64    `gorm:"primarykey" json:"id"`
	UserID      uint64    `gorm:"comment:用户ID" json:"user_id"`
	ContentType int8      `gorm:"comment:1文章 2课程" json:"content_type"`
	ContentID   uint64    `gorm:"comment:内容ID" json:"content_id"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"created_at"`
}

// Comment represents the comments table.
type Comment struct {
	ID          uint64    `gorm:"primarykey" json:"id"`
	UserID      uint64    `gorm:"comment:用户ID" json:"user_id"`
	ContentType int8      `gorm:"comment:1文章 2课程" json:"content_type"`
	ContentID   uint64    `gorm:"comment:内容ID" json:"content_id"`
	ParentID    uint64    `gorm:"default:0;comment:回复评论ID" json:"parent_id"`
	Content     string    `gorm:"type:text;not null;comment:内容" json:"content"`
	Status      int8      `gorm:"default:1;comment:0待审核 1通过 2拒绝" json:"status"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"created_at"`
	User        *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Replies     []Comment `gorm:"-" json:"replies,omitempty"`
}
