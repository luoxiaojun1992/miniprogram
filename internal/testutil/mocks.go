// Package testutil provides mock implementations of repository and service interfaces for testing.
package testutil

import (
	"context"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
)

// ==================== Repository Mocks ====================

// MockUserRepository is a test double for repository.UserRepository.
type MockUserRepository struct {
	GetByIDFn     func(ctx context.Context, id uint64) (*entity.User, error)
	GetByOpenIDFn func(ctx context.Context, openID string) (*entity.User, error)
	CreateFn      func(ctx context.Context, user *entity.User) error
	UpdateFn      func(ctx context.Context, user *entity.User) error
	DeleteFn      func(ctx context.Context, id uint64) error
	ListFn        func(ctx context.Context, page, pageSize int, keyword string, userType, status *int8) ([]*entity.User, int64, error)
	GetWithTagsFn func(ctx context.Context, id uint64) (*entity.User, error)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint64) (*entity.User, error) {
		if m.GetByIDFn != nil {
			return m.GetByIDFn(ctx, id)
		}
		return nil, nil
}
func (m *MockUserRepository) GetByOpenID(ctx context.Context, openID string) (*entity.User, error) {
		if m.GetByOpenIDFn != nil {
			return m.GetByOpenIDFn(ctx, openID)
		}
		return nil, nil
}
func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, user)
		}
		return nil
}
func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
		if m.UpdateFn != nil {
			return m.UpdateFn(ctx, user)
		}
		return nil
}
func (m *MockUserRepository) Delete(ctx context.Context, id uint64) error {
		if m.DeleteFn != nil {
			return m.DeleteFn(ctx, id)
		}
		return nil
}
func (m *MockUserRepository) List(ctx context.Context, page, pageSize int, keyword string, userType, status *int8) ([]*entity.User, int64, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, page, pageSize, keyword, userType, status)
		}
		return nil, 0, nil
}
func (m *MockUserRepository) GetWithTags(ctx context.Context, id uint64) (*entity.User, error) {
		if m.GetWithTagsFn != nil {
			return m.GetWithTagsFn(ctx, id)
		}
		return nil, nil
}

// MockAdminUserRepository is a test double for repository.AdminUserRepository.
type MockAdminUserRepository struct {
	GetByEmailFn      func(ctx context.Context, email string) (*entity.AdminUser, error)
	GetByUserIDFn     func(ctx context.Context, userID uint64) (*entity.AdminUser, error)
	CreateFn          func(ctx context.Context, admin *entity.AdminUser) error
	UpdateLastLoginFn func(ctx context.Context, id uint64) error
}

func (m *MockAdminUserRepository) GetByEmail(ctx context.Context, email string) (*entity.AdminUser, error) {
		if m.GetByEmailFn != nil {
			return m.GetByEmailFn(ctx, email)
		}
		return nil, nil
}
func (m *MockAdminUserRepository) GetByUserID(ctx context.Context, userID uint64) (*entity.AdminUser, error) {
		if m.GetByUserIDFn != nil {
			return m.GetByUserIDFn(ctx, userID)
		}
		return nil, nil
}
func (m *MockAdminUserRepository) Create(ctx context.Context, admin *entity.AdminUser) error {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, admin)
		}
		return nil
}
func (m *MockAdminUserRepository) UpdateLastLogin(ctx context.Context, id uint64) error {
		if m.UpdateLastLoginFn != nil {
			return m.UpdateLastLoginFn(ctx, id)
		}
		return nil
}

// MockUserTagRepository is a test double for repository.UserTagRepository.
type MockUserTagRepository struct {
	GetByUserIDFn func(ctx context.Context, userID uint64) ([]*entity.UserTag, error)
	CreateFn      func(ctx context.Context, tag *entity.UserTag) error
	DeleteFn      func(ctx context.Context, id uint) error
}

func (m *MockUserTagRepository) GetByUserID(ctx context.Context, userID uint64) ([]*entity.UserTag, error) {
		if m.GetByUserIDFn != nil {
			return m.GetByUserIDFn(ctx, userID)
		}
		return nil, nil
}
func (m *MockUserTagRepository) Create(ctx context.Context, tag *entity.UserTag) error {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, tag)
		}
		return nil
}
func (m *MockUserTagRepository) Delete(ctx context.Context, id uint) error {
		if m.DeleteFn != nil {
			return m.DeleteFn(ctx, id)
		}
		return nil
}

// MockRoleRepository is a test double for repository.RoleRepository.
type MockRoleRepository struct {
	GetByIDFn          func(ctx context.Context, id uint) (*entity.Role, error)
	GetWithPermsFn     func(ctx context.Context, id uint) (*entity.Role, error)
	ListFn             func(ctx context.Context) ([]*entity.Role, error)
	CreateFn           func(ctx context.Context, role *entity.Role) error
	UpdateFn           func(ctx context.Context, role *entity.Role) error
	DeleteFn           func(ctx context.Context, id uint) error
	AssignPermsFn      func(ctx context.Context, roleID uint, permissionIDs []uint) error
	GetUserRolesFn     func(ctx context.Context, userID uint64) ([]*entity.Role, error)
	AssignUserRolesFn  func(ctx context.Context, userID uint64, roleIDs []uint) error
	HasUsersFn         func(ctx context.Context, roleID uint) (bool, error)
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id uint) (*entity.Role, error) {
		if m.GetByIDFn != nil {
			return m.GetByIDFn(ctx, id)
		}
		return nil, nil
}
func (m *MockRoleRepository) GetWithPermissions(ctx context.Context, id uint) (*entity.Role, error) {
	return m.GetWithPermsFn(ctx, id)
}
func (m *MockRoleRepository) List(ctx context.Context) ([]*entity.Role, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx)
		}
		return nil, nil
}
func (m *MockRoleRepository) Create(ctx context.Context, role *entity.Role) error {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, role)
		}
		return nil
}
func (m *MockRoleRepository) Update(ctx context.Context, role *entity.Role) error {
		if m.UpdateFn != nil {
			return m.UpdateFn(ctx, role)
		}
		return nil
}
func (m *MockRoleRepository) Delete(ctx context.Context, id uint) error {
		if m.DeleteFn != nil {
			return m.DeleteFn(ctx, id)
		}
		return nil
}
func (m *MockRoleRepository) AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	return m.AssignPermsFn(ctx, roleID, permissionIDs)
}
func (m *MockRoleRepository) GetUserRoles(ctx context.Context, userID uint64) ([]*entity.Role, error) {
		if m.GetUserRolesFn != nil {
			return m.GetUserRolesFn(ctx, userID)
		}
		return nil, nil
}
func (m *MockRoleRepository) AssignUserRoles(ctx context.Context, userID uint64, roleIDs []uint) error {
		if m.AssignUserRolesFn != nil {
			return m.AssignUserRolesFn(ctx, userID, roleIDs)
		}
		return nil
}
func (m *MockRoleRepository) HasUsers(ctx context.Context, roleID uint) (bool, error) {
		if m.HasUsersFn != nil {
			return m.HasUsersFn(ctx, roleID)
		}
		return false, nil
}

