package app

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/luoxiaojun1992/miniprogram/internal/controller"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/cosutil"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/wechat"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
	"github.com/luoxiaojun1992/miniprogram/internal/service"
)

// Provider holds all initialized components.
type Provider struct {
	Config *Config
	Log    *logrus.Logger
	DB     *gorm.DB
	Redis  redis.UniversalClient

	// Repositories
	UserRepo              repository.UserRepository
	AdminUserRepo         repository.AdminUserRepository
	UserTagRepo           repository.UserTagRepository
	RoleRepo              repository.RoleRepository
	PermissionRepo        repository.PermissionRepository
	ModuleRepo            repository.ModuleRepository
	ModulePageRepo        repository.ModulePageRepository
	BannerRepo            repository.BannerRepository
	ArticleRepo           repository.ArticleRepository
	CourseRepo            repository.CourseRepository
	CourseUnitRepo        repository.CourseUnitRepository
	CourseUnitAttachRepo  repository.CourseUnitAttachmentRepository
	FileRepo              repository.FileRepository
	ArticleAttachmentRepo repository.ArticleAttachmentRepository
	CourseAttachmentRepo  repository.CourseAttachmentRepository
	ContentPermissionRepo repository.ContentPermissionRepository
	StudyRecordRepo       repository.StudyRecordRepository
	CollectionRepo        repository.CollectionRepository
	LikeRepo              repository.LikeRepository
	FollowRepo            repository.FollowRepository
	CommentRepo           repository.CommentRepository
	NotificationRepo      repository.NotificationRepository
	WechatConfigRepo      repository.WechatConfigRepository
	AuditLogRepo          repository.AuditLogRepository
	LogConfigRepo         repository.LogConfigRepository
	AttributeRepo         repository.AttributeRepository
	UserAttributeRepo     repository.UserAttributeRepository
	SensitiveWordRepo     repository.SensitiveWordRepository

	// Services
	AuthSvc         service.AuthService
	UserSvc         service.UserService
	RoleSvc         service.RoleService
	PermissionSvc   service.PermissionService
	ModuleSvc       service.ModuleService
	BannerSvc       service.BannerService
	ArticleSvc      service.ArticleService
	CourseSvc       service.CourseService
	StudyRecordSvc  service.StudyRecordService
	CollectionSvc   service.CollectionService
	LikeSvc         service.LikeService
	FollowSvc       service.FollowService
	CommentSvc      service.CommentService
	NotificationSvc service.NotificationService
	WechatConfigSvc service.WechatConfigService
	AuditLogSvc     service.AuditLogService
	LogConfigSvc    service.LogConfigService
	AttributeSvc    service.AttributeService
	UploadFileSvc   service.UploadFileService

	// Controllers
	AuthCtrl         *controller.AuthController
	UserCtrl         *controller.UserController
	RoleCtrl         *controller.RoleController
	PermissionCtrl   *controller.PermissionController
	ModuleCtrl       *controller.ModuleController
	BannerCtrl       *controller.BannerController
	ArticleCtrl      *controller.ArticleController
	CourseCtrl       *controller.CourseController
	StudyRecordCtrl  *controller.StudyRecordController
	CollectionCtrl   *controller.CollectionController
	LikeCtrl         *controller.LikeController
	FollowCtrl       *controller.FollowController
	CommentCtrl      *controller.CommentController
	NotificationCtrl *controller.NotificationController
	SystemCtrl       *controller.SystemController
	UploadCtrl       *controller.UploadController
	DebugCtrl        *controller.DebugController
	AttributeCtrl    *controller.AttributeController
}

// NewProvider initializes all components in order.
func NewProvider(cfg *Config) (*Provider, error) {
	p := &Provider{Config: cfg}

	// 1. Init Logger
	p.Log = initLogger(cfg.Log.Level)

	// 2. Init Database
	db, err := initDatabase(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}
	p.DB = db

	// 3. Init Redis
	p.Redis = initRedis(cfg.Redis)

	// 4. Init Repositories
	p.initRepositories()

	// 5. Init Services
	p.initServices()

	// 6. Init Controllers
	p.initControllers()

	return p, nil
}

func initLogger(level string) *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	log.SetLevel(lvl)
	return log
}

func initDatabase(cfg DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	return db, nil
}

