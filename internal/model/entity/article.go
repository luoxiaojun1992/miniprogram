package entity

import "time"

// Article represents the articles table.
type Article struct {
	ID                uint64     `gorm:"primarykey" json:"id"`
	Title             string     `gorm:"size:200;not null;comment:标题" json:"title"`
	Summary           string     `gorm:"size:500;comment:摘要" json:"summary"`
	Content           string     `gorm:"type:longtext;comment:内容" json:"content"`
	ContentType       int8       `gorm:"default:1;comment:1富文本 2HTML 3Markdown" json:"content_type"`
	CoverImage        string     `gorm:"size:255;comment:封面图" json:"cover_image"`
	AuthorID          uint64     `gorm:"comment:作者ID" json:"author_id"`
	ModuleID          uint       `gorm:"comment:模块ID" json:"module_id"`
	Status            int8       `gorm:"default:0;comment:0草稿 1已发布 2定时发布" json:"status"`
	PublishTime       *time.Time `gorm:"comment:发布时间" json:"publish_time,omitempty"`
	ViewCount         uint       `gorm:"default:0;comment:浏览量" json:"view_count"`
	LikeCount         uint       `gorm:"default:0;comment:点赞数" json:"like_count"`
	CollectCount      uint       `gorm:"default:0;comment:收藏数" json:"collect_count"`
	CommentCount      uint       `gorm:"default:0;comment:评论数" json:"comment_count"`
	ShareCount        uint       `gorm:"default:0;comment:分享数" json:"share_count"`
	SortOrder         int        `gorm:"default:0;comment:排序" json:"sort_order"`
	CreatedAt         time.Time  `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"comment:更新时间" json:"updated_at"`
	Author            *User      `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	AttachmentFileIDs []uint64   `gorm:"-" json:"attachment_file_ids,omitempty"`
}
