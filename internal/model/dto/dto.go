package dto

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// ==================== Auth DTOs ====================

// WechatLoginRequest is the request body for wechat login.
type WechatLoginRequest struct {
	Code          string `json:"code"`
	EncryptedData string `json:"encrypted_data"`
	IV            string `json:"iv"`
}

// Validate validates WechatLoginRequest.
func (r WechatLoginRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Code, validation.Required),
	)
}

// AdminLoginRequest is the request body for admin login.
type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Captcha  string `json:"captcha"`
}

// Validate validates AdminLoginRequest.
func (r AdminLoginRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Email, validation.Required, is.Email),
		validation.Field(&r.Password, validation.Required, validation.Length(6, 128)),
	)
}

// LoginResponseData is the data field of the login response.
type LoginResponseData struct {
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int         `json:"expires_in"`
	UserInfo    interface{} `json:"user_info"`
}

// ==================== User DTOs ====================

// UserProfileUpdateRequest is the request body for updating user profile.
type UserProfileUpdateRequest struct {
	Nickname     string `json:"nickname"`
	AvatarURL    string `json:"avatar_url"`
	AvatarFileID uint64 `json:"avatar_file_id"`
}

// Validate validates UserProfileUpdateRequest.
func (r UserProfileUpdateRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Nickname, validation.Length(0, 64)),
		validation.Field(&r.AvatarURL, validation.Length(0, 255)),
	)
}

// AttachmentPermissionRequest defines optional role permissions for a file attachment.
type AttachmentPermissionRequest struct {
	FileID          uint64 `json:"file_id"`
	RolePermissions []uint `json:"role_permissions"`
}

// CreateAdminUserRequest is the request body for creating an admin user.
type CreateAdminUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	UserType int8   `json:"user_type"`
}

// Validate validates CreateAdminUserRequest.
func (r CreateAdminUserRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Email, validation.Required, is.Email),
		validation.Field(&r.Password, validation.Required, validation.Length(6, 128)),
		validation.Field(&r.UserType, validation.Required, validation.In(int8(2), int8(3))),
	)
}

// UpdateUserRequest is the request body for updating a user.
type UpdateUserRequest struct {
	Nickname      string     `json:"nickname"`
	UserType      int8       `json:"user_type"`
	Status        int8       `json:"status"`
	FreezeEndTime *time.Time `json:"freeze_end_time"`
}

// Validate validates UpdateUserRequest.
func (r UpdateUserRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Nickname, validation.Length(0, 64)),
	)
}

// AssignRolesRequest is the request body for assigning roles to a user.
type AssignRolesRequest struct {
	RoleIDs []uint `json:"role_ids"`
}

// Validate validates AssignRolesRequest.
func (r AssignRolesRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.RoleIDs, validation.Required),
	)
}

// AddTagRequest is the request body for adding a tag to a user.
type AddTagRequest struct {
	TagName string `json:"tag_name"`
}

// Validate validates AddTagRequest.
func (r AddTagRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.TagName, validation.Required, validation.Length(1, 32)),
	)
}

// ==================== Role DTOs ====================

// CreateRoleRequest is the request body for creating a role.
type CreateRoleRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	ParentID      uint   `json:"parent_id"`
	PermissionIDs []uint `json:"permission_ids"`
}

// Validate validates CreateRoleRequest.
func (r CreateRoleRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required, validation.Length(1, 64)),
		validation.Field(&r.Description, validation.Length(0, 255)),
	)
}

// UpdateRoleRequest is the request body for updating a role.
type UpdateRoleRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	ParentID      uint   `json:"parent_id"`
	PermissionIDs []uint `json:"permission_ids"`
}

// Validate validates UpdateRoleRequest.
func (r UpdateRoleRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required, validation.Length(1, 64)),
		validation.Field(&r.Description, validation.Length(0, 255)),
	)
}

// ==================== Module DTOs ====================

// CreateModuleRequest is the request body for creating a module.
type CreateModuleRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

// Validate validates CreateModuleRequest.
func (r CreateModuleRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Title, validation.Required, validation.Length(1, 128)),
	)
}

// CreateModulePageRequest is the request body for creating a module page.
type CreateModulePageRequest struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	ContentType int8   `json:"content_type"`
	SortOrder   int    `json:"sort_order"`
}

// Validate validates CreateModulePageRequest.
func (r CreateModulePageRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Title, validation.Required, validation.Length(1, 128)),
		validation.Field(&r.Content, validation.Required),
	)
}

// ==================== Banner DTOs ====================

// CreateBannerRequest is the request body for creating/updating a banner.
type CreateBannerRequest struct {
	Title       string `json:"title"`
	ImageFileID uint64 `json:"image_file_id"`
	LinkURL     string `json:"link_url"`
	SortOrder   int    `json:"sort_order"`
	Status      int8   `json:"status"`
}

