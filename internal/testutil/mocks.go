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
	return m.GetByIDFn(ctx, id)
}
func (m *MockUserRepository) GetByOpenID(ctx context.Context, openID string) (*entity.User, error) {
	return m.GetByOpenIDFn(ctx, openID)
}
func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	return m.CreateFn(ctx, user)
}
func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	return m.UpdateFn(ctx, user)
}
func (m *MockUserRepository) Delete(ctx context.Context, id uint64) error {
	return m.DeleteFn(ctx, id)
}
func (m *MockUserRepository) List(ctx context.Context, page, pageSize int, keyword string, userType, status *int8) ([]*entity.User, int64, error) {
	return m.ListFn(ctx, page, pageSize, keyword, userType, status)
}
func (m *MockUserRepository) GetWithTags(ctx context.Context, id uint64) (*entity.User, error) {
	return m.GetWithTagsFn(ctx, id)
}

// MockAdminUserRepository is a test double for repository.AdminUserRepository.
type MockAdminUserRepository struct {
	GetByEmailFn      func(ctx context.Context, email string) (*entity.AdminUser, error)
	GetByUserIDFn     func(ctx context.Context, userID uint64) (*entity.AdminUser, error)
	CreateFn          func(ctx context.Context, admin *entity.AdminUser) error
	UpdateLastLoginFn func(ctx context.Context, id uint64) error
}

func (m *MockAdminUserRepository) GetByEmail(ctx context.Context, email string) (*entity.AdminUser, error) {
	return m.GetByEmailFn(ctx, email)
}
func (m *MockAdminUserRepository) GetByUserID(ctx context.Context, userID uint64) (*entity.AdminUser, error) {
	return m.GetByUserIDFn(ctx, userID)
}
func (m *MockAdminUserRepository) Create(ctx context.Context, admin *entity.AdminUser) error {
	return m.CreateFn(ctx, admin)
}
func (m *MockAdminUserRepository) UpdateLastLogin(ctx context.Context, id uint64) error {
	return m.UpdateLastLoginFn(ctx, id)
}

// MockUserTagRepository is a test double for repository.UserTagRepository.
type MockUserTagRepository struct {
	GetByUserIDFn func(ctx context.Context, userID uint64) ([]*entity.UserTag, error)
	CreateFn      func(ctx context.Context, tag *entity.UserTag) error
	DeleteFn      func(ctx context.Context, id uint) error
}

func (m *MockUserTagRepository) GetByUserID(ctx context.Context, userID uint64) ([]*entity.UserTag, error) {
	return m.GetByUserIDFn(ctx, userID)
}
func (m *MockUserTagRepository) Create(ctx context.Context, tag *entity.UserTag) error {
	return m.CreateFn(ctx, tag)
}
func (m *MockUserTagRepository) Delete(ctx context.Context, id uint) error {
	return m.DeleteFn(ctx, id)
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
	return m.GetByIDFn(ctx, id)
}
func (m *MockRoleRepository) GetWithPermissions(ctx context.Context, id uint) (*entity.Role, error) {
	return m.GetWithPermsFn(ctx, id)
}
func (m *MockRoleRepository) List(ctx context.Context) ([]*entity.Role, error) {
	return m.ListFn(ctx)
}
func (m *MockRoleRepository) Create(ctx context.Context, role *entity.Role) error {
	return m.CreateFn(ctx, role)
}
func (m *MockRoleRepository) Update(ctx context.Context, role *entity.Role) error {
	return m.UpdateFn(ctx, role)
}
func (m *MockRoleRepository) Delete(ctx context.Context, id uint) error {
	return m.DeleteFn(ctx, id)
}
func (m *MockRoleRepository) AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	return m.AssignPermsFn(ctx, roleID, permissionIDs)
}
func (m *MockRoleRepository) GetUserRoles(ctx context.Context, userID uint64) ([]*entity.Role, error) {
	return m.GetUserRolesFn(ctx, userID)
}
func (m *MockRoleRepository) AssignUserRoles(ctx context.Context, userID uint64, roleIDs []uint) error {
	return m.AssignUserRolesFn(ctx, userID, roleIDs)
}
func (m *MockRoleRepository) HasUsers(ctx context.Context, roleID uint) (bool, error) {
	return m.HasUsersFn(ctx, roleID)
}