// MockPermissionRepository is a test double for repository.PermissionRepository.
type MockPermissionRepository struct {
	ListFn               func(ctx context.Context) ([]*entity.Permission, error)
	GetByIDFn            func(ctx context.Context, id uint) (*entity.Permission, error)
	GetUserPermissionsFn func(ctx context.Context, userID uint64) ([]*entity.Permission, error)
}

func (m *MockPermissionRepository) List(ctx context.Context) ([]*entity.Permission, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx)
		}
		return nil, nil
}
func (m *MockPermissionRepository) GetByID(ctx context.Context, id uint) (*entity.Permission, error) {
		if m.GetByIDFn != nil {
			return m.GetByIDFn(ctx, id)
		}
		return nil, nil
}
func (m *MockPermissionRepository) GetUserPermissions(ctx context.Context, userID uint64) ([]*entity.Permission, error) {
		if m.GetUserPermissionsFn != nil {
			return m.GetUserPermissionsFn(ctx, userID)
		}
		return nil, nil
}

// MockModuleRepository is a test double for repository.ModuleRepository.
type MockModuleRepository struct {
	GetByIDFn func(ctx context.Context, id uint) (*entity.Module, error)
	ListFn    func(ctx context.Context, status *int8) ([]*entity.Module, error)
	CreateFn  func(ctx context.Context, module *entity.Module) error
	UpdateFn  func(ctx context.Context, module *entity.Module) error
	DeleteFn  func(ctx context.Context, id uint) error
}

func (m *MockModuleRepository) GetByID(ctx context.Context, id uint) (*entity.Module, error) {
		if m.GetByIDFn != nil {
			return m.GetByIDFn(ctx, id)
		}
		return nil, nil
}
func (m *MockModuleRepository) List(ctx context.Context, status *int8) ([]*entity.Module, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, status)
		}
		return nil, nil
}
func (m *MockModuleRepository) Create(ctx context.Context, module *entity.Module) error {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, module)
		}
		return nil
}
func (m *MockModuleRepository) Update(ctx context.Context, module *entity.Module) error {
		if m.UpdateFn != nil {
			return m.UpdateFn(ctx, module)
		}
		return nil
}
func (m *MockModuleRepository) Delete(ctx context.Context, id uint) error {
		if m.DeleteFn != nil {
			return m.DeleteFn(ctx, id)
		}
		return nil
}

// MockModulePageRepository is a test double for repository.ModulePageRepository.
type MockModulePageRepository struct {
	GetByIDFn        func(ctx context.Context, id uint) (*entity.ModulePage, error)
	ListByModuleIDFn func(ctx context.Context, moduleID uint) ([]*entity.ModulePage, error)
	CreateFn         func(ctx context.Context, page *entity.ModulePage) error
	UpdateFn         func(ctx context.Context, page *entity.ModulePage) error
	DeleteFn         func(ctx context.Context, id uint) error
}

func (m *MockModulePageRepository) GetByID(ctx context.Context, id uint) (*entity.ModulePage, error) {
		if m.GetByIDFn != nil {
			return m.GetByIDFn(ctx, id)
		}
		return nil, nil
}
func (m *MockModulePageRepository) ListByModuleID(ctx context.Context, moduleID uint) ([]*entity.ModulePage, error) {
		if m.ListByModuleIDFn != nil {
			return m.ListByModuleIDFn(ctx, moduleID)
		}
		return nil, nil
}
func (m *MockModulePageRepository) Create(ctx context.Context, page *entity.ModulePage) error {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, page)
		}
		return nil
}
func (m *MockModulePageRepository) Update(ctx context.Context, page *entity.ModulePage) error {
		if m.UpdateFn != nil {
			return m.UpdateFn(ctx, page)
		}
		return nil
}
func (m *MockModulePageRepository) Delete(ctx context.Context, id uint) error {
		if m.DeleteFn != nil {
			return m.DeleteFn(ctx, id)
		}
		return nil
}

// MockArticleRepository is a test double for repository.ArticleRepository.
type MockArticleRepository struct {
	GetByIDFn       func(ctx context.Context, id uint64) (*entity.Article, error)
	ListFn          func(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8, sort string) ([]*entity.Article, int64, error)
	CreateFn        func(ctx context.Context, article *entity.Article) error
	UpdateFn        func(ctx context.Context, article *entity.Article) error
	DeleteFn        func(ctx context.Context, id uint64) error
	IncrViewCountFn func(ctx context.Context, id uint64) error
}

func (m *MockArticleRepository) GetByID(ctx context.Context, id uint64) (*entity.Article, error) {
		if m.GetByIDFn != nil {
			return m.GetByIDFn(ctx, id)
		}
		return nil, nil
}
func (m *MockArticleRepository) List(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8, sort string) ([]*entity.Article, int64, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, page, pageSize, keyword, moduleID, status, sort)
		}
		return nil, 0, nil
}
func (m *MockArticleRepository) Create(ctx context.Context, article *entity.Article) error {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, article)
		}
		return nil
}
func (m *MockArticleRepository) Update(ctx context.Context, article *entity.Article) error {
		if m.UpdateFn != nil {
			return m.UpdateFn(ctx, article)
		}
		return nil
}
func (m *MockArticleRepository) Delete(ctx context.Context, id uint64) error {
		if m.DeleteFn != nil {
			return m.DeleteFn(ctx, id)
		}
		return nil
}
func (m *MockArticleRepository) IncrViewCount(ctx context.Context, id uint64) error {
		if m.IncrViewCountFn != nil {
			return m.IncrViewCountFn(ctx, id)
		}
		return nil
}

// MockCourseRepository is a test double for repository.CourseRepository.
type MockCourseRepository struct {
	GetByIDFn       func(ctx context.Context, id uint64) (*entity.Course, error)
	ListFn          func(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8, isFree *bool) ([]*entity.Course, int64, error)
	CreateFn        func(ctx context.Context, course *entity.Course) error
	UpdateFn        func(ctx context.Context, course *entity.Course) error
	DeleteFn        func(ctx context.Context, id uint64) error
	IncrViewCountFn func(ctx context.Context, id uint64) error
}