// Validate validates CreateBannerRequest.
func (r CreateBannerRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Title, validation.Required, validation.Length(1, 128)),
		validation.Field(&r.ImageFileID, validation.Required, validation.Min(uint64(1))),
		validation.Field(&r.Status, validation.In(int8(0), int8(1))),
		validation.Field(&r.LinkURL, validation.Length(0, 255)),
	)
}

// ==================== Article DTOs ====================

// CreateArticleRequest is the request body for creating an article.
type CreateArticleRequest struct {
	Title                 string                        `json:"title"`
	Summary               string                        `json:"summary"`
	Content               string                        `json:"content"`
	ContentType           int8                          `json:"content_type"`
	CoverImage            string                        `json:"cover_image"`
	CoverFileID           uint64                        `json:"cover_file_id"`
	ModuleID              uint                          `json:"module_id"`
	Status                int8                          `json:"status"`
	PublishTime           *time.Time                    `json:"publish_time"`
	AttachmentFileIDs     []uint64                      `json:"attachment_file_ids"`
	AttachmentPermissions []AttachmentPermissionRequest `json:"attachment_permissions"`
	RolePermissions       []uint                        `json:"role_permissions"`
}

// Validate validates CreateArticleRequest.
func (r CreateArticleRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Title, validation.Required, validation.Length(1, 200)),
		validation.Field(&r.Content, validation.Required),
		validation.Field(&r.Summary, validation.Length(0, 500)),
		validation.Field(&r.CoverImage, validation.Length(0, 255)),
	)
}

// UpdateArticleRequest is the request body for updating an article.
type UpdateArticleRequest struct {
	Title                 string                        `json:"title"`
	Summary               string                        `json:"summary"`
	Content               string                        `json:"content"`
	ContentType           int8                          `json:"content_type"`
	CoverImage            string                        `json:"cover_image"`
	CoverFileID           uint64                        `json:"cover_file_id"`
	ModuleID              uint                          `json:"module_id"`
	Status                int8                          `json:"status"`
	PublishTime           *time.Time                    `json:"publish_time"`
	AttachmentFileIDs     []uint64                      `json:"attachment_file_ids"`
	AttachmentPermissions []AttachmentPermissionRequest `json:"attachment_permissions"`
	RolePermissions       []uint                        `json:"role_permissions"`
}

// Validate validates UpdateArticleRequest.
func (r UpdateArticleRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Title, validation.Required, validation.Length(1, 200)),
		validation.Field(&r.Content, validation.Required),
		validation.Field(&r.Summary, validation.Length(0, 500)),
		validation.Field(&r.CoverImage, validation.Length(0, 255)),
	)
}

// PublishArticleRequest is the request body for publishing/unpublishing an article.
type PublishArticleRequest struct {
	Status int8 `json:"status"`
}

// Validate validates PublishArticleRequest.
func (r PublishArticleRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Status, validation.Required, validation.In(int8(0), int8(1))),
	)
}

// PinArticleRequest is the request body for pinning article order.
type PinArticleRequest struct {
	SortOrder int `json:"sort_order"`
}

// Validate validates PinArticleRequest.
func (r PinArticleRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.SortOrder, validation.Required),
	)
}

// ==================== Course DTOs ====================

// CreateCourseRequest is the request body for creating a course.
type CreateCourseRequest struct {
	Title                 string                        `json:"title"`
	Description           string                        `json:"description"`
	CoverImage            string                        `json:"cover_image"`
	CoverFileID           uint64                        `json:"cover_file_id"`
	Price                 float64                       `json:"price"`
	ModuleID              uint                          `json:"module_id"`
	Status                int8                          `json:"status"`
	PublishTime           *time.Time                    `json:"publish_time"`
	AttachmentFileIDs     []uint64                      `json:"attachment_file_ids"`
	AttachmentPermissions []AttachmentPermissionRequest `json:"attachment_permissions"`
	RolePermissions       []uint                        `json:"role_permissions"`
}

// Validate validates CreateCourseRequest.
func (r CreateCourseRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Title, validation.Required, validation.Length(1, 200)),
		validation.Field(&r.Price, validation.Min(0.0)),
	)
}

// UpdateCourseRequest is the request body for updating a course.
type UpdateCourseRequest struct {
	Title                 string                        `json:"title"`
	Description           string                        `json:"description"`
	CoverImage            string                        `json:"cover_image"`
	CoverFileID           uint64                        `json:"cover_file_id"`
	Price                 float64                       `json:"price"`
	ModuleID              uint                          `json:"module_id"`
	Status                int8                          `json:"status"`
	PublishTime           *time.Time                    `json:"publish_time"`
	AttachmentFileIDs     []uint64                      `json:"attachment_file_ids"`
	AttachmentPermissions []AttachmentPermissionRequest `json:"attachment_permissions"`
	RolePermissions       []uint                        `json:"role_permissions"`
}

// Validate validates UpdateCourseRequest.
func (r UpdateCourseRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Title, validation.Required, validation.Length(1, 200)),
		validation.Field(&r.Price, validation.Min(0.0)),
	)
}