// MockPermissionRepository is a test double for repository.PermissionRepository.
type MockPermissionRepository struct {
	ListFn               func(ctx context.Context) ([]*entity.Permission, error)
	GetByIDFn            func(ctx context.Context, id uint) (*entity.Permission, error)
	GetUserPermissionsFn func(ctx context.Context, userID uint64) ([]*entity.Permission, error)
}

func (m *MockPermissionRepository) List(ctx context.Context) ([]*entity.Permission, error) {
	return m.ListFn(ctx)
}
func (m *MockPermissionRepository) GetByID(ctx context.Context, id uint) (*entity.Permission, error) {
	return m.GetByIDFn(ctx, id)
}
func (m *MockPermissionRepository) GetUserPermissions(ctx context.Context, userID uint64) ([]*entity.Permission, error) {
	return m.GetUserPermissionsFn(ctx, userID)
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
	return m.GetByIDFn(ctx, id)
}
func (m *MockModuleRepository) List(ctx context.Context, status *int8) ([]*entity.Module, error) {
	return m.ListFn(ctx, status)
}
func (m *MockModuleRepository) Create(ctx context.Context, module *entity.Module) error {
	return m.CreateFn(ctx, module)
}
func (m *MockModuleRepository) Update(ctx context.Context, module *entity.Module) error {
	return m.UpdateFn(ctx, module)
}
func (m *MockModuleRepository) Delete(ctx context.Context, id uint) error {
	return m.DeleteFn(ctx, id)
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
	return m.GetByIDFn(ctx, id)
}
func (m *MockModulePageRepository) ListByModuleID(ctx context.Context, moduleID uint) ([]*entity.ModulePage, error) {
	return m.ListByModuleIDFn(ctx, moduleID)
}
func (m *MockModulePageRepository) Create(ctx context.Context, page *entity.ModulePage) error {
	return m.CreateFn(ctx, page)
}
func (m *MockModulePageRepository) Update(ctx context.Context, page *entity.ModulePage) error {
	return m.UpdateFn(ctx, page)
}
func (m *MockModulePageRepository) Delete(ctx context.Context, id uint) error {
	return m.DeleteFn(ctx, id)
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
	return m.GetByIDFn(ctx, id)
}
func (m *MockArticleRepository) List(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8, sort string) ([]*entity.Article, int64, error) {
	return m.ListFn(ctx, page, pageSize, keyword, moduleID, status, sort)
}
func (m *MockArticleRepository) Create(ctx context.Context, article *entity.Article) error {
	return m.CreateFn(ctx, article)
}
func (m *MockArticleRepository) Update(ctx context.Context, article *entity.Article) error {
	return m.UpdateFn(ctx, article)
}
func (m *MockArticleRepository) Delete(ctx context.Context, id uint64) error {
	return m.DeleteFn(ctx, id)
}
func (m *MockArticleRepository) IncrViewCount(ctx context.Context, id uint64) error {
	return m.IncrViewCountFn(ctx, id)
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
	return m.GetByIDFn(ctx, id)
}
func (m *MockCourseRepository) List(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8, isFree *bool) ([]*entity.Course, int64, error) {
	return m.ListFn(ctx, page, pageSize, keyword, moduleID, status, isFree)
}
func (m *MockCourseRepository) Create(ctx context.Context, course *entity.Course) error {
	return m.CreateFn(ctx, course)
}
func (m *MockCourseRepository) Update(ctx context.Context, course *entity.Course) error {
	return m.UpdateFn(ctx, course)
}
func (m *MockCourseRepository) Delete(ctx context.Context, id uint64) error {
	return m.DeleteFn(ctx, id)
}
func (m *MockCourseRepository) IncrViewCount(ctx context.Context, id uint64) error {
	return m.IncrViewCountFn(ctx, id)
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
	return m.GetByIDFn(ctx, id)
}
func (m *MockCourseUnitRepository) ListByCourseID(ctx context.Context, courseID uint64) ([]*entity.CourseUnit, error) {
	return m.ListByCourseIDFn(ctx, courseID)
}
func (m *MockCourseUnitRepository) Create(ctx context.Context, unit *entity.CourseUnit) error {
	return m.CreateFn(ctx, unit)
}
func (m *MockCourseUnitRepository) Update(ctx context.Context, unit *entity.CourseUnit) error {
	return m.UpdateFn(ctx, unit)
}
func (m *MockCourseUnitRepository) Delete(ctx context.Context, id uint64) error {
	return m.DeleteFn(ctx, id)
}

