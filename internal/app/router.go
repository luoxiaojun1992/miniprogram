package app

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
)

// InitRouter sets up the Gin router with all routes.
func InitRouter(p *Provider) *gin.Engine {
	gin.SetMode(p.Config.Server.Mode)
	r := gin.New()

	// Global middleware
	r.Use(middleware.RecoveryMiddleware(p.Log))
	r.Use(middleware.ErrorMiddleware(p.Log))
	r.Use(middleware.CorsMiddleware())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware(p.Log))
	if p.Config.RateLimit.Enabled {
		r.Use(middleware.RateLimitMiddleware(
			middleware.NewRedisRateLimitStore(p.Redis),
			int64(p.Config.RateLimit.Requests),
			time.Duration(p.Config.RateLimit.WindowSeconds)*time.Second,
			p.Log,
		))
	}

	// Health check
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/v1")
	optionalJWT := middleware.OptionalJWTAuthMiddleware(p.Config.JWT.Secret)
	requiredJWT := middleware.JWTAuthMiddleware(p.Config.JWT.Secret)

	// ==================== Auth ====================
	auth := v1.Group("/auth")
	{
		auth.POST("/wechat-login", p.AuthCtrl.WechatLogin)
		auth.POST("/admin-login", p.AuthCtrl.AdminLogin)
		auth.POST("/refresh", requiredJWT, p.AuthCtrl.RefreshToken)
	}

	// ==================== Must JWT APIs ====================
	users := v1.Group("/users", requiredJWT)
	{
		users.GET("/profile", p.UserCtrl.GetProfile)
		users.PUT("/profile", p.UserCtrl.UpdateProfile)
		users.GET("/permissions", p.UserCtrl.GetPermissions)
	}

	// ==================== Optional JWT APIs ====================
	optional := v1.Group("", optionalJWT)
	{
		optional.GET("/modules", p.ModuleCtrl.List)
		optional.GET("/banners", p.BannerCtrl.List)

		articles := optional.Group("/articles")
		{
			articles.GET("", p.ArticleCtrl.List)
			articles.GET("/:id", p.ArticleCtrl.GetByID)
			articles.GET("/:id/attachments", p.ArticleCtrl.GetAttachments)
		}

		courses := optional.Group("/courses")
		{
			courses.GET("", p.CourseCtrl.List)
			courses.GET("/:id", p.CourseCtrl.GetByID)
			courses.GET("/:id/units", p.CourseCtrl.GetUnits)
			courses.GET("/:id/attachments", p.CourseCtrl.GetAttachments)
			courses.GET("/:id/units/:unit_id/attachments", p.CourseCtrl.GetUnitAttachments)
		}

		optional.GET("/download/course/video/:file_id", p.UploadCtrl.GenerateCourseVideoDownloadURL)
		optional.GET("/download/article/attachment/:file_id", p.UploadCtrl.GenerateArticleAttachmentDownloadURL)
		optional.GET("/download/course/attachment/:file_id", p.UploadCtrl.GenerateCourseAttachmentDownloadURL)
		optional.GET("/download/course/unit/attachment/:file_id", p.UploadCtrl.GenerateCourseUnitAttachmentDownloadURL)
		optional.GET("/download/banner/media/:file_id", p.UploadCtrl.GenerateBannerMediaDownloadURL)
		optional.GET("/comments/:content_type/:content_id", p.CommentCtrl.List)
	}

	// ==================== JWT Required APIs ====================
	authRequired := v1.Group("", requiredJWT)
	{
		// Study records
		authRequired.GET("/study-records", p.StudyRecordCtrl.List)
		authRequired.POST("/study-records", p.StudyRecordCtrl.Update)

		// Collections
		authRequired.GET("/collections", p.CollectionCtrl.List)
		authRequired.POST("/collections/:content_type/:content_id", p.CollectionCtrl.Add)
		authRequired.DELETE("/collections/:content_type/:content_id", p.CollectionCtrl.Remove)

		// Likes
		authRequired.POST("/likes/:content_type/:content_id", p.LikeCtrl.Add)
		authRequired.DELETE("/likes/:content_type/:content_id", p.LikeCtrl.Remove)

		// Comments
		authRequired.POST("/comments/:content_type/:content_id", p.CommentCtrl.Create)

		// Notifications
		authRequired.GET("/notifications", p.NotificationCtrl.List)
		authRequired.PUT("/notifications/read-all", p.NotificationCtrl.MarkAllRead)
		authRequired.PUT("/notifications/:id/read", p.NotificationCtrl.MarkRead)

		// Upload
		authRequired.POST("/upload/avatar", p.UploadCtrl.UploadAvatar)
	}
	v1.GET("/download/static/:file_id", p.UploadCtrl.GenerateStaticMaterialURL)

	// ==================== Admin ====================
	admin := v1.Group(
		"/admin",
		requiredJWT,
		middleware.RequireAdmin(),
		middleware.AuditLogMiddleware(p.AuditLogSvc),
	)
	{
		// Users
		admin.GET("/users", p.UserCtrl.AdminListUsers)
		admin.POST("/users", p.UserCtrl.AdminCreateUser)
		admin.GET("/users/:id", p.UserCtrl.AdminGetUser)
		admin.PUT("/users/:id", p.UserCtrl.AdminUpdateUser)
		admin.DELETE("/users/:id", p.UserCtrl.AdminDeleteUser)
		admin.PUT("/users/:id/roles", p.UserCtrl.AdminAssignRoles)
		admin.POST("/users/:id/tags", p.UserCtrl.AdminAddUserTag)
		admin.DELETE("/users/:id/tags", p.UserCtrl.AdminDeleteUserTag)
		admin.GET("/users/:id/attributes", p.AttributeCtrl.ListUserAttributes)
		admin.POST("/users/:id/attributes", p.AttributeCtrl.SetUserAttribute)
		admin.DELETE("/users/:id/attributes", p.AttributeCtrl.DeleteUserAttribute)

		// Roles
		admin.GET("/roles", p.RoleCtrl.List)
		admin.POST("/roles", p.RoleCtrl.Create)
		admin.GET("/roles/:id", p.RoleCtrl.GetByID)
		admin.PUT("/roles/:id", p.RoleCtrl.Update)
		admin.DELETE("/roles/:id", p.RoleCtrl.Delete)

		// Permissions
		admin.GET("/permissions", p.PermissionCtrl.GetTree)

		// Modules
		admin.POST("/modules", p.ModuleCtrl.Create)
		admin.PUT("/modules/:id", p.ModuleCtrl.Update)
		admin.DELETE("/modules/:id", p.ModuleCtrl.Delete)
		admin.GET("/modules/:id/pages", p.ModuleCtrl.GetPages)
		admin.POST("/modules/:id/pages", p.ModuleCtrl.CreatePage)
		admin.PUT("/modules/:id/pages/:page_id", p.ModuleCtrl.UpdatePage)
		admin.DELETE("/modules/:id/pages/:page_id", p.ModuleCtrl.DeletePage)

		// Banners
		admin.GET("/banners", p.BannerCtrl.AdminList)
		admin.POST("/banners", p.BannerCtrl.AdminCreate)
		admin.PUT("/banners/:id", p.BannerCtrl.AdminUpdate)
		admin.DELETE("/banners/:id", p.BannerCtrl.AdminDelete)

		// Articles
		admin.GET("/articles", p.ArticleCtrl.AdminList)
		admin.POST("/articles", p.ArticleCtrl.AdminCreate)
		admin.GET("/articles/:id", p.ArticleCtrl.AdminGetByID)
		admin.PUT("/articles/:id", p.ArticleCtrl.AdminUpdate)
		admin.DELETE("/articles/:id", p.ArticleCtrl.AdminDelete)
		admin.POST("/articles/:id/publish", p.ArticleCtrl.AdminPublish)
		admin.POST("/articles/:id/pin", p.ArticleCtrl.AdminPin)
		admin.POST("/articles/:id/copy", p.ArticleCtrl.AdminCopy)

		// Courses
		admin.GET("/courses", p.CourseCtrl.AdminList)
		admin.POST("/courses", p.CourseCtrl.AdminCreate)
		admin.GET("/courses/:id", p.CourseCtrl.AdminGetByID)
		admin.PUT("/courses/:id", p.CourseCtrl.AdminUpdate)
		admin.DELETE("/courses/:id", p.CourseCtrl.AdminDelete)
		admin.POST("/courses/:id/publish", p.CourseCtrl.AdminPublish)
		admin.POST("/courses/:id/pin", p.CourseCtrl.AdminPin)
		admin.POST("/courses/:id/copy", p.CourseCtrl.AdminCopy)
		admin.GET("/courses/:id/units", p.CourseCtrl.AdminGetUnits)
		admin.POST("/courses/:id/units", p.CourseCtrl.AdminCreateUnit)
		admin.PUT("/courses/:id/units/:unit_id", p.CourseCtrl.AdminUpdateUnit)
		admin.DELETE("/courses/:id/units/:unit_id", p.CourseCtrl.AdminDeleteUnit)

		// Comments
		admin.GET("/comments", p.CommentCtrl.AdminList)
		admin.PUT("/comments/:id/audit", p.CommentCtrl.AdminAudit)
		admin.DELETE("/comments/:id", p.CommentCtrl.AdminDelete)

		// System
		admin.GET("/wechat-config", p.SystemCtrl.GetWechatConfig)
		admin.PUT("/wechat-config", p.SystemCtrl.UpdateWechatConfig)
		admin.GET("/audit-logs", p.SystemCtrl.ListAuditLogs)
		admin.GET("/log-config", p.SystemCtrl.GetLogConfig)
		admin.PUT("/log-config", p.SystemCtrl.UpdateLogConfig)

		// Upload
		admin.GET("/upload/files/presign", p.UploadCtrl.GenerateAdminUploadPresignURL)
		admin.GET("/upload/banner/media/presign", p.UploadCtrl.GenerateBannerMediaPresignURL)
		admin.GET("/upload/course/video/presign", p.UploadCtrl.GenerateCourseVideoPresignURL)
		admin.GET("/upload/article/attachment/presign", p.UploadCtrl.GenerateArticleAttachmentPresignURL)
		admin.GET("/upload/course/attachment/presign", p.UploadCtrl.GenerateCourseAttachmentPresignURL)
		admin.GET("/upload/course/unit/attachment/presign", p.UploadCtrl.GenerateCourseUnitAttachmentPresignURL)

		// Attributes
		admin.GET("/attributes", p.AttributeCtrl.List)
		admin.POST("/attributes", p.AttributeCtrl.Create)
		admin.PUT("/attributes/:id", p.AttributeCtrl.Update)
		admin.DELETE("/attributes/:id", p.AttributeCtrl.Delete)
	}

	// ==================== Debug (disabled by default) ====================
	// This route MUST only be enabled in non-production environments.
	if p.Config.Debug.EnableTestToken {
		v1.POST("/debug/token", p.DebugCtrl.GenerateTestToken)
	}

	return r
}