// PublishCourseRequest is the request body for publishing/unpublishing a course.
type PublishCourseRequest struct {
	Status int8 `json:"status"`
}

// Validate validates PublishCourseRequest.
func (r PublishCourseRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Status, validation.Required, validation.In(int8(0), int8(1))),
	)
}

// PinCourseRequest is the request body for pinning course order.
type PinCourseRequest struct {
	SortOrder int `json:"sort_order"`
}

// Validate validates PinCourseRequest.
func (r PinCourseRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.SortOrder, validation.Required),
	)
}

// CreateCourseUnitRequest is the request body for creating a course unit.
type CreateCourseUnitRequest struct {
	Title                 string                        `json:"title"`
	VideoFileID           uint64                        `json:"video_file_id"`
	Duration              uint                          `json:"duration"`
	SortOrder             int                           `json:"sort_order"`
	AttachmentFileIDs     []uint64                      `json:"attachment_file_ids"`
	AttachmentPermissions []AttachmentPermissionRequest `json:"attachment_permissions"`
	RolePermissions       []uint                        `json:"role_permissions"`
}

// Validate validates CreateCourseUnitRequest.
func (r CreateCourseUnitRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Title, validation.Required, validation.Length(1, 200)),
	)
}

// ==================== Interaction DTOs ====================

// UpdateStudyRecordRequest is the request body for updating a study record.
type UpdateStudyRecordRequest struct {
	UnitID   uint64 `json:"unit_id"`
	Progress uint   `json:"progress"`
	Status   int8   `json:"status"`
}

// Validate validates UpdateStudyRecordRequest.
func (r UpdateStudyRecordRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.UnitID, validation.Required),
		validation.Field(&r.Progress, validation.Required),
	)
}

// CreateCommentRequest is the request body for creating a comment.
type CreateCommentRequest struct {
	Content  string `json:"content"`
	ParentID uint64 `json:"parent_id"`
}

// Validate validates CreateCommentRequest.
func (r CreateCommentRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Content, validation.Required, validation.Length(1, 1000)),
	)
}

// AuditCommentRequest is the request body for auditing a comment.
type AuditCommentRequest struct {
	Status int8 `json:"status"`
}

// Validate validates AuditCommentRequest.
func (r AuditCommentRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Status, validation.Required, validation.In(int8(1), int8(2))),
	)
}

// ==================== System DTOs ====================

// UpdateWechatConfigRequest is the request body for updating wechat config.
type UpdateWechatConfigRequest struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
	APIToken  string `json:"api_token"`
}

// Validate validates UpdateWechatConfigRequest.
func (r UpdateWechatConfigRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.AppID, validation.Required, validation.Length(1, 32)),
		validation.Field(&r.AppSecret, validation.Required, validation.Length(1, 64)),
	)
}

// UpdateLogConfigRequest is the request body for updating log config.
type UpdateLogConfigRequest struct {
	RetentionDays int `json:"retention_days"`
}

// Validate validates UpdateLogConfigRequest.
func (r UpdateLogConfigRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.RetentionDays, validation.Required, validation.Min(1), validation.Max(3650)),
	)
}

// ListQuery holds common pagination parameters.
type ListQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
	Sort     string `form:"sort"`
}

// GetPage returns page with default 1.
func (q *ListQuery) GetPage() int {
	if q.Page <= 0 {
		return 1
	}
	return q.Page
}

// GetPageSize returns page_size with default 20.
func (q *ListQuery) GetPageSize() int {
	if q.PageSize <= 0 {
		return 20
	}
	if q.PageSize > 100 {
		return 100
	}
	return q.PageSize
}

// GetOffset returns the offset for the query.
func (q *ListQuery) GetOffset() int {
	return (q.GetPage() - 1) * q.GetPageSize()
}

// ==================== Attribute DTOs ====================

// CreateAttributeRequest is the request body for creating an attribute.
type CreateAttributeRequest struct {
	Name string `json:"name"`
	Type int8   `json:"type"`
}

// Validate validates CreateAttributeRequest.
func (r CreateAttributeRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required, validation.Length(1, 64)),
		validation.Field(&r.Type, validation.In(int8(0), int8(1), int8(2))),
	)
}

// UpdateAttributeRequest is the request body for updating an attribute.
type UpdateAttributeRequest struct {
	Name string `json:"name"`
	Type int8   `json:"type"`
}

// Validate validates UpdateAttributeRequest.
func (r UpdateAttributeRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required, validation.Length(1, 64)),
		validation.Field(&r.Type, validation.In(int8(0), int8(1), int8(2))),
	)
}

// SetUserAttributeRequest is the request body for setting a user attribute value.
type SetUserAttributeRequest struct {
	AttributeID uint   `json:"attribute_id"`
	Value       string `json:"value"`
	ValueString string `json:"value_string"`
	ValueBigint *int64 `json:"value_bigint"`
}

// Validate validates SetUserAttributeRequest.
func (r SetUserAttributeRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.AttributeID, validation.Required),
		validation.Field(&r.Value, validation.Length(0, 255)),
	)
}