// MockContentPermissionRepository is a test double for repository.ContentPermissionRepository.
type MockContentPermissionRepository struct {
	GetByContentFn        func(ctx context.Context, contentType int8, contentID uint64) ([]*entity.ContentPermission, error)
	SetContentPermsFn     func(ctx context.Context, contentType int8, contentID uint64, roleIDs []uint) error
}

func (m *MockContentPermissionRepository) GetByContent(ctx context.Context, contentType int8, contentID uint64) ([]*entity.ContentPermission, error) {
	return m.GetByContentFn(ctx, contentType, contentID)
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
	return m.GetByUserAndUnitFn(ctx, userID, unitID)
}
func (m *MockStudyRecordRepository) ListByUser(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.UserStudyRecord, int64, error) {
	return m.ListByUserFn(ctx, userID, page, pageSize)
}
func (m *MockStudyRecordRepository) Upsert(ctx context.Context, record *entity.UserStudyRecord) error {
	return m.UpsertFn(ctx, record)
}

// MockCollectionRepository is a test double for repository.CollectionRepository.
type MockCollectionRepository struct {
	GetFn    func(ctx context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Collection, error)
	ListFn   func(ctx context.Context, userID uint64, page, pageSize int, contentType *int8) ([]*entity.Collection, int64, error)
	CreateFn func(ctx context.Context, collection *entity.Collection) error
	DeleteFn func(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
}

func (m *MockCollectionRepository) Get(ctx context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Collection, error) {
	return m.GetFn(ctx, userID, contentType, contentID)
}
func (m *MockCollectionRepository) List(ctx context.Context, userID uint64, page, pageSize int, contentType *int8) ([]*entity.Collection, int64, error) {
	return m.ListFn(ctx, userID, page, pageSize, contentType)
}
func (m *MockCollectionRepository) Create(ctx context.Context, collection *entity.Collection) error {
	return m.CreateFn(ctx, collection)
}
func (m *MockCollectionRepository) Delete(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	return m.DeleteFn(ctx, userID, contentType, contentID)
}

// MockLikeRepository is a test double for repository.LikeRepository.
type MockLikeRepository struct {
	GetFn    func(ctx context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Like, error)
	CreateFn func(ctx context.Context, like *entity.Like) error
	DeleteFn func(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
}

func (m *MockLikeRepository) Get(ctx context.Context, userID uint64, contentType int8, contentID uint64) (*entity.Like, error) {
	return m.GetFn(ctx, userID, contentType, contentID)
}
func (m *MockLikeRepository) Create(ctx context.Context, like *entity.Like) error {
	return m.CreateFn(ctx, like)
}
func (m *MockLikeRepository) Delete(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	return m.DeleteFn(ctx, userID, contentType, contentID)
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
	return m.GetByIDFn(ctx, id)
}
func (m *MockCommentRepository) List(ctx context.Context, contentType int8, contentID uint64, page, pageSize int) ([]*entity.Comment, int64, error) {
	return m.ListFn(ctx, contentType, contentID, page, pageSize)
}
func (m *MockCommentRepository) ListAdmin(ctx context.Context, page, pageSize int, status *int8) ([]*entity.Comment, int64, error) {
	return m.ListAdminFn(ctx, page, pageSize, status)
}
func (m *MockCommentRepository) Create(ctx context.Context, comment *entity.Comment) error {
	return m.CreateFn(ctx, comment)
}
func (m *MockCommentRepository) UpdateStatus(ctx context.Context, id uint64, status int8) error {
	return m.UpdateStatusFn(ctx, id, status)
}
func (m *MockCommentRepository) Delete(ctx context.Context, id uint64) error {
	return m.DeleteFn(ctx, id)
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
	return m.GetByIDFn(ctx, id)
}
func (m *MockNotificationRepository) List(ctx context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, error) {
	return m.ListFn(ctx, userID, page, pageSize, isRead)
}
func (m *MockNotificationRepository) UnreadCount(ctx context.Context, userID uint64) (int64, error) {
	return m.UnreadCountFn(ctx, userID)
}
func (m *MockNotificationRepository) MarkRead(ctx context.Context, id uint64) error {
	return m.MarkReadFn(ctx, id)
}
func (m *MockNotificationRepository) MarkAllRead(ctx context.Context, userID uint64) error {
	return m.MarkAllReadFn(ctx, userID)
}

// MockWechatConfigRepository is a test double for repository.WechatConfigRepository.
type MockWechatConfigRepository struct {
	GetFn    func(ctx context.Context) (*entity.WechatConfig, error)
	UpdateFn func(ctx context.Context, cfg *entity.WechatConfig) error
}

func (m *MockWechatConfigRepository) Get(ctx context.Context) (*entity.WechatConfig, error) {
	return m.GetFn(ctx)
}
func (m *MockWechatConfigRepository) Update(ctx context.Context, cfg *entity.WechatConfig) error {
	return m.UpdateFn(ctx, cfg)
}

// MockAuditLogRepository is a test double for repository.AuditLogRepository.
type MockAuditLogRepository struct {
	ListFn   func(ctx context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error)
	CreateFn func(ctx context.Context, log *entity.AuditLog) error
}

func (m *MockAuditLogRepository) List(ctx context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error) {
	return m.ListFn(ctx, page, pageSize, module, action, startTime, endTime)
}
func (m *MockAuditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	return m.CreateFn(ctx, log)
}

// MockLogConfigRepository is a test double for repository.LogConfigRepository.
type MockLogConfigRepository struct {
	GetFn    func(ctx context.Context) (*entity.LogConfig, error)
	UpdateFn func(ctx context.Context, cfg *entity.LogConfig) error
}

func (m *MockLogConfigRepository) Get(ctx context.Context) (*entity.LogConfig, error) {
	return m.GetFn(ctx)
}
func (m *MockLogConfigRepository) Update(ctx context.Context, cfg *entity.LogConfig) error {
	return m.UpdateFn(ctx, cfg)
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
	return m.GetProfileFn(ctx, userID)
}
func (m *MockUserService) UpdateProfile(ctx context.Context, userID uint64, req *dto.UserProfileUpdateRequest) error {
	return m.UpdateProfileFn(ctx, userID, req)
}
func (m *MockUserService) GetPermissions(ctx context.Context, userID uint64) ([]string, []string, error) {
	return m.GetPermissionsFn(ctx, userID)
}
func (m *MockUserService) List(ctx context.Context, page, pageSize int, keyword string, userType, status *int8) ([]*entity.User, int64, error) {
	return m.ListFn(ctx, page, pageSize, keyword, userType, status)
}
func (m *MockUserService) GetByID(ctx context.Context, id uint64) (*entity.User, error) {
	return m.GetByIDFn(ctx, id)
}
func (m *MockUserService) CreateAdminUser(ctx context.Context, req *dto.CreateAdminUserRequest) (uint64, error) {
	return m.CreateAdminUserFn(ctx, req)
}
func (m *MockUserService) UpdateUser(ctx context.Context, id uint64, req *dto.UpdateUserRequest, operatorID uint64) error {
	return m.UpdateUserFn(ctx, id, req, operatorID)
}
func (m *MockUserService) DeleteUser(ctx context.Context, id uint64) error {
	return m.DeleteUserFn(ctx, id)
}
func (m *MockUserService) AssignRoles(ctx context.Context, userID uint64, req *dto.AssignRolesRequest) error {
	return m.AssignRolesFn(ctx, userID, req)
}
func (m *MockUserService) AddTag(ctx context.Context, userID uint64, req *dto.AddTagRequest) (uint, error) {
	return m.AddTagFn(ctx, userID, req)
}
func (m *MockUserService) DeleteTag(ctx context.Context, userID, tagID uint64) error {
	return m.DeleteTagFn(ctx, userID, tagID)
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
	return m.ListFn(ctx)
}
func (m *MockRoleService) GetByID(ctx context.Context, id uint) (*entity.Role, error) {
	return m.GetByIDFn(ctx, id)
}
func (m *MockRoleService) Create(ctx context.Context, req *dto.CreateRoleRequest) (uint, error) {
	return m.CreateFn(ctx, req)
}
func (m *MockRoleService) Update(ctx context.Context, id uint, req *dto.UpdateRoleRequest) error {
	return m.UpdateFn(ctx, id, req)
}
func (m *MockRoleService) Delete(ctx context.Context, id uint) error {
	return m.DeleteFn(ctx, id)
}

// MockPermissionService is a test double for service.PermissionService.
type MockPermissionService struct {
	GetTreeFn func(ctx context.Context) ([]*entity.Permission, error)
}

func (m *MockPermissionService) GetTree(ctx context.Context) ([]*entity.Permission, error) {
	return m.GetTreeFn(ctx)
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
	return m.ListFn(ctx, status)
}
func (m *MockModuleService) Create(ctx context.Context, req *dto.CreateModuleRequest) (uint, error) {
	return m.CreateFn(ctx, req)
}
func (m *MockModuleService) Update(ctx context.Context, id uint, req *dto.CreateModuleRequest) error {
	return m.UpdateFn(ctx, id, req)
}
func (m *MockModuleService) Delete(ctx context.Context, id uint) error {
	return m.DeleteFn(ctx, id)
}
func (m *MockModuleService) GetPages(ctx context.Context, moduleID uint) ([]*entity.ModulePage, error) {
	return m.GetPagesFn(ctx, moduleID)
}
func (m *MockModuleService) CreatePage(ctx context.Context, moduleID uint, req *dto.CreateModulePageRequest) (uint, error) {
	return m.CreatePageFn(ctx, moduleID, req)
}
func (m *MockModuleService) UpdatePage(ctx context.Context, moduleID, pageID uint, req *dto.CreateModulePageRequest) error {
	return m.UpdatePageFn(ctx, moduleID, pageID, req)
}
func (m *MockModuleService) DeletePage(ctx context.Context, moduleID, pageID uint) error {
	return m.DeletePageFn(ctx, moduleID, pageID)
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
	return m.ListFn(ctx, page, pageSize, keyword, moduleID, sort, userID)
}
func (m *MockArticleService) GetByID(ctx context.Context, id uint64, userID *uint64) (*entity.Article, error) {
	return m.GetByIDFn(ctx, id, userID)
}
func (m *MockArticleService) AdminList(ctx context.Context, page, pageSize int, keyword string, moduleID *uint, status *int8) ([]*entity.Article, int64, error) {
	return m.AdminListFn(ctx, page, pageSize, keyword, moduleID, status)
}
func (m *MockArticleService) AdminGetByID(ctx context.Context, id uint64) (*entity.Article, error) {
	return m.AdminGetByIDFn(ctx, id)
}
func (m *MockArticleService) Create(ctx context.Context, req *dto.CreateArticleRequest, authorID uint64) (uint64, error) {
	return m.CreateFn(ctx, req, authorID)
}
func (m *MockArticleService) Update(ctx context.Context, id uint64, req *dto.UpdateArticleRequest) error {
	return m.UpdateFn(ctx, id, req)
}
func (m *MockArticleService) Delete(ctx context.Context, id uint64) error {
	return m.DeleteFn(ctx, id)
}
func (m *MockArticleService) Publish(ctx context.Context, id uint64, req *dto.PublishArticleRequest) error {
	return m.PublishFn(ctx, id, req)
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
	return m.ListFn(ctx, page, pageSize, keyword, moduleID, isFree, userID)
}
func (m *MockCourseService) GetByID(ctx context.Context, id uint64, userID *uint64) (*entity.Course, error) {
	return m.GetByIDFn(ctx, id, userID)
}
func (m *MockCourseService) AdminList(ctx context.Context, page, pageSize int, keyword string, status *int8) ([]*entity.Course, int64, error) {
	return m.AdminListFn(ctx, page, pageSize, keyword, status)
}
func (m *MockCourseService) AdminGetByID(ctx context.Context, id uint64) (*entity.Course, error) {
	return m.AdminGetByIDFn(ctx, id)
}
func (m *MockCourseService) Create(ctx context.Context, req *dto.CreateCourseRequest, authorID uint64) (uint64, error) {
	return m.CreateFn(ctx, req, authorID)
}
func (m *MockCourseService) Update(ctx context.Context, id uint64, req *dto.UpdateCourseRequest) error {
	return m.UpdateFn(ctx, id, req)
}
func (m *MockCourseService) Delete(ctx context.Context, id uint64) error {
	return m.DeleteFn(ctx, id)
}
func (m *MockCourseService) Publish(ctx context.Context, id uint64, req *dto.PublishCourseRequest) error {
	return m.PublishFn(ctx, id, req)
}
func (m *MockCourseService) GetUnits(ctx context.Context, courseID uint64) ([]*entity.CourseUnit, error) {
	return m.GetUnitsFn(ctx, courseID)
}
func (m *MockCourseService) CreateUnit(ctx context.Context, courseID uint64, req *dto.CreateCourseUnitRequest) (uint64, error) {
	return m.CreateUnitFn(ctx, courseID, req)
}
func (m *MockCourseService) UpdateUnit(ctx context.Context, courseID, unitID uint64, req *dto.CreateCourseUnitRequest) error {
	return m.UpdateUnitFn(ctx, courseID, unitID, req)
}
func (m *MockCourseService) DeleteUnit(ctx context.Context, courseID, unitID uint64) error {
	return m.DeleteUnitFn(ctx, courseID, unitID)
}

// MockStudyRecordService is a test double for service.StudyRecordService.
type MockStudyRecordService struct {
	ListFn   func(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.UserStudyRecord, int64, error)
	UpdateFn func(ctx context.Context, userID uint64, req *dto.UpdateStudyRecordRequest) error
}

func (m *MockStudyRecordService) List(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.UserStudyRecord, int64, error) {
	return m.ListFn(ctx, userID, page, pageSize)
}
func (m *MockStudyRecordService) Update(ctx context.Context, userID uint64, req *dto.UpdateStudyRecordRequest) error {
	return m.UpdateFn(ctx, userID, req)
}

// MockCollectionService is a test double for service.CollectionService.
type MockCollectionService struct {
	ListFn   func(ctx context.Context, userID uint64, page, pageSize int, contentType *int8) ([]*entity.Collection, int64, error)
	AddFn    func(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
	RemoveFn func(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
}

func (m *MockCollectionService) List(ctx context.Context, userID uint64, page, pageSize int, contentType *int8) ([]*entity.Collection, int64, error) {
	return m.ListFn(ctx, userID, page, pageSize, contentType)
}
func (m *MockCollectionService) Add(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	return m.AddFn(ctx, userID, contentType, contentID)
}
func (m *MockCollectionService) Remove(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	return m.RemoveFn(ctx, userID, contentType, contentID)
}

// MockLikeService is a test double for service.LikeService.
type MockLikeService struct {
	AddFn    func(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
	RemoveFn func(ctx context.Context, userID uint64, contentType int8, contentID uint64) error
}

func (m *MockLikeService) Add(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	return m.AddFn(ctx, userID, contentType, contentID)
}
func (m *MockLikeService) Remove(ctx context.Context, userID uint64, contentType int8, contentID uint64) error {
	return m.RemoveFn(ctx, userID, contentType, contentID)
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
	return m.ListFn(ctx, contentType, contentID, page, pageSize)
}
func (m *MockCommentService) Create(ctx context.Context, userID uint64, contentType int8, contentID uint64, req *dto.CreateCommentRequest) (*entity.Comment, error) {
	return m.CreateFn(ctx, userID, contentType, contentID, req)
}
func (m *MockCommentService) AdminList(ctx context.Context, page, pageSize int, status *int8) ([]*entity.Comment, int64, error) {
	return m.AdminListFn(ctx, page, pageSize, status)
}
func (m *MockCommentService) Audit(ctx context.Context, id uint64, req *dto.AuditCommentRequest) error {
	return m.AuditFn(ctx, id, req)
}
func (m *MockCommentService) Delete(ctx context.Context, id uint64) error {
	return m.DeleteFn(ctx, id)
}

// MockNotificationService is a test double for service.NotificationService.
type MockNotificationService struct {
	ListFn       func(ctx context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, int64, error)
	MarkReadFn   func(ctx context.Context, id uint64) error
	MarkAllReadFn func(ctx context.Context, userID uint64) error
}

func (m *MockNotificationService) List(ctx context.Context, userID uint64, page, pageSize int, isRead *bool) ([]*entity.Notification, int64, int64, error) {
	return m.ListFn(ctx, userID, page, pageSize, isRead)
}
func (m *MockNotificationService) MarkRead(ctx context.Context, id uint64) error {
	return m.MarkReadFn(ctx, id)
}
func (m *MockNotificationService) MarkAllRead(ctx context.Context, userID uint64) error {
	return m.MarkAllReadFn(ctx, userID)
}

// MockWechatConfigService is a test double for service.WechatConfigService.
type MockWechatConfigService struct {
	GetFn    func(ctx context.Context) (*entity.WechatConfig, error)
	UpdateFn func(ctx context.Context, req *dto.UpdateWechatConfigRequest) error
}

func (m *MockWechatConfigService) Get(ctx context.Context) (*entity.WechatConfig, error) {
	return m.GetFn(ctx)
}
func (m *MockWechatConfigService) Update(ctx context.Context, req *dto.UpdateWechatConfigRequest) error {
	return m.UpdateFn(ctx, req)
}

// MockAuditLogService is a test double for service.AuditLogService.
type MockAuditLogService struct {
	ListFn func(ctx context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error)
	LogFn  func(ctx context.Context, log *entity.AuditLog)
}

func (m *MockAuditLogService) List(ctx context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error) {
	return m.ListFn(ctx, page, pageSize, module, action, startTime, endTime)
}
func (m *MockAuditLogService) Log(ctx context.Context, log *entity.AuditLog) {
	m.LogFn(ctx, log)
}

// MockLogConfigService is a test double for service.LogConfigService.
type MockLogConfigService struct {
	GetFn    func(ctx context.Context) (*entity.LogConfig, error)
	UpdateFn func(ctx context.Context, req *dto.UpdateLogConfigRequest) error
}

func (m *MockLogConfigService) Get(ctx context.Context) (*entity.LogConfig, error) {
	return m.GetFn(ctx)
}
func (m *MockLogConfigService) Update(ctx context.Context, req *dto.UpdateLogConfigRequest) error {
	return m.UpdateFn(ctx, req)
}
