package repository

import (
	"context"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	GetByID(ctx context.Context, id uint64) (*entity.User, error)
	GetByOpenID(ctx context.Context, openID string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, page, pageSize int, keyword string, userType, status *int8) ([]*entity.User, int64, error)
	GetWithTags(ctx context.Context, id uint64) (*entity.User, error)
}

// AdminUserRepository defines the interface for admin user data access.
type AdminUserRepository interface {
	GetByEmail(ctx context.Context, email string) (*entity.AdminUser, error)
	GetByUserID(ctx context.Context, userID uint64) (*entity.AdminUser, error)
	Create(ctx context.Context, admin *entity.AdminUser) error
	UpdateLastLogin(ctx context.Context, id uint64) error
}

// UserTagRepository defines the interface for user tag data access.
type UserTagRepository interface {
	GetByUserID(ctx context.Context, userID uint64) ([]*entity.UserTag, error)
	Create(ctx context.Context, tag *entity.UserTag) error
	Delete(ctx context.Context, id uint) error
}

// RoleRepository defines the interface for role data access.
type RoleRepository interface {
	GetByID(ctx context.Context, id uint) (*entity.Role, error)
	GetWithPermissions(ctx context.Context, id uint) (*entity.Role, error)
	List(ctx context.Context) ([]*entity.Role, error)
	Create(ctx context.Context, role *entity.Role) error
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id uint) error
	AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error
	GetUserRoles(ctx context.Context, userID uint64) ([]*entity.Role, error)
	AssignUserRoles(ctx context.Context, userID uint64, roleIDs []uint) error
	HasUsers(ctx context.Context, roleID uint) (bool, error)
}

// PermissionRepository defines the interface for permission data access.
type PermissionRepository interface {
	List(ctx context.Context) ([]*entity.Permission, error)
	GetByID(ctx context.Context, id uint) (*entity.Permission, error)
	GetUserPermissions(ctx context.Context, userID uint64) ([]*entity.Permission, error)
}

// ModuleRepository defines the interface for module data access.
type ModuleRepository interface {
	GetByID(ctx context.Context, id uint) (*entity.Module, error)
	List(ctx context.Context, status *int8) ([]*entity.Module, error)
	Create(ctx context.Context, module *entity.Module) error
	Update(ctx context.Context, module *entity.Module) error
	Delete(ctx context.Context, id uint) error
}

// ModulePageRepository defines the interface for module page data access.
type ModulePageRepository interface {
	GetByID(ctx context.Context, id uint) (*entity.ModulePage, error)
	ListByModuleID(ctx context.Context, moduleID uint) ([]*entity.ModulePage, error)
	Create(ctx context.Context, page *entity.ModulePage) error
	Update(ctx context.Context, page *entity.ModulePage) error
	Delete(ctx context.Context, id uint) error
}

// ArticleRepository defines the interface for article data access.
type ArticleRepository interface {
	GetByID(ctx context.Context, id uint64) (*entity.Article, error)
	List(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8, sort string) ([]*entity.Article, int64, error)
	Create(ctx context.Context, article *entity.Article) error
	Update(ctx context.Context, article *entity.Article) error
	Delete(ctx context.Context, id uint64) error
	IncrViewCount(ctx context.Context, id uint64) error
}

// CourseRepository defines the interface for course data access.
type CourseRepository interface {
	GetByID(ctx context.Context, id uint64) (*entity.Course, error)
	List(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8, isFree *bool) ([]*entity.Course, int64, error)
	Create(ctx context.Context, course *entity.Course) error
	Update(ctx context.Context, course *entity.Course) error
	Delete(ctx context.Context, id uint64) error
	IncrViewCount(ctx context.Context, id uint64) error
}

// CourseUnitRepository defines the interface for course unit data access.
type CourseUnitRepository interface {
	GetByID(ctx context.Context, id uint64) (*entity.CourseUnit, error)
	ListByCourseID(ctx context.Context, courseID uint64) ([]*entity.CourseUnit, error)
	Create(ctx context.Context, unit *entity.CourseUnit) error
	Update(ctx context.Context, unit *entity.CourseUnit) error
	Delete(ctx context.Context, id uint64) error
}