func initRedis(cfg RedisConfig) redis.UniversalClient {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

func (p *Provider) initRepositories() {
	p.UserRepo = repository.NewUserRepository(p.DB)
	p.AdminUserRepo = repository.NewAdminUserRepository(p.DB)
	p.UserTagRepo = repository.NewUserTagRepository(p.DB)
	p.RoleRepo = repository.NewRoleRepository(p.DB)
	p.PermissionRepo = repository.NewPermissionRepository(p.DB)
	p.ModuleRepo = repository.NewModuleRepository(p.DB)
	p.ModulePageRepo = repository.NewModulePageRepository(p.DB)
	p.BannerRepo = repository.NewBannerRepository(p.DB)
	p.ArticleRepo = repository.NewArticleRepository(p.DB)
	p.CourseRepo = repository.NewCourseRepository(p.DB)
	p.CourseUnitRepo = repository.NewCourseUnitRepository(p.DB)
	p.CourseUnitAttachRepo = repository.NewCourseUnitAttachmentRepository(p.DB)
	p.FileRepo = repository.NewFileRepository(p.DB)
	p.ArticleAttachmentRepo = repository.NewArticleAttachmentRepository(p.DB)
	p.CourseAttachmentRepo = repository.NewCourseAttachmentRepository(p.DB)
	p.ContentPermissionRepo = repository.NewContentPermissionRepository(p.DB)
	p.StudyRecordRepo = repository.NewStudyRecordRepository(p.DB)
	p.CollectionRepo = repository.NewCollectionRepository(p.DB)
	p.LikeRepo = repository.NewLikeRepository(p.DB)
	p.FollowRepo = repository.NewFollowRepository(p.DB)
	p.CommentRepo = repository.NewCommentRepository(p.DB)
	p.NotificationRepo = repository.NewNotificationRepository(p.DB)
	p.WechatConfigRepo = repository.NewWechatConfigRepository(p.DB)
	p.AuditLogRepo = repository.NewAuditLogRepository(p.DB)
	p.LogConfigRepo = repository.NewLogConfigRepository(p.DB)
	p.AttributeRepo = repository.NewAttributeRepository(p.DB)
	p.UserAttributeRepo = repository.NewUserAttributeRepository(p.DB)
	p.SensitiveWordRepo = repository.NewSensitiveWordRepository(p.DB)
}

func (p *Provider) initServices() {
	wechatClient := wechat.NewClient(p.Config.Wechat.AppID, p.Config.Wechat.AppSecret)
	var cosClient *cosutil.Client
	if p.Config.Upload.Provider == "cos" && p.Config.Upload.COSEndpoint != "" && p.Config.Upload.COSBucket != "" {
		client, err := cosutil.NewClient(
			p.Config.Upload.COSEndpoint,
			p.Config.Upload.BaseURL,
			p.Config.Upload.COSBucket,
			p.Config.Upload.COSSecretID,
			p.Config.Upload.COSSecretKey,
		)
		if err != nil {
			p.Log.WithError(err).Warn("初始化COS SDK失败")
		} else {
			cosClient = client
			p.UploadFileSvc = service.NewUploadFileService(p.FileRepo, cosClient, p.Log)
		}
	}

	p.AuthSvc = service.NewAuthService(
		p.UserRepo, p.AdminUserRepo, wechatClient,
		p.Config.JWT.Secret, p.Config.JWT.Expiry, p.Log,
	)
	p.UserSvc = service.NewUserService(
		p.UserRepo, p.AdminUserRepo, p.UserTagRepo,
		p.RoleRepo, p.PermissionRepo, p.Log, p.AttributeRepo, p.UserAttributeRepo,
	)
	p.RoleSvc = service.NewRoleService(p.RoleRepo, p.Log)
	p.PermissionSvc = service.NewPermissionService(p.PermissionRepo, p.Log)
	p.ModuleSvc = service.NewModuleService(p.ModuleRepo, p.ModulePageRepo, p.Log)
	p.BannerSvc = service.NewBannerService(p.BannerRepo, p.Log, p.FileRepo, cosClient)
	p.ArticleSvc = service.NewArticleService(
		p.ArticleRepo, p.ContentPermissionRepo, p.Log, p.SensitiveWordRepo, p.ArticleAttachmentRepo, p.RoleRepo, p.FileRepo, cosClient,
	)
	p.CourseSvc = service.NewCourseService(
		p.CourseRepo, p.CourseUnitRepo, p.ContentPermissionRepo, p.Log, p.SensitiveWordRepo, p.CourseAttachmentRepo, p.RoleRepo, p.CourseUnitAttachRepo, p.FileRepo, cosClient,
	)
	p.StudyRecordSvc = service.NewStudyRecordService(p.StudyRecordRepo, p.CourseUnitRepo, p.CourseRepo, p.Log)
	p.CollectionSvc = service.NewCollectionService(p.CollectionRepo, p.ArticleRepo, p.CourseRepo, p.Log)
	p.LikeSvc = service.NewLikeService(
		p.LikeRepo, p.ArticleRepo, p.CourseRepo, p.NotificationRepo, p.Log, p.AttributeRepo, p.UserAttributeRepo,
	)
	p.FollowSvc = service.NewFollowService(
		p.FollowRepo, p.UserRepo, p.NotificationRepo, p.Log, p.AttributeRepo, p.UserAttributeRepo,
	)
	p.CommentSvc = service.NewCommentService(p.CommentRepo, p.ArticleRepo, p.CourseRepo, p.NotificationRepo, p.Log, p.SensitiveWordRepo)
	p.NotificationSvc = service.NewNotificationService(p.NotificationRepo, p.Log)
	p.WechatConfigSvc = service.NewWechatConfigService(p.WechatConfigRepo, p.Log)
	p.AuditLogSvc = service.NewAuditLogService(p.AuditLogRepo, p.Log)
	p.LogConfigSvc = service.NewLogConfigService(p.LogConfigRepo, p.Log)
	p.AttributeSvc = service.NewAttributeService(p.AttributeRepo, p.UserAttributeRepo, p.UserRepo, p.Log)
}

func (p *Provider) initControllers() {
	p.AuthCtrl = controller.NewAuthController(p.AuthSvc, p.Log)
	p.UserCtrl = controller.NewUserController(p.UserSvc, p.Log, p.UploadFileSvc)
	p.RoleCtrl = controller.NewRoleController(p.RoleSvc, p.Log)
	p.PermissionCtrl = controller.NewPermissionController(p.PermissionSvc, p.Log)
	p.ModuleCtrl = controller.NewModuleController(p.ModuleSvc, p.Log, p.ContentPermissionRepo, p.RoleRepo)
	p.BannerCtrl = controller.NewBannerController(p.BannerSvc, p.Log)
	p.ArticleCtrl = controller.NewArticleController(
		p.ArticleSvc, p.Log, p.ContentPermissionRepo, p.RoleRepo, p.ModuleRepo, p.ArticleRepo, p.ArticleAttachmentRepo,
	)
	p.CourseCtrl = controller.NewCourseController(
		p.CourseSvc, p.Log, p.ContentPermissionRepo, p.RoleRepo, p.ModuleRepo, p.CourseRepo, p.CourseUnitRepo, p.CourseAttachmentRepo, p.CourseUnitAttachRepo,
	)
	p.StudyRecordCtrl = controller.NewStudyRecordController(p.StudyRecordSvc, p.Log)
	p.CollectionCtrl = controller.NewCollectionController(p.CollectionSvc, p.Log)
	p.LikeCtrl = controller.NewLikeController(p.LikeSvc, p.Log)
	p.FollowCtrl = controller.NewFollowController(p.FollowSvc, p.Log)
	p.CommentCtrl = controller.NewCommentController(p.CommentSvc, p.Log)
	p.NotificationCtrl = controller.NewNotificationController(p.NotificationSvc, p.Log)
	p.SystemCtrl = controller.NewSystemController(p.WechatConfigSvc, p.AuditLogSvc, p.LogConfigSvc, p.Log)
	if p.Config.Upload.Provider == "cos" && p.Config.Upload.COSEndpoint != "" && p.Config.Upload.COSBucket != "" {
		p.UploadCtrl = controller.NewUploadControllerWithCOS(
			p.Config.Upload.Dir,
			p.Config.Upload.BaseURL,
			p.Config.Upload.COSEndpoint,
			p.Config.Upload.COSBucket,
			p.Log,
		).WithAuditRepo(p.AuditLogRepo).WithUploadService(p.UploadFileSvc).WithPermissionRepos(
			p.ContentPermissionRepo, p.RoleRepo, p.ArticleRepo, p.CourseRepo, p.CourseUnitRepo, p.ArticleAttachmentRepo, p.CourseAttachmentRepo, p.CourseUnitAttachRepo,
		)
	} else {
		p.UploadCtrl = controller.NewUploadController(
			p.Config.Upload.Dir, p.Config.Upload.BaseURL, p.Log,
		).WithAuditRepo(p.AuditLogRepo).WithUploadService(p.UploadFileSvc).WithPermissionRepos(
			p.ContentPermissionRepo, p.RoleRepo, p.ArticleRepo, p.CourseRepo, p.CourseUnitRepo, p.ArticleAttachmentRepo, p.CourseAttachmentRepo, p.CourseUnitAttachRepo,
		)
	}
	p.DebugCtrl = controller.NewDebugController(
		p.UserRepo, p.Config.JWT.Secret, p.Config.JWT.Expiry, p.Log,
	)
	p.AttributeCtrl = controller.NewAttributeController(p.AttributeSvc, p.Log)
}