func (m *MockCourseRepository) GetByID(ctx context.Context, id uint64) (*entity.Course, error) {
		if m.GetByIDFn != nil {
			return m.GetByIDFn(ctx, id)
		}
		return nil, nil
}
func (m *MockCourseRepository) List(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8, isFree *bool) ([]*entity.Course, int64, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, page, pageSize, keyword, moduleID, status, isFree)
		}
		return nil, 0, nil
}
func (m *MockCourseRepository) Create(ctx context.Context, course *entity.Course) error {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, course)
		}
		return nil
}
func (m *MockCourseRepository) Update(ctx context.Context, course *entity.Course) error {
		if m.UpdateFn != nil {
			return m.UpdateFn(ctx, course)
		}
		return nil
}
func (m *MockCourseRepository) Delete(ctx context.Context, id uint64) error {
		if m.DeleteFn != nil {
			return m.DeleteFn(ctx, id)
		}
		return nil
}
func (m *MockCourseRepository) IncrViewCount(ctx context.Context, id uint64) error {
		if m.IncrViewCountFn != nil {
			return m.IncrViewCountFn(ctx, id)
		}
		return nil
}

// MockCourseUnitRepository is a test double for repository.CourseUnitRepository.
type MockCourseUnitRepository struct {
	GetByIDFn        func(ctx context.Context, id uint64) (*entity.CourseUnit, error)
	ListByCourseIDFn func(ctx context.Context, courseID uint64) ([]*entity.CourseUnit, error)
	CreateFn         func(ctx context.Context, unit *entity.CourseUnit) error
	UpdateFn         func(ctx context.Context, unit *entity.CourseUnit) error
	DeleteFn         func(ctx context.Context, id uint64) error
}

func (m *MockCourseUnitRepository) GetByID(ctx context.Context, id uint64) (*entity.CourseUnit, error) {
		if m.GetByIDFn != nil {
			return m.GetByIDFn(ctx, id)
		}
		return nil, nil
}
func (m *MockCourseUnitRepository) ListByCourseID(ctx context.Context, courseID uint64) ([]*entity.CourseUnit, error) {
		if m.ListByCourseIDFn != nil {
			return m.ListByCourseIDFn(ctx, courseID)
		}
		return nil, nil
}
func (m *MockCourseUnitRepository) Create(ctx context.Context, unit *entity.CourseUnit) error {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, unit)
		}
		return nil
}
func (m *MockCourseUnitRepository) Update(ctx context.Context, unit *entity.CourseUnit) error {
		if m.UpdateFn != nil {
			return m.UpdateFn(ctx, unit)
		}
		return nil
}
func (m *MockCourseUnitRepository) Delete(ctx context.Context, id uint64) error {
		if m.DeleteFn != nil {
			return m.DeleteFn(ctx, id)
		}
		return nil
}

// MockContentPermissionRepository is a test double for repository.ContentPermissionRepository.
type MockContentPermissionRepository struct {
	GetByContentFn        func(ctx context.Context, contentType int8, contentID uint64) ([]*entity.ContentPermission, error)
	SetContentPermsFn     func(ctx context.Context, contentType int8, contentID uint64, roleIDs []uint) error
}

func (m *MockContentPermissionRepository) GetByContent(ctx context.Context, contentType int8, contentID uint64) ([]*entity.ContentPermission, error) {
		if m.GetByContentFn != nil {
			return m.GetByContentFn(ctx, contentType, contentID)
		}
		return nil, nil
}
func (m *MockContentPermissionRepository) SetContentPermissions(ctx context.Context, contentType int8, contentID uint64, roleIDs []uint) error {
	return m.SetContentPermsFn(ctx, contentType, contentID, roleIDs)
}

// MockStudyRecordRepository is a test double for repository.StudyRecordRepository.
type MockStudyRecordRepository struct {
	GetByUserAndUnitFn func(ctx context.Context, userID, unitID uint64) (*entity.UserStudyRecord, error)
	ListByUserFn       func(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.UserStudyRecord, int64, error)
	UpsertFn           func(ctx context.Context, record *entity.UserStudyRecord) error
}

func (m *MockStudyRecordRepository) GetByUserAndUnit(ctx context.Context, userID, unitID uint64) (*entity.UserStudyRecord, error) {
		if m.GetByUserAndUnitFn != nil {
			return m.GetByUserAndUnitFn(ctx, userID, unitID)
		}
		return nil, nil
}
func (m *MockStudyRecordRepository) ListByUser(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.UserStudyRecord, int64, error) {
		if m.ListByUserFn != nil {
			return m.ListByUserFn(ctx, userID, page, pageSize)
		}
		return nil, 0, nil
}
func (m *MockStudyRecordRepository) Upsert(ctx context.Context, record *entity.UserStudyRecord) error {
		if m.UpsertFn != nil {
			return m.UpsertFn(ctx, record)
		}
		return nil
}

// MockCollectionRepository is a test double for repository.CollectionRepository.
type MockCollectionRepository struct {
	GetFn    func(ctx context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Collection, error)
	ListFn   func(ctx context.Context, userID uint64, page, pageSize int, contentType *int8) ([]*entity.Collection, int64, error)
	CreateFn func(ctx context.Context, collection *entity.Collection) error
	DeleteFn func(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
}

func (m *MockCollectionRepository) Get(ctx context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Collection, error) {
		if m.GetFn != nil {
			return m.GetFn(ctx, userID, contentType, contentID)
		}
		return nil, nil
}
func (m *MockCollectionRepository) List(ctx context.Context, userID uint64, page, pageSize int, contentType *int8) ([]*entity.Collection, int64, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, userID, page, pageSize, contentType)
		}
		return nil, 0, nil
}
func (m *MockCollectionRepository) Create(ctx context.Context, collection *entity.Collection) error {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, collection)
		}
		return nil
}
func (m *MockCollectionRepository) Delete(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
		if m.DeleteFn != nil {
			return m.DeleteFn(ctx, userID, contentType, contentID)
		}
		return nil
}

// MockLikeRepository is a test double for repository.LikeRepository.
type MockLikeRepository struct {
	GetFn    func(ctx context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Like, error)
	CreateFn func(ctx context.Context, like *entity.Like) error
	DeleteFn func(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
}

func (m *MockLikeRepository) Get(ctx context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Like, error) {
		if m.GetFn != nil {
			return m.GetFn(ctx, userID, contentType, contentID)
		}
		return nil, nil
}
func (m *MockLikeRepository) Create(ctx context.Context, like *entity.Like) error {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, like)
		}
		return nil
}
func (m *MockLikeRepository) Delete(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
		if m.DeleteFn != nil {
			return m.DeleteFn(ctx, userID, contentType, contentID)
		}
		return nil
}

