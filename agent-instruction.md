# Agent Guide: MiniApp Backend (Gin)

## 项目定位
基于 Gin 框架的微信小程序后端服务，采用 Clean Architecture 四层架构。

## 技术栈（强制使用）
- Web: Gin
- Config: Viper（支持多环境）
- Log: Logrus（JSON 格式）
- ORM: Gorm + MySQL
- Migration: golang-migrate/migrate
- API Spec: OpenAPI 2.0 (Swagger)
- Task Runner: Task（替代 Makefile）
- DI: Google Wire（可选）

## 目录结构规范
project/
├── cmd/server/main.go          # 唯一入口
├── internal/
│   ├── app/provider.go         # 组件初始化（必须按顺序）
│   ├── controller/             # HTTP Handler，仅处理请求响应
│   ├── service/                # 业务逻辑，定义接口
│   ├── repository/             # 数据访问，定义接口
│   ├── model/                  # 领域模型 + DTO
│   │   └── dto/                # Request/Response 结构体
│   ├── middleware/             # Gin 中间件
│   └── pkg/                    # 内部工具包
├── api/swagger.yaml            # OpenAPI 2.0 规范（唯一真相源）
├── migrations/                 # SQL 文件，遵循 migrate 命名规范
├── configs/config.yaml         # 主配置
└── Taskfile.yml                # 命令定义

## Provider 初始化顺序（严禁更改）
1. InitConfig()      // Viper 加载
2. InitLogger()      // Logrus 初始化
3. InitDatabase()    // Gorm 连接池
4. InitMigrate()     // golang-migrate 执行
5. InitRepositories() // 数据层
6. InitServices()    // 业务层
7. InitControllers() // 控制层
8. InitRouter()      // Gin 路由注册

## 编码规范

### 1. Controller 层
- 仅处理：参数绑定 → 调用 Service → 返回响应
- 禁止：直接操作 DB、业务逻辑、外部调用
- 必须：使用统一的 response 包返回 JSON
- 每个 Handler 上方必须添加 Swagger 注释

func (c *UserController) GetUser(ctx *gin.Context) {
    id := ctx.Param("id")
    user, err := c.service.GetByID(ctx, id)
    if err != nil {
        response.Error(ctx, http.StatusInternalServerError, "查询失败", err)
        return
    }
    response.Success(ctx, user)
}

### 2. Service 层
- 必须定义接口，方便 Mock 测试
- 处理业务规则、事务协调
- 禁止：直接操作 SQL，必须通过 Repository

type UserService interface {
    GetByID(ctx context.Context, id string) (*model.User, error)
}

type userService struct {
    repo UserRepository
    log  *logrus.Logger
}

### 3. Repository 层
- 必须定义接口
- 仅处理单表 CRUD，复杂查询用 Gorm Scopes
- 返回：(*Model, error)，查不到返回 (nil, nil) 而非 error

type UserRepository interface {
    GetByID(ctx context.Context, id uint) (*model.User, error)
    Create(ctx context.Context, user *model.User) error
    Update(ctx context.Context, user *model.User) error
    Delete(ctx context.Context, id uint) error
}

### 4. Model 定义
- 领域模型放在 model/entity/
- DTO 放在 model/dto/，按 Request/Response 分组
- Gorm 标签必须包含 comment 说明字段用途

type User struct {
    ID        uint           `gorm:"primarykey;comment:用户ID"`
    OpenID    string         `gorm:"uniqueIndex;size:64;comment:微信OpenID"`
    Nickname  string         `gorm:"size:50;comment:昵称"`
    CreatedAt time.Time      `gorm:"comment:创建时间"`
    UpdatedAt time.Time      `gorm:"comment:更新时间"`
    DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间"`
}

## API 开发流程（必须遵循）

1. 先写 OpenAPI：在 api/swagger.yaml 定义接口
2. 生成/更新文档：task generate-swagger
3. 生成 DTO：根据 swagger 定义创建 dto 结构体
4. 实现 Repository：先写接口，再实现
5. 实现 Service：组合 Repository，编写业务逻辑
6. 实现 Controller：绑定路由，调用 Service
7. 注册路由：在 InitRouter() 中添加
8. 数据库迁移：task migrate-create -- <name> 创建 SQL

## Task 命令规范
所有操作必须通过 Task，禁止手写 Makefile：

task dev              # 热重载开发 (air)
task build            # 构建二进制
task test             # 运行测试
task lint             # 代码检查
task migrate-create   # 创建迁移文件
task migrate-up       # 执行迁移
task migrate-down     # 回滚迁移
task generate         # 生成代码 (Wire + Swagger)

## 配置规范（Viper）
- 配置文件：configs/config.yaml（开发）、config.prod.yaml（生产）
- 环境变量：以 APP_ 为前缀，自动映射（如 APP_SERVER_PORT）
- 结构体标签：mapstructure:"key"

## 日志规范（Logrus）
- 使用 JSON Formatter
- 必须包含 trace_id（从 Gin Context 获取）
- 层级：Debug（开发）、Info（生产）、Error（异常）
- 禁止：使用 fmt.Println，必须用 log.WithField().Info()

## 中间件（必须实现）
- LoggerMiddleware：请求日志（含耗时、状态码）
- RecoveryMiddleware：panic 恢复
- JWTAuthMiddleware：JWT 校验
- CorsMiddleware：跨域处理
- RequestIDMiddleware：生成 trace_id

## 数据库规范
- 迁移文件命名：000001_description.up.sql, 000001_description.down.sql
- 禁止：Gorm AutoMigrate（生产环境），必须用 migrate 工具
- 索引命名：idx_table_column
- 所有表必须有：id, created_at, updated_at, deleted_at

## 错误处理
- 统一错误包：internal/pkg/errors
- 错误码规范：
  - 400: 参数错误
  - 401: 未认证
  - 403: 无权限
  - 404: 资源不存在
  - 500: 服务器错误
- 返回格式：{"code": 500, "message": "错误描述", "data": null}

## 微信小程序特定
- 登录流程：前端 code → 后端调微信 auth.code2Session → 返回自定义 token
- Token 使用 JWT，存储在 Redis（支持多端登录控制）
- 敏感数据解密使用微信提供的算法

## 禁止事项
- 禁止在 Controller 写业务逻辑
- 禁止 Repository 返回具体错误（包装为 domain error）
- 禁止在代码中硬编码配置（必须用 Viper）
- 禁止直接 import 外部包到 Controller/Service（通过接口解耦）
- 禁止修改已发布的 Migration 文件（只能新增）

## Git 提交规范
- feat: 新功能
- fix: 修复
- refactor: 重构
- docs: 文档
- chore: 构建/工具

## 示例指令（给 Agent 的 Prompt 模板）

帮我实现用户模块：
1. 在 api/swagger.yaml 添加 /users 的 CRUD 接口定义
2. 生成对应的 DTO 结构体
3. 创建 UserRepository 接口和实现（Gorm）
4. 创建 UserService 处理业务逻辑
5. 创建 UserController 绑定路由
6. 创建 000002_create_users_table 迁移文件

## 上下文记忆（Agent 必须记住）
- 所有数据库操作必须经过 Repository 接口
- 所有外部调用（微信、OSS）必须封装在 pkg/ 下
- 配置变更后必须重启（Viper 不支持热重载，除非显式实现）
- 开发环境用 task dev，生产用编译后的二进制