// ContentPermissionRepository defines the interface for content permission data access.
type ContentPermissionRepository interface {
	GetByContent(ctx context.Context, contentType int8, contentID uint64) ([]*entity.ContentPermission, error)
	SetContentPermissions(ctx context.Context, contentType int8, contentID uint64, roleIDs []uint) error
}

// StudyRecordRepository defines the interface for study record data access.
type StudyRecordRepository interface {
	GetByUserAndUnit(ctx context.Context, userID, unitID uint64) (*entity.UserStudyRecord, error)
	ListByUser(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.UserStudyRecord, int64, error)
	Upsert(ctx context.Context, record *entity.UserStudyRecord) error
}

// CollectionRepository defines the interface for collection data access.
type CollectionRepository interface {
	Get(ctx context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Collection, error)
	List(ctx context.Context, userID uint64, page, pageSize int, contentType *int8) ([]*entity.Collection, int64, error)
	Create(ctx context.Context, collection *entity.Collection) error
	Delete(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
}

// LikeRepository defines the interface for like data access.
type LikeRepository interface {
	Get(ctx context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Like, error)
	Create(ctx context.Context, like *entity.Like) error
	Delete(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
}

// CommentRepository defines the interface for comment data access.
type CommentRepository interface {
	GetByID(ctx context.Context, id uint64) (*entity.Comment, error)
	List(ctx context.Context, contentType int8, contentID uint64, page, pageSize int) ([]*entity.Comment, int64, error)
	ListAdmin(ctx context.Context, page, pageSize int, status *int8) ([]*entity.Comment, int64, error)
	Create(ctx context.Context, comment *entity.Comment) error
	UpdateStatus(ctx context.Context, id uint64, status int8) error
	Delete(ctx context.Context, id uint64) error
}

// NotificationRepository defines the interface for notification data access.
type NotificationRepository interface {
	GetByID(ctx context.Context, id uint64) (*entity.Notification, error)
	List(ctx context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, error)
	UnreadCount(ctx context.Context, userID uint64) (int64, error)
	MarkRead(ctx context.Context, id uint64) error
	MarkAllRead(ctx context.Context, userID uint64) error
}

// WechatConfigRepository defines the interface for wechat config data access.
type WechatConfigRepository interface {
	Get(ctx context.Context) (*entity.WechatConfig, error)
	Update(ctx context.Context, cfg *entity.WechatConfig) error
}

// AuditLogRepository defines the interface for audit log data access.
type AuditLogRepository interface {
	GetByID(ctx context.Context, id uint64) (*entity.AuditLog, error)
	List(ctx context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error)
	Create(ctx context.Context, log *entity.AuditLog) error
}

// LogConfigRepository defines the interface for log config data access.
type LogConfigRepository interface {
	Get(ctx context.Context) (*entity.LogConfig, error)
	Update(ctx context.Context, cfg *entity.LogConfig) error
}

// AttributeRepository defines the interface for attribute data access.
type AttributeRepository interface {
	GetByID(ctx context.Context, id uint) (*entity.Attribute, error)
	List(ctx context.Context) ([]*entity.Attribute, error)
	Create(ctx context.Context, attr *entity.Attribute) error
	Update(ctx context.Context, attr *entity.Attribute) error
	Delete(ctx context.Context, id uint) error
}

// UserAttributeRepository defines the interface for user attribute data access.
type UserAttributeRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]*entity.UserAttribute, error)
	Upsert(ctx context.Context, ua *entity.UserAttribute) error
	Delete(ctx context.Context, userID uint64, attributeID uint) error
}

// SensitiveWordRepository defines the interface for sensitive word data access.
type SensitiveWordRepository interface {
	ListEnabledWords(ctx context.Context) ([]string, error)
}