// MockCommentRepository is a test double for repository.CommentRepository.
type MockCommentRepository struct {
	GetByIDFn      func(ctx context.Context, id uint64) (*entity.Comment, error)
	ListFn         func(ctx context.Context, contentType int8, contentID uint64, page, pageSize int) ([]*entity.Comment, int64, error)
	ListAdminFn    func(ctx context.Context, page, pageSize int, status *int8) ([]*entity.Comment, int64, error)
	CreateFn       func(ctx context.Context, comment *entity.Comment) error
	UpdateStatusFn func(ctx context.Context, id uint64, status int8) error
	DeleteFn       func(ctx context.Context, id uint64) error
}

func (m *MockCommentRepository) GetByID(ctx context.Context, id uint64) (*entity.Comment, error) {
		if m.GetByIDFn != nil {
			return m.GetByIDFn(ctx, id)
		}
		return nil, nil
}
func (m *MockCommentRepository) List(ctx context.Context, contentType int8, contentID uint64, page, pageSize int) ([]*entity.Comment, int64, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, contentType, contentID, page, pageSize)
		}
		return nil, 0, nil
}
func (m *MockCommentRepository) ListAdmin(ctx context.Context, page, pageSize int, status *int8) ([]*entity.Comment, int64, error) {
		if m.ListAdminFn != nil {
			return m.ListAdminFn(ctx, page, pageSize, status)
		}
		return nil, 0, nil
}
func (m *MockCommentRepository) Create(ctx context.Context, comment *entity.Comment) error {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, comment)
		}
		return nil
}
func (m *MockCommentRepository) UpdateStatus(ctx context.Context, id uint64, status int8) error {
		if m.UpdateStatusFn != nil {
			return m.UpdateStatusFn(ctx, id, status)
		}
		return nil
}
func (m *MockCommentRepository) Delete(ctx context.Context, id uint64) error {
		if m.DeleteFn != nil {
			return m.DeleteFn(ctx, id)
		}
		return nil
}

// MockNotificationRepository is a test double for repository.NotificationRepository.
type MockNotificationRepository struct {
	GetByIDFn    func(ctx context.Context, id uint64) (*entity.Notification, error)
	ListFn       func(ctx context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, error)
	UnreadCountFn func(ctx context.Context, userID uint64) (int64, error)
	MarkReadFn   func(ctx context.Context, id uint64) error
	MarkAllReadFn func(ctx context.Context, userID uint64) error
}

func (m *MockNotificationRepository) GetByID(ctx context.Context, id uint64) (*entity.Notification, error) {
		if m.GetByIDFn != nil {
			return m.GetByIDFn(ctx, id)
		}
		return nil, nil
}
func (m *MockNotificationRepository) List(ctx context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, userID, page, pageSize, isRead)
		}
		return nil, 0, nil
}
func (m *MockNotificationRepository) UnreadCount(ctx context.Context, userID uint64) (int64, error) {
		if m.UnreadCountFn != nil {
			return m.UnreadCountFn(ctx, userID)
		}
		return 0, nil
}
func (m *MockNotificationRepository) MarkRead(ctx context.Context, id uint64) error {
		if m.MarkReadFn != nil {
			return m.MarkReadFn(ctx, id)
		}
		return nil
}
func (m *MockNotificationRepository) MarkAllRead(ctx context.Context, userID uint64) error {
		if m.MarkAllReadFn != nil {
			return m.MarkAllReadFn(ctx, userID)
		}
		return nil
}

// MockWechatConfigRepository is a test double for repository.WechatConfigRepository.
type MockWechatConfigRepository struct {
	GetFn    func(ctx context.Context) (*entity.WechatConfig, error)
	UpdateFn func(ctx context.Context, cfg *entity.WechatConfig) error
}

func (m *MockWechatConfigRepository) Get(ctx context.Context) (*entity.WechatConfig, error) {
		if m.GetFn != nil {
			return m.GetFn(ctx)
		}
		return nil, nil
}
func (m *MockWechatConfigRepository) Update(ctx context.Context, cfg *entity.WechatConfig) error {
		if m.UpdateFn != nil {
			return m.UpdateFn(ctx, cfg)
		}
		return nil
}

// MockAuditLogRepository is a test double for repository.AuditLogRepository.
type MockAuditLogRepository struct {
	GetByIDFn func(ctx context.Context, id uint64) (*entity.AuditLog, error)
	ListFn   func(ctx context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error)
	CreateFn func(ctx context.Context, log *entity.AuditLog) error
}

func (m *MockAuditLogRepository) GetByID(ctx context.Context, id uint64) (*entity.AuditLog, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockAuditLogRepository) List(ctx context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, page, pageSize, module, action, startTime, endTime)
		}
		return nil, 0, nil
}
func (m *MockAuditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, log)
		}
		return nil
}

// MockLogConfigRepository is a test double for repository.LogConfigRepository.
type MockLogConfigRepository struct {
	GetFn    func(ctx context.Context) (*entity.LogConfig, error)
	UpdateFn func(ctx context.Context, cfg *entity.LogConfig) error
}

func (m *MockLogConfigRepository) Get(ctx context.Context) (*entity.LogConfig, error) {
		if m.GetFn != nil {
			return m.GetFn(ctx)
		}
		return nil, nil
}
func (m *MockLogConfigRepository) Update(ctx context.Context, cfg *entity.LogConfig) error {
		if m.UpdateFn != nil {
			return m.UpdateFn(ctx, cfg)
		}
		return nil
}

// MockWechatClient is a test double for wechat.Client.
type MockWechatClient struct {
	Code2SessionFn func(ctx context.Context, code string) (string, error)
}

func (m *MockWechatClient) Code2Session(ctx context.Context, code string) (string, error) {
	return m.Code2SessionFn(ctx, code)
}

// ==================== Service Mocks ====================

// MockAuthService is a test double for service.AuthService.
type MockAuthService struct {
	WechatLoginFn  func(ctx context.Context, req *dto.WechatLoginRequest) (*dto.LoginResponseData, error)
	AdminLoginFn   func(ctx context.Context, req *dto.AdminLoginRequest) (*dto.LoginResponseData, error)
	RefreshTokenFn func(ctx context.Context, userID uint64, userType int8) (*dto.LoginResponseData, error)
}

