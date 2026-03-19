package service

import (
	"context"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
)

// AuthService handles authentication operations.
type AuthService interface {
	WechatLogin(ctx context.Context, req *dto.WechatLoginRequest) (*dto.LoginResponseData, error)
	AdminLogin(ctx context.Context, req *dto.AdminLoginRequest) (*dto.LoginResponseData, error)
	RefreshToken(ctx context.Context, userID uint64, userType int8) (*dto.LoginResponseData, error)
}

// UserService handles user operations.
type UserService interface {
	GetProfile(ctx context.Context, userID uint64) (*entity.User, error)
	UpdateProfile(ctx context.Context, userID uint64, req *dto.UserProfileUpdateRequest) error
	GetPermissions(ctx context.Context, userID uint64) (roles, permissions []string, err error)
	List(ctx context.Context, page, pageSize int, keyword string, userType *int8) ([]*entity.User, int64, error)
	GetByID(ctx context.Context, id uint64) (*entity.User, error)
	CreateAdminUser(ctx context.Context, req *dto.CreateAdminUserRequest) (uint64, error)
	UpdateUser(ctx context.Context, id uint64, req *dto.UpdateUserRequest, operatorID uint64) error
	DeleteUser(ctx context.Context, id uint64) error
	AssignRoles(ctx context.Context, userID uint64, req *dto.AssignRolesRequest) error
	AddTag(ctx context.Context, userID uint64, req *dto.AddTagRequest) (uint, error)
	DeleteTag(ctx context.Context, userID, tagID uint64) error
}

// RoleService handles role operations.
type RoleService interface {
	List(ctx context.Context) ([]*entity.Role, error)
	GetByID(ctx context.Context, id uint) (*entity.Role, error)
	Create(ctx context.Context, req *dto.CreateRoleRequest) (uint, error)
	Update(ctx context.Context, id uint, req *dto.UpdateRoleRequest) error
	Delete(ctx context.Context, id uint) error
}

// PermissionService handles permission operations.
type PermissionService interface {
	GetTree(ctx context.Context) ([]*entity.Permission, error)
}

// ModuleService handles module operations.
type ModuleService interface {
	List(ctx context.Context, status *int8) ([]*entity.Module, error)
	Create(ctx context.Context, req *dto.CreateModuleRequest) (uint, error)
	Update(ctx context.Context, id uint, req *dto.CreateModuleRequest) error
	Delete(ctx context.Context, id uint) error
	GetPages(ctx context.Context, moduleID uint) ([]*entity.ModulePage, error)
	CreatePage(ctx context.Context, moduleID uint, req *dto.CreateModulePageRequest) (uint, error)
	UpdatePage(ctx context.Context, moduleID, pageID uint, req *dto.CreateModulePageRequest) error
	DeletePage(ctx context.Context, moduleID, pageID uint) error
}

// BannerService handles banner operations.
type BannerService interface {
	List(ctx context.Context) ([]*entity.Banner, error)
	AdminList(ctx context.Context, status *int8) ([]*entity.Banner, error)
	Create(ctx context.Context, req *dto.CreateBannerRequest) (uint64, error)
	Update(ctx context.Context, id uint64, req *dto.CreateBannerRequest) error
	Delete(ctx context.Context, id uint64) error
}

// ArticleService handles article operations.
type ArticleService interface {
	List(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, sort string, userID *uint64) ([]*entity.Article, int64, error)
	GetByID(ctx context.Context, id uint64, userID *uint64) (*entity.Article, error)
	AdminList(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8) ([]*entity.Article, int64, error)
	AdminGetByID(ctx context.Context, id uint64) (*entity.Article, error)
	Create(ctx context.Context, req *dto.CreateArticleRequest, authorID uint64) (uint64, error)
	Update(ctx context.Context, id uint64, req *dto.UpdateArticleRequest) error
	Delete(ctx context.Context, id uint64) error
	Publish(ctx context.Context, id uint64, req *dto.PublishArticleRequest) error
	Pin(ctx context.Context, id uint64, req *dto.PinArticleRequest) error
	Copy(ctx context.Context, id uint64, authorID uint64) (uint64, error)
}

