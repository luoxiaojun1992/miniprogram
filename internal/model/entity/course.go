package entity

import "time"

// Course represents the courses table.
type Course struct {
	ID                uint64       `gorm:"primarykey" json:"id"`
	Title             string       `gorm:"size:200;not null;comment:标题" json:"title"`
	Description       string       `gorm:"type:text;comment:描述" json:"description"`
	CoverImage        string       `gorm:"size:255;comment:封面图" json:"cover_image"`
	VideoFileID       *uint64      `gorm:"comment:视频文件ID" json:"video_file_id,omitempty"`
	Duration          uint         `gorm:"comment:总课时分钟" json:"duration"`
	AuthorID          uint64       `gorm:"comment:作者ID" json:"author_id"`
	ModuleID          uint         `gorm:"comment:模块ID" json:"module_id"`
	Status            int8         `gorm:"default:0;comment:0草稿 1已发布 2定时发布" json:"status"`
	PublishTime       *time.Time   `gorm:"comment:发布时间" json:"publish_time,omitempty"`
	Price             float64      `gorm:"type:decimal(10,2);default:0.00;comment:价格" json:"price"`
	ViewCount         uint         `gorm:"default:0;comment:浏览量" json:"view_count"`
	LikeCount         uint         `gorm:"default:0;comment:点赞数" json:"like_count"`
	CollectCount      uint         `gorm:"default:0;comment:收藏数" json:"collect_count"`
	CommentCount      uint         `gorm:"default:0;comment:评论数" json:"comment_count"`
	ShareCount        uint         `gorm:"default:0;comment:分享数" json:"share_count"`
	StudyCount        uint         `gorm:"default:0;comment:学习人数" json:"study_count"`
	SortOrder         int          `gorm:"default:0;comment:排序" json:"sort_order"`
	CreatedAt         time.Time    `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt         time.Time    `gorm:"comment:更新时间" json:"updated_at"`
	Author            *User        `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Units             []CourseUnit `gorm:"foreignKey:CourseID" json:"units,omitempty"`
	AttachmentFileIDs []uint64     `gorm:"-" json:"attachment_file_ids,omitempty"`
}

// CourseUnit represents the course_units table.
type CourseUnit struct {
	ID          uint64    `gorm:"primarykey" json:"id"`
	CourseID    uint64    `gorm:"comment:课程ID" json:"course_id"`
	Title       string    `gorm:"size:200;comment:标题" json:"title"`
	VideoFileID *uint64   `gorm:"comment:视频文件ID" json:"video_file_id,omitempty"`
	Duration    uint      `gorm:"comment:课时分钟" json:"duration"`
	SortOrder   int       `gorm:"default:0;comment:排序" json:"sort_order"`
	Status      int8      `gorm:"default:1;comment:状态" json:"status"`
	CreatedAt   time.Time `gorm:"comment:创建时间" json:"created_at"`
}