func (m *MockAuthService) WechatLogin(ctx context.Context, req *dto.WechatLoginRequest) (*dto.LoginResponseData, error) {
	return m.WechatLoginFn(ctx, req)
}
func (m *MockAuthService) AdminLogin(ctx context.Context, req *dto.AdminLoginRequest) (*dto.LoginResponseData, error) {
	return m.AdminLoginFn(ctx, req)
}
func (m *MockAuthService) RefreshToken(ctx context.Context, userID uint64, userType int8) (*dto.LoginResponseData, error) {
	return m.RefreshTokenFn(ctx, userID, userType)
}

// MockUserService is a test double for service.UserService.
type MockUserService struct {
	GetProfileFn      func(ctx context.Context, userID uint64) (*entity.User, error)
	UpdateProfileFn   func(ctx context.Context, userID uint64, req *dto.UserProfileUpdateRequest) error
	GetPermissionsFn  func(ctx context.Context, userID uint64) ([]string, []string, error)
	ListFn            func(ctx context.Context, page, pageSize int, keyword string, userType, status *int8) ([]*entity.User, int64, error)
	GetByIDFn         func(ctx context.Context, id uint64) (*entity.User, error)
	CreateAdminUserFn func(ctx context.Context, req *dto.CreateAdminUserRequest) (uint64, error)
	UpdateUserFn      func(ctx context.Context, id uint64, req *dto.UpdateUserRequest, operatorID uint64) error
	DeleteUserFn      func(ctx context.Context, id uint64) error
	AssignRolesFn     func(ctx context.Context, userID uint64, req *dto.AssignRolesRequest) error
	AddTagFn          func(ctx context.Context, userID uint64, req *dto.AddTagRequest) (uint, error)
	DeleteTagFn       func(ctx context.Context, userID, tagID uint64) error
}

func (m *MockUserService) GetProfile(ctx context.Context, userID uint64) (*entity.User, error) {
		if m.GetProfileFn != nil {
			return m.GetProfileFn(ctx, userID)
		}
		return nil, nil
}
func (m *MockUserService) UpdateProfile(ctx context.Context, userID uint64, req *dto.UserProfileUpdateRequest) error {
		if m.UpdateProfileFn != nil {
			return m.UpdateProfileFn(ctx, userID, req)
		}
		return nil
}
func (m *MockUserService) GetPermissions(ctx context.Context, userID uint64) ([]string, []string, error) {
	return m.GetPermissionsFn(ctx, userID)
}
func (m *MockUserService) List(ctx context.Context, page, pageSize int, keyword string, userType, status *int8) ([]*entity.User, int64, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, page, pageSize, keyword, userType, status)
		}
		return nil, 0, nil
}
func (m *MockUserService) GetByID(ctx context.Context, id uint64) (*entity.User, error) {
		if m.GetByIDFn != nil {
			return m.GetByIDFn(ctx, id)
		}
		return nil, nil
}
func (m *MockUserService) CreateAdminUser(ctx context.Context, req *dto.CreateAdminUserRequest) (uint64, error) {
		if m.CreateAdminUserFn != nil {
			return m.CreateAdminUserFn(ctx, req)
		}
		return 0, nil
}
func (m *MockUserService) UpdateUser(ctx context.Context, id uint64, req *dto.UpdateUserRequest, operatorID uint64) error {
		if m.UpdateUserFn != nil {
			return m.UpdateUserFn(ctx, id, req, operatorID)
		}
		return nil
}
func (m *MockUserService) DeleteUser(ctx context.Context, id uint64) error {
		if m.DeleteUserFn != nil {
			return m.DeleteUserFn(ctx, id)
		}
		return nil
}
func (m *MockUserService) AssignRoles(ctx context.Context, userID uint64, req *dto.AssignRolesRequest) error {
		if m.AssignRolesFn != nil {
			return m.AssignRolesFn(ctx, userID, req)
		}
		return nil
}
func (m *MockUserService) AddTag(ctx context.Context, userID uint64, req *dto.AddTagRequest) (uint, error) {
		if m.AddTagFn != nil {
			return m.AddTagFn(ctx, userID, req)
		}
		return 0, nil
}
func (m *MockUserService) DeleteTag(ctx context.Context, userID, tagID uint64) error {
		if m.DeleteTagFn != nil {
			return m.DeleteTagFn(ctx, userID, tagID)
		}
		return nil
}

// MockRoleService is a test double for service.RoleService.
type MockRoleService struct {
	ListFn     func(ctx context.Context) ([]*entity.Role, error)
	GetByIDFn  func(ctx context.Context, id uint) (*entity.Role, error)
	CreateFn   func(ctx context.Context, req *dto.CreateRoleRequest) (uint, error)
	UpdateFn   func(ctx context.Context, id uint, req *dto.UpdateRoleRequest) error
	DeleteFn   func(ctx context.Context, id uint) error
}

func (m *MockRoleService) List(ctx context.Context) ([]*entity.Role, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx)
		}
		return nil, nil
}
func (m *MockRoleService) GetByID(ctx context.Context, id uint) (*entity.Role, error) {
		if m.GetByIDFn != nil {
			return m.GetByIDFn(ctx, id)
		}
		return nil, nil
}
func (m *MockRoleService) Create(ctx context.Context, req *dto.CreateRoleRequest) (uint, error) {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, req)
		}
		return 0, nil
}
func (m *MockRoleService) Update(ctx context.Context, id uint, req *dto.UpdateRoleRequest) error {
		if m.UpdateFn != nil {
			return m.UpdateFn(ctx, id, req)
		}
		return nil
}
func (m *MockRoleService) Delete(ctx context.Context, id uint) error {
		if m.DeleteFn != nil {
			return m.DeleteFn(ctx, id)
		}
		return nil
}

// MockPermissionService is a test double for service.PermissionService.
type MockPermissionService struct {
	GetTreeFn func(ctx context.Context) ([]*entity.Permission, error)
}

func (m *MockPermissionService) GetTree(ctx context.Context) ([]*entity.Permission, error) {
		if m.GetTreeFn != nil {
			return m.GetTreeFn(ctx)
		}
		return nil, nil
}

// MockModuleService is a test double for service.ModuleService.
type MockModuleService struct {
	ListFn       func(ctx context.Context, status *int8) ([]*entity.Module, error)
	CreateFn     func(ctx context.Context, req *dto.CreateModuleRequest) (uint, error)
	UpdateFn     func(ctx context.Context, id uint, req *dto.CreateModuleRequest) error
	DeleteFn     func(ctx context.Context, id uint) error
	GetPagesFn   func(ctx context.Context, moduleID uint) ([]*entity.ModulePage, error)
	CreatePageFn func(ctx context.Context, moduleID uint, req *dto.CreateModulePageRequest) (uint, error)
	UpdatePageFn func(ctx context.Context, moduleID, pageID uint, req *dto.CreateModulePageRequest) error
	DeletePageFn func(ctx context.Context, moduleID, pageID uint) error
}