// CourseService handles course operations.
type CourseService interface {
	List(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, isFree *bool, userID *uint64) ([]*entity.Course, int64, error)
	GetByID(ctx context.Context, id uint64, userID *uint64) (*entity.Course, error)
	AdminList(ctx context.Context, page, pageSize int, keyword string, status *int8) ([]*entity.Course, int64, error)
	AdminGetByID(ctx context.Context, id uint64) (*entity.Course, error)
	Create(ctx context.Context, req *dto.CreateCourseRequest, authorID uint64) (uint64, error)
	Update(ctx context.Context, id uint64, req *dto.UpdateCourseRequest) error
	Delete(ctx context.Context, id uint64) error
	Publish(ctx context.Context, id uint64, req *dto.PublishCourseRequest) error
	Pin(ctx context.Context, id uint64, req *dto.PinCourseRequest) error
	Copy(ctx context.Context, id uint64, authorID uint64) (uint64, error)
	GetUnits(ctx context.Context, courseID uint64) ([]*entity.CourseUnit, error)
	CreateUnit(ctx context.Context, courseID uint64, req *dto.CreateCourseUnitRequest) (uint64, error)
	UpdateUnit(ctx context.Context, courseID, unitID uint64, req *dto.CreateCourseUnitRequest) error
	DeleteUnit(ctx context.Context, courseID, unitID uint64) error
}

// StudyRecordService handles study record operations.
type StudyRecordService interface {
	List(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.UserStudyRecord, int64, error)
	Update(ctx context.Context, userID uint64, req *dto.UpdateStudyRecordRequest) error
}

// CollectionService handles collection operations.
type CollectionService interface {
	List(ctx context.Context, userID uint64, page, pageSize int, contentType *int8) ([]*entity.Collection, int64, error)
	Add(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
	Remove(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
}

// LikeService handles like operations.
type LikeService interface {
	Add(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
	Remove(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
}

// FollowService handles follow operations.
type FollowService interface {
	Add(ctx context.Context, followerID, followedID uint64) error
	Remove(ctx context.Context, followerID, followedID uint64) error
}

// CommentService handles comment operations.
type CommentService interface {
	List(ctx context.Context, contentType int8, contentID uint64, page, pageSize int) ([]*entity.Comment, int64, error)
	Create(ctx context.Context, userID uint64, contentType int8, contentID uint64, req *dto.CreateCommentRequest) (*entity.Comment, error)
	AdminList(ctx context.Context, page, pageSize int, status *int8) ([]*entity.Comment, int64, error)
	Audit(ctx context.Context, id uint64, req *dto.AuditCommentRequest) error
	Delete(ctx context.Context, id uint64) error
}

// NotificationService handles notification operations.
type NotificationService interface {
	List(ctx context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, int64, error)
	MarkRead(ctx context.Context, id uint64) error
	MarkAllRead(ctx context.Context, userID uint64) error
	Send(ctx context.Context, notification *entity.Notification) error
}

// WechatConfigService handles wechat config operations.
type WechatConfigService interface {
	Get(ctx context.Context) (*entity.WechatConfig, error)
	Update(ctx context.Context, req *dto.UpdateWechatConfigRequest) error
}

// AuditLogService handles audit log operations.
type AuditLogService interface {
	List(ctx context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error)
	Log(ctx context.Context, log *entity.AuditLog)
}

// LogConfigService handles log config operations.
type LogConfigService interface {
	Get(ctx context.Context) (*entity.LogConfig, error)
	Update(ctx context.Context, req *dto.UpdateLogConfigRequest) error
}

// AttributeService handles attribute operations.
type AttributeService interface {
	List(ctx context.Context) ([]*entity.Attribute, error)
	Create(ctx context.Context, req *dto.CreateAttributeRequest) (uint, error)
	Update(ctx context.Context, id uint, req *dto.UpdateAttributeRequest) error
	Delete(ctx context.Context, id uint) error
	ListUserAttributes(ctx context.Context, userID uint64) ([]*entity.UserAttribute, error)
	SetUserAttribute(ctx context.Context, userID uint64, req *dto.SetUserAttributeRequest) error
	DeleteUserAttribute(ctx context.Context, userID uint64, attributeID uint) error
}