func (m *MockModuleService) List(ctx context.Context, status *int8) ([]*entity.Module, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, status)
		}
		return nil, nil
}
func (m *MockModuleService) Create(ctx context.Context, req *dto.CreateModuleRequest) (uint, error) {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, req)
		}
		return 0, nil
}
func (m *MockModuleService) Update(ctx context.Context, id uint, req *dto.CreateModuleRequest) error {
		if m.UpdateFn != nil {
			return m.UpdateFn(ctx, id, req)
		}
		return nil
}
func (m *MockModuleService) Delete(ctx context.Context, id uint) error {
		if m.DeleteFn != nil {
			return m.DeleteFn(ctx, id)
		}
		return nil
}
func (m *MockModuleService) GetPages(ctx context.Context, moduleID uint) ([]*entity.ModulePage, error) {
		if m.GetPagesFn != nil {
			return m.GetPagesFn(ctx, moduleID)
		}
		return nil, nil
}
func (m *MockModuleService) CreatePage(ctx context.Context, moduleID uint, req *dto.CreateModulePageRequest) (uint, error) {
		if m.CreatePageFn != nil {
			return m.CreatePageFn(ctx, moduleID, req)
		}
		return 0, nil
}
func (m *MockModuleService) UpdatePage(ctx context.Context, moduleID, pageID uint, req *dto.CreateModulePageRequest) error {
		if m.UpdatePageFn != nil {
			return m.UpdatePageFn(ctx, moduleID, pageID, req)
		}
		return nil
}
func (m *MockModuleService) DeletePage(ctx context.Context, moduleID, pageID uint) error {
		if m.DeletePageFn != nil {
			return m.DeletePageFn(ctx, moduleID, pageID)
		}
		return nil
}

// MockArticleService is a test double for service.ArticleService.
type MockArticleService struct {
	ListFn        func(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, sort string, userID *uint64) ([]*entity.Article, int64, error)
	GetByIDFn     func(ctx context.Context, id uint64, userID *uint64) (*entity.Article, error)
	AdminListFn   func(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8) ([]*entity.Article, int64, error)
	AdminGetByIDFn func(ctx context.Context, id uint64) (*entity.Article, error)
	CreateFn      func(ctx context.Context, req *dto.CreateArticleRequest, authorID uint64) (uint64, error)
	UpdateFn      func(ctx context.Context, id uint64, req *dto.UpdateArticleRequest) error
	DeleteFn      func(ctx context.Context, id uint64) error
	PublishFn     func(ctx context.Context, id uint64, req *dto.PublishArticleRequest) error
}

func (m *MockArticleService) List(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, sort string, userID *uint64) ([]*entity.Article, int64, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, page, pageSize, keyword, moduleID, sort, userID)
		}
		return nil, 0, nil
}
func (m *MockArticleService) GetByID(ctx context.Context, id uint64, userID *uint64) (*entity.Article, error) {
		if m.GetByIDFn != nil {
			return m.GetByIDFn(ctx, id, userID)
		}
		return nil, nil
}
func (m *MockArticleService) AdminList(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8) ([]*entity.Article, int64, error) {
		if m.AdminListFn != nil {
			return m.AdminListFn(ctx, page, pageSize, keyword, moduleID, status)
		}
		return nil, 0, nil
}
func (m *MockArticleService) AdminGetByID(ctx context.Context, id uint64) (*entity.Article, error) {
		if m.AdminGetByIDFn != nil {
			return m.AdminGetByIDFn(ctx, id)
		}
		return nil, nil
}
func (m *MockArticleService) Create(ctx context.Context, req *dto.CreateArticleRequest, authorID uint64) (uint64, error) {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, req, authorID)
		}
		return 0, nil
}
func (m *MockArticleService) Update(ctx context.Context, id uint64, req *dto.UpdateArticleRequest) error {
		if m.UpdateFn != nil {
			return m.UpdateFn(ctx, id, req)
		}
		return nil
}
func (m *MockArticleService) Delete(ctx context.Context, id uint64) error {
		if m.DeleteFn != nil {
			return m.DeleteFn(ctx, id)
		}
		return nil
}
func (m *MockArticleService) Publish(ctx context.Context, id uint64, req *dto.PublishArticleRequest) error {
		if m.PublishFn != nil {
			return m.PublishFn(ctx, id, req)
		}
		return nil
}

// MockCourseService is a test double for service.CourseService.
type MockCourseService struct {
	ListFn         func(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, isFree *bool, userID *uint64) ([]*entity.Course, int64, error)
	GetByIDFn      func(ctx context.Context, id uint64, userID *uint64) (*entity.Course, error)
	AdminListFn    func(ctx context.Context, page, pageSize int, keyword string, status *int8) ([]*entity.Course, int64, error)
	AdminGetByIDFn func(ctx context.Context, id uint64) (*entity.Course, error)
	CreateFn       func(ctx context.Context, req *dto.CreateCourseRequest, authorID uint64) (uint64, error)
	UpdateFn       func(ctx context.Context, id uint64, req *dto.UpdateCourseRequest) error
	DeleteFn       func(ctx context.Context, id uint64) error
	PublishFn      func(ctx context.Context, id uint64, req *dto.PublishCourseRequest) error
	GetUnitsFn     func(ctx context.Context, courseID uint64) ([]*entity.CourseUnit, error)
	CreateUnitFn   func(ctx context.Context, courseID uint64, req *dto.CreateCourseUnitRequest) (uint64, error)
	UpdateUnitFn   func(ctx context.Context, courseID, unitID uint64, req *dto.CreateCourseUnitRequest) error
	DeleteUnitFn   func(ctx context.Context, courseID, unitID uint64) error
}

func (m *MockCourseService) List(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, isFree *bool, userID *uint64) ([]*entity.Course, int64, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, page, pageSize, keyword, moduleID, isFree, userID)
		}
		return nil, 0, nil
}
func (m *MockCourseService) GetByID(ctx context.Context, id uint64, userID *uint64) (*entity.Course, error) {
		if m.GetByIDFn != nil {
			return m.GetByIDFn(ctx, id, userID)
		}
		return nil, nil
}
func (m *MockCourseService) AdminList(ctx context.Context, page, pageSize int, keyword string, status *int8) ([]*entity.Course, int64, error) {
		if m.AdminListFn != nil {
			return m.AdminListFn(ctx, page, pageSize, keyword, status)
		}
		return nil, 0, nil
}
func (m *MockCourseService) AdminGetByID(ctx context.Context, id uint64) (*entity.Course, error) {
		if m.AdminGetByIDFn != nil {
			return m.AdminGetByIDFn(ctx, id)
		}
		return nil, nil
}
func (m *MockCourseService) Create(ctx context.Context, req *dto.CreateCourseRequest, authorID uint64) (uint64, error) {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, req, authorID)
		}
		return 0, nil
}
func (m *MockCourseService) Update(ctx context.Context, id uint64, req *dto.UpdateCourseRequest) error {
		if m.UpdateFn != nil {
			return m.UpdateFn(ctx, id, req)
		}
		return nil
}
func (m *MockCourseService) Delete(ctx context.Context, id uint64) error {
		if m.DeleteFn != nil {
			return m.DeleteFn(ctx, id)
		}
		return nil
}
func (m *MockCourseService) Publish(ctx context.Context, id uint64, req *dto.PublishCourseRequest) error {
		if m.PublishFn != nil {
			return m.PublishFn(ctx, id, req)
		}
		return nil
}
func (m *MockCourseService) GetUnits(ctx context.Context, courseID uint64) ([]*entity.CourseUnit, error) {
		if m.GetUnitsFn != nil {
			return m.GetUnitsFn(ctx, courseID)
		}
		return nil, nil
}
func (m *MockCourseService) CreateUnit(ctx context.Context, courseID uint64, req *dto.CreateCourseUnitRequest) (uint64, error) {
		if m.CreateUnitFn != nil {
			return m.CreateUnitFn(ctx, courseID, req)
		}
		return 0, nil
}
func (m *MockCourseService) UpdateUnit(ctx context.Context, courseID, unitID uint64, req *dto.CreateCourseUnitRequest) error {
		if m.UpdateUnitFn != nil {
			return m.UpdateUnitFn(ctx, courseID, unitID, req)
		}
		return nil
}
func (m *MockCourseService) DeleteUnit(ctx context.Context, courseID, unitID uint64) error {
		if m.DeleteUnitFn != nil {
			return m.DeleteUnitFn(ctx, courseID, unitID)
		}
		return nil
}

// MockStudyRecordService is a test double for service.StudyRecordService.
type MockStudyRecordService struct {
	ListFn   func(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.UserStudyRecord, int64, error)
	UpdateFn func(ctx context.Context, userID uint64, req *dto.UpdateStudyRecordRequest) error
}

func (m *MockStudyRecordService) List(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.UserStudyRecord, int64, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, userID, page, pageSize)
		}
		return nil, 0, nil
}
func (m *MockStudyRecordService) Update(ctx context.Context, userID uint64, req *dto.UpdateStudyRecordRequest) error {
		if m.UpdateFn != nil {
			return m.UpdateFn(ctx, userID, req)
		}
		return nil
}

// MockCollectionService is a test double for service.CollectionService.
type MockCollectionService struct {
	ListFn   func(ctx context.Context, userID uint64, page, pageSize int, contentType *int8) ([]*entity.Collection, int64, error)
	AddFn    func(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
	RemoveFn func(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
}

func (m *MockCollectionService) List(ctx context.Context, userID uint64, page, pageSize int, contentType *int8) ([]*entity.Collection, int64, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, userID, page, pageSize, contentType)
		}
		return nil, 0, nil
}
func (m *MockCollectionService) Add(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
		if m.AddFn != nil {
			return m.AddFn(ctx, userID, contentType, contentID)
		}
		return nil
}
func (m *MockCollectionService) Remove(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
		if m.RemoveFn != nil {
			return m.RemoveFn(ctx, userID, contentType, contentID)
		}
		return nil
}

// MockLikeService is a test double for service.LikeService.
type MockLikeService struct {
	AddFn    func(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
	RemoveFn func(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
}

func (m *MockLikeService) Add(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
		if m.AddFn != nil {
			return m.AddFn(ctx, userID, contentType, contentID)
		}
		return nil
}
func (m *MockLikeService) Remove(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
		if m.RemoveFn != nil {
			return m.RemoveFn(ctx, userID, contentType, contentID)
		}
		return nil
}

// MockCommentService is a test double for service.CommentService.
type MockCommentService struct {
	ListFn      func(ctx context.Context, contentType int8, contentID uint64, page, pageSize int) ([]*entity.Comment, int64, error)
	CreateFn    func(ctx context.Context, userID uint64, contentType int8, contentID uint64, req *dto.CreateCommentRequest) (*entity.Comment, error)
	AdminListFn func(ctx context.Context, page, pageSize int, status *int8) ([]*entity.Comment, int64, error)
	AuditFn     func(ctx context.Context, id uint64, req *dto.AuditCommentRequest) error
	DeleteFn    func(ctx context.Context, id uint64) error
}

func (m *MockCommentService) List(ctx context.Context, contentType int8, contentID uint64, page, pageSize int) ([]*entity.Comment, int64, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, contentType, contentID, page, pageSize)
		}
		return nil, 0, nil
}
func (m *MockCommentService) Create(ctx context.Context, userID uint64, contentType int8, contentID uint64, req *dto.CreateCommentRequest) (*entity.Comment, error) {
		if m.CreateFn != nil {
			return m.CreateFn(ctx, userID, contentType, contentID, req)
		}
		return nil, nil
}
func (m *MockCommentService) AdminList(ctx context.Context, page, pageSize int, status *int8) ([]*entity.Comment, int64, error) {
		if m.AdminListFn != nil {
			return m.AdminListFn(ctx, page, pageSize, status)
		}
		return nil, 0, nil
}
func (m *MockCommentService) Audit(ctx context.Context, id uint64, req *dto.AuditCommentRequest) error {
		if m.AuditFn != nil {
			return m.AuditFn(ctx, id, req)
		}
		return nil
}
func (m *MockCommentService) Delete(ctx context.Context, id uint64) error {
		if m.DeleteFn != nil {
			return m.DeleteFn(ctx, id)
		}
		return nil
}

// MockNotificationService is a test double for service.NotificationService.
type MockNotificationService struct {
	ListFn       func(ctx context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, int64, error)
	MarkReadFn   func(ctx context.Context, id uint64) error
	MarkAllReadFn func(ctx context.Context, userID uint64) error
}

func (m *MockNotificationService) List(ctx context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, int64, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, userID, page, pageSize, isRead)
		}
		return nil, 0, 0, nil
}
func (m *MockNotificationService) MarkRead(ctx context.Context, id uint64) error {
		if m.MarkReadFn != nil {
			return m.MarkReadFn(ctx, id)
		}
		return nil
}
func (m *MockNotificationService) MarkAllRead(ctx context.Context, userID uint64) error {
		if m.MarkAllReadFn != nil {
			return m.MarkAllReadFn(ctx, userID)
		}
		return nil
}

// MockWechatConfigService is a test double for service.WechatConfigService.
type MockWechatConfigService struct {
	GetFn    func(ctx context.Context) (*entity.WechatConfig, error)
	UpdateFn func(ctx context.Context, req *dto.UpdateWechatConfigRequest) error
}

func (m *MockWechatConfigService) Get(ctx context.Context) (*entity.WechatConfig, error) {
		if m.GetFn != nil {
			return m.GetFn(ctx)
		}
		return nil, nil
}
func (m *MockWechatConfigService) Update(ctx context.Context, req *dto.UpdateWechatConfigRequest) error {
		if m.UpdateFn != nil {
			return m.UpdateFn(ctx, req)
		}
		return nil
}

// MockAuditLogService is a test double for service.AuditLogService.
type MockAuditLogService struct {
	ListFn func(ctx context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error)
	LogFn  func(ctx context.Context, log *entity.AuditLog)
}

func (m *MockAuditLogService) List(ctx context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error) {
		if m.ListFn != nil {
			return m.ListFn(ctx, page, pageSize, module, action, startTime, endTime)
		}
		return nil, 0, nil
}
func (m *MockAuditLogService) Log(ctx context.Context, log *entity.AuditLog) {
		if m.LogFn != nil {
			m.LogFn(ctx, log)
		}
		return
}

// MockLogConfigService is a test double for service.LogConfigService.
type MockLogConfigService struct {
	GetFn    func(ctx context.Context) (*entity.LogConfig, error)
	UpdateFn func(ctx context.Context, req *dto.UpdateLogConfigRequest) error
}

func (m *MockLogConfigService) Get(ctx context.Context) (*entity.LogConfig, error) {
		if m.GetFn != nil {
			return m.GetFn(ctx)
		}
		return nil, nil
}
func (m *MockLogConfigService) Update(ctx context.Context, req *dto.UpdateLogConfigRequest) error {
		if m.UpdateFn != nil {
			return m.UpdateFn(ctx, req)
		}
		return nil
}

// ==================== Attribute Mocks ====================

// MockAttributeRepository is a test double for repository.AttributeRepository.
type MockAttributeRepository struct {
	GetByIDFn func(ctx context.Context, id uint) (*entity.Attribute, error)
	ListFn    func(ctx context.Context) ([]*entity.Attribute, error)
	CreateFn  func(ctx context.Context, attr *entity.Attribute) error
	UpdateFn  func(ctx context.Context, attr *entity.Attribute) error
	DeleteFn  func(ctx context.Context, id uint) error
}

func (m *MockAttributeRepository) GetByID(ctx context.Context, id uint) (*entity.Attribute, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *MockAttributeRepository) List(ctx context.Context) ([]*entity.Attribute, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx)
	}
	return nil, nil
}
func (m *MockAttributeRepository) Create(ctx context.Context, attr *entity.Attribute) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, attr)
	}
	return nil
}
func (m *MockAttributeRepository) Update(ctx context.Context, attr *entity.Attribute) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, attr)
	}
	return nil
}
func (m *MockAttributeRepository) Delete(ctx context.Context, id uint) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

// MockUserAttributeRepository is a test double for repository.UserAttributeRepository.
type MockUserAttributeRepository struct {
	ListByUserIDFn func(ctx context.Context, userID uint64) ([]*entity.UserAttribute, error)
	UpsertFn       func(ctx context.Context, ua *entity.UserAttribute) error
	DeleteFn       func(ctx context.Context, userID uint64, attributeID uint) error
}

func (m *MockUserAttributeRepository) ListByUserID(ctx context.Context, userID uint64) ([]*entity.UserAttribute, error) {
	if m.ListByUserIDFn != nil {
		return m.ListByUserIDFn(ctx, userID)
	}
	return nil, nil
}
func (m *MockUserAttributeRepository) Upsert(ctx context.Context, ua *entity.UserAttribute) error {
	if m.UpsertFn != nil {
		return m.UpsertFn(ctx, ua)
	}
	return nil
}
func (m *MockUserAttributeRepository) Delete(ctx context.Context, userID uint64, attributeID uint) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, userID, attributeID)
	}
	return nil
}

// MockAttributeService is a test double for service.AttributeService.
type MockAttributeService struct {
	ListFn              func(ctx context.Context) ([]*entity.Attribute, error)
	CreateFn            func(ctx context.Context, req *dto.CreateAttributeRequest) (uint, error)
	UpdateFn            func(ctx context.Context, id uint, req *dto.UpdateAttributeRequest) error
	DeleteFn            func(ctx context.Context, id uint) error
	ListUserAttrsFn     func(ctx context.Context, userID uint64) ([]*entity.UserAttribute, error)
	SetUserAttrFn       func(ctx context.Context, userID uint64, req *dto.SetUserAttributeRequest) error
	DeleteUserAttrFn    func(ctx context.Context, userID uint64, attributeID uint) error
}

func (m *MockAttributeService) List(ctx context.Context) ([]*entity.Attribute, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx)
	}
	return nil, nil
}
func (m *MockAttributeService) Create(ctx context.Context, req *dto.CreateAttributeRequest) (uint, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, req)
	}
	return 0, nil
}
func (m *MockAttributeService) Update(ctx context.Context, id uint, req *dto.UpdateAttributeRequest) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, id, req)
	}
	return nil
}
func (m *MockAttributeService) Delete(ctx context.Context, id uint) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}
func (m *MockAttributeService) ListUserAttributes(ctx context.Context, userID uint64) ([]*entity.UserAttribute, error) {
	if m.ListUserAttrsFn != nil {
		return m.ListUserAttrsFn(ctx, userID)
	}
	return nil, nil
}
func (m *MockAttributeService) SetUserAttribute(ctx context.Context, userID uint64, req *dto.SetUserAttributeRequest) error {
	if m.SetUserAttrFn != nil {
		return m.SetUserAttrFn(ctx, userID, req)
	}
	return nil
}
func (m *MockAttributeService) DeleteUserAttribute(ctx context.Context, userID uint64, attributeID uint) error {
	if m.DeleteUserAttrFn != nil {
		return m.DeleteUserAttrFn(ctx, userID, attributeID)
	}
	return nil
}
