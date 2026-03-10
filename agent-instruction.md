# Agent Guide: MiniApp Backend (Gin)

## 项目定位
基于 Gin 框架的微信小程序后端服务，采用 Clean Architecture 四层架构。

## 技术栈（强制使用）
- Web: Gin
- Config: Viper（支持多环境）
- Log: Logrus（JSON 格式）
- ORM: Gorm + MySQL
- Migration: golang-migrate/migrate（CLI 工具）
- API Spec: OpenAPI 2.0 (Swagger)
- Task Runner: Task（替代 Makefile）
- DI: Google Wire
- Validation: ozzo-validation
- Testing: Ginkgo + Gomega（BDD 风格）
- Mock: mockery（生成接口 mock）

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
├── migrations/                 # SQL 迁移文件（直接放这里）
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   └── 000002_add_users.up.sql
├── configs/config.yaml         # 主配置
├── mocks/                      # mockery 生成的 mock 文件
├── Dockerfile                  # 多阶段构建
├── docker-compose.yml          # 本地编排
└── Taskfile.yml                # 命令定义

## Provider 初始化顺序（严禁更改）
1. InitConfig()      // Viper 加载
2. InitLogger()      // Logrus 初始化
3. InitDatabase()    // Gorm 连接池
4. InitRepositories() // 数据层
5. InitServices()    // 业务层
6. InitControllers() // 控制层
7. InitRouter()      // Gin 路由注册

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

### 5. 参数校验（ozzo-validation）
- DTO 结构体实现 Validate() 方法
- Controller 中显式调用验证
- 错误信息统一格式化

type CreateUserRequest struct {
    Nickname string `json:"nickname"`
    Email    string `json:"email"`
    Age      int    `json:"age"`
}

func (r CreateUserRequest) Validate() error {
    return validation.ValidateStruct(&r,
        validation.Field(&r.Nickname, validation.Required, validation.Length(2, 50)),
        validation.Field(&r.Email, validation.Required, is.Email),
        validation.Field(&r.Age, validation.Required, validation.Min(0), validation.Max(150)),
    )
}

// Controller 中使用
var req CreateUserRequest
if err := ctx.ShouldBindJSON(&req); err != nil {
    response.Error(ctx, http.StatusBadRequest, "参数绑定失败", err)
    return
}
if err := req.Validate(); err != nil {
    response.Error(ctx, http.StatusBadRequest, "参数校验失败", err)
    return
}

## 测试规范（Ginkgo + mockery）

### 1. 测试文件命名与位置
- 单元测试：与被测文件同目录，后缀 _test.go
- 集成测试：tests/integration/ 目录
- Ginkgo 套件：xxx_suite_test.go 启动文件

### 2. mockery 配置（.mockery.yaml）
quiet: false
keeptree: true
mockname: "{{.InterfaceName}}"
filename: "{{.MockName}}.go"
outpkg: mocks
dir: mocks
packages:
  github.com/your/project/internal/service:
    interfaces:
      UserService:
      ArticleService:
  github.com/your/project/internal/repository:
    interfaces:
      UserRepository:
      ArticleRepository:

### 3. 生成 mock
task mock-generate

### 4. Service 层测试示例（Ginkgo 风格）
package service_test

import (
    "context"
    "errors"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "github.com/sirupsen/logrus"

    "project/internal/model"
    "project/internal/service"
    "project/mocks"
)

var _ = Describe("UserService", func() {
    var (
        mockRepo *mocks.UserRepository
        svc      service.UserService
        ctx      context.Context
    )

    BeforeEach(func() {
        mockRepo = new(mocks.UserRepository)
        svc = service.NewUserService(mockRepo, logrus.New())
        ctx = context.Background()
    })

    Describe("GetByID", func() {
        Context("当用户存在时", func() {
            It("应返回用户信息", func() {
                expected := &model.User{ID: 1, Nickname: "test"}
                mockRepo.On("GetByID", ctx, uint(1)).Return(expected, nil)

                user, err := svc.GetByID(ctx, "1")

                Expect(err).To(BeNil())
                Expect(user.ID).To(Equal(uint(1)))
                Expect(user.Nickname).To(Equal("test"))
                mockRepo.AssertExpectations(GinkgoT())
            })
        })

        Context("当用户不存在时", func() {
            It("应返回 nil 且无错误", func() {
                mockRepo.On("GetByID", ctx, uint(999)).Return(nil, nil)

                user, err := svc.GetByID(ctx, "999")

                Expect(err).To(BeNil())
                Expect(user).To(BeNil())
            })
        })

        Context("当数据库出错时", func() {
            It("应返回错误", func() {
                mockRepo.On("GetByID", ctx, uint(1)).Return(nil, errors.New("db error"))

                user, err := svc.GetByID(ctx, "1")

                Expect(err).To(HaveOccurred())
                Expect(user).To(BeNil())
            })
        })
    })

    Describe("Create", func() {
        It("应成功创建用户", func() {
            user := &model.User{Nickname: "new", Email: "test@example.com"}
            mockRepo.On("Create", ctx, user).Return(nil)

            err := svc.Create(ctx, user)

            Expect(err).To(BeNil())
        })
    })
})

### 5. Controller 层测试（使用 Gin Test）
package controller_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"

    "project/internal/controller"
    "project/mocks"
)

var _ = Describe("UserController", func() {
    var (
        mockSvc *mocks.UserService
        ctrl    *controller.UserController
        router  *gin.Engine
        rec     *httptest.ResponseRecorder
    )

    BeforeEach(func() {
        gin.SetMode(gin.TestMode)
        mockSvc = new(mocks.UserService)
        ctrl = controller.NewUserController(mockSvc, logrus.New())
        router = gin.New()
        rec = httptest.NewRecorder()
    })

    Describe("POST /users", func() {
        Context("当参数有效时", func() {
            It("应创建用户并返回 201", func() {
                mockSvc.On("Create", mock.Anything, mock.Anything).Return(nil)

                body, _ := json.Marshal(map[string]string{
                    "nickname": "test",
                    "email":    "test@example.com",
                })
                req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
                req.Header.Set("Content-Type", "application/json")

                router.POST("/users", ctrl.Create)
                router.ServeHTTP(rec, req)

                Expect(rec.Code).To(Equal(http.StatusCreated))
            })
        })

        Context("当参数校验失败时", func() {
            It("应返回 400", func() {
                body, _ := json.Marshal(map[string]string{
                    "nickname": "",
                })
                req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
                req.Header.Set("Content-Type", "application/json")

                router.POST("/users", ctrl.Create)
                router.ServeHTTP(rec, req)

                Expect(rec.Code).To(Equal(http.StatusBadRequest))
            })
        })
    })
})

### 6. 测试命令
task test                 # 运行所有测试
task test-unit            # 仅单元测试（跳过集成）
task test-coverage        # 生成覆盖率报告
task test-watch           # 监听文件变化自动运行

## Docker 规范（多阶段构建）

### 1. Dockerfile（必须遵循）
# 阶段1：构建（使用官方 Go 镜像）
FROM golang:1.22-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 复制依赖文件（利用缓存层）
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 构建二进制（静态链接，无 CGO）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -a -installsuffix cgo \
    -o bin/server \
    ./cmd/server

# 阶段2：运行（使用纯净 Alpine）
FROM alpine:3.19

# 安装基础依赖（ca-certificates 用于 HTTPS，tzdata 用于时区）
RUN apk add --no-cache ca-certificates tzdata

# 创建非 root 用户运行
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从 builder 阶段复制二进制
COPY --from=builder /build/bin/server /app/server

# 复制配置文件和迁移文件（根据实际需要）
COPY --from=builder /build/configs /app/configs
COPY --from=builder /build/migrations /app/migrations

# 设置权限
RUN chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 运行
ENTRYPOINT ["/app/server"]

### 2. docker-compose.yml（本地开发）
version: "3.8"

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=development
      - APP_SERVER_PORT=8080
      - APP_DATABASE_HOST=mysql
      - APP_DATABASE_PORT=3306
      - APP_DATABASE_USER=root
      - APP_DATABASE_PASSWORD=secret
      - APP_DATABASE_NAME=miniapp
      - APP_LOG_LEVEL=debug
    volumes:
      - ./configs:/app/configs:ro
      - ./storage/logs:/app/storage/logs
    depends_on:
      mysql:
        condition: service_healthy
    networks:
      - app-network

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: secret
      MYSQL_DATABASE: miniapp
    ports:
      - "3306:3306"
    volumes:
      - mysql-data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 5s
      timeout: 3s
      retries: 5
    networks:
      - app-network

  migrate:
    image: migrate/migrate:v4.17.0
    volumes:
      - ./migrations:/migrations:ro
    environment:
      - MIGRATE_DATABASE_URL=mysql://root:secret@tcp(mysql:3306)/miniapp?multiStatements=true
    command: ["-path", "/migrations", "-database", "${MIGRATE_DATABASE_URL}", "up"]
    depends_on:
      mysql:
        condition: service_healthy
    networks:
      - app-network

volumes:
  mysql-data:

networks:
  app-network:
    driver: bridge

### 3. 迁移策略（二选一）

#### 方案A：独立 migrate 容器（推荐，已包含在 docker-compose.yml）
- 使用官方镜像 migrate/migrate:v4.17.0
- SQL 文件直接放在项目根目录 migrations/ 下（无需子目录）
- 命名规范：000001_description.up.sql / 000001_description.down.sql
- 启动顺序：mysql healthy → migrate 执行 → app 启动

#### 方案B：应用内迁移（Provider.InitMigrate）
- 在 provider.go 中使用 golang-migrate/migrate 库
- 启动时自动执行（配置 database.auto_migrate: true）
- 无需独立容器，但需确保应用有数据库写权限

### 4. 迁移文件命名规范
- 格式：{版本号}_{描述}.{方向}.sql
- 版本号：6位数字，递增（000001, 000002...）
- 方向：up（升级）或 down（回滚）
- 示例：
  - 000001_init_schema.up.sql
  - 000001_init_schema.down.sql
  - 000002_add_users.up.sql
  - 000002_add_users.down.sql

### 5. Docker 相关 Task 命令
task docker-build         # 构建镜像
task docker-run           # 运行容器
task docker-compose-up    # 启动完整环境（含 MySQL + migrate）
task docker-compose-down  # 停止环境
task docker-push          # 推送到仓库
task migrate-local        # 本地执行迁移（使用 migrate CLI）

## API 开发流程（必须遵循）

1. 先写 OpenAPI：在 api/swagger.yaml 定义接口
2. 生成/更新文档：task generate-swagger
3. 生成 DTO：根据 swagger 定义创建 dto 结构体，实现 Validate()
4. 实现 Repository：先写接口，再实现
5. 实现 Service：组合 Repository，编写业务逻辑
6. 实现 Controller：绑定路由，调用 Service，调用 Validate()
7. 注册路由：在 InitRouter() 中添加
8. 数据库迁移：task migrate-create -- <name> 创建 SQL
9. 生成 mock：task mock-generate
10. 编写测试：Ginkgo BDD 风格，覆盖成功/失败场景
11. 构建镜像：task docker-build 验证打包

## Task 命令规范
所有操作必须通过 Task，禁止手写 Makefile：

task dev                  # 热重载开发 (air)
task build                # 构建二进制
task test                 # 运行所有测试（Ginkgo）
task test-unit            # 仅单元测试
task test-coverage        # 覆盖率报告
task test-watch           # 监听模式
task lint                 # 代码检查
task mock-generate        # 生成 mock 文件
task migrate-create       # 创建迁移文件（使用 migrate CLI）
task migrate-local        # 本地执行迁移
task migrate-up           # Docker 环境执行迁移
task migrate-down         # Docker 环境回滚迁移
task generate             # 生成代码 (Wire + Swagger + mockery)
task docker-build         # Docker 多阶段构建
task docker-run           # 运行容器
task docker-compose-up    # 启动完整环境（含 migrate 容器）
task docker-compose-down  # 停止环境

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
- 迁移文件位置：项目根目录 migrations/ 下（无需子目录）
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
- 禁止在 Controller 直接使用 gin 的 binding 验证器（必须用 ozzo-validation）
- 禁止在单元测试中连接真实数据库（必须用 mock）
- 禁止在测试中使用 testify/suite（必须用 Ginkgo）
- 禁止单阶段 Dockerfile（必须用多阶段构建）
- 禁止在最终镜像中保留 Go 编译器（必须 copy 到 alpine）
- 禁止以 root 用户运行容器（必须创建 appuser）
- 禁止为 migrations 创建独立 Dockerfile（直接使用官方 migrate/migrate 镜像或应用内迁移）

## Git 提交规范
- feat: 新功能
- fix: 修复
- refactor: 重构
- docs: 文档
- chore: 构建/工具
- test: 测试相关

## 示例指令（给 Agent 的 Prompt 模板）

帮我实现用户模块：
1. 在 api/swagger.yaml 添加 /users 的 CRUD 接口定义
2. 生成对应的 DTO 结构体（实现 ozzo-validation 的 Validate 方法）
3. 创建 UserRepository 接口和实现（Gorm）
4. 创建 UserService 接口和实现，编写业务逻辑
5. 创建 UserController 绑定路由（调用 DTO.Validate()）
6. 创建 000002_add_users_table 迁移文件（放在 migrations/ 根目录）
7. 运行 task mock-generate 生成 mock
8. 创建 service/user_service_test.go 编写 Ginkgo BDD 测试（覆盖正常/异常场景）
9. 创建 controller/user_controller_test.go 编写 HTTP 层测试
10. 验证 task docker-build 成功构建多阶段镜像
11. 验证 task docker-compose-up 能自动执行迁移并启动服务

## 上下文记忆（Agent 必须记住）
- 所有数据库操作必须经过 Repository 接口
- 所有外部调用（微信、OSS）必须封装在 pkg/ 下
- 所有请求参数必须通过 ozzo-validation 校验
- 所有接口必须定义在 Service/Repository 层用于 mock
- 所有测试使用 Ginkgo + Gomega 风格，mock 用 mockery 生成
- 所有构建必须通过多阶段 Dockerfile（golang:alpine → alpine）
- 迁移文件直接放在 migrations/ 根目录（无需子目录，无需独立 Dockerfile）
- Docker 环境使用官方 migrate/migrate 镜像执行迁移
- 配置变更后必须重启（Viper 不支持热重载，除非显式实现）
- 开发环境用 task dev，生产用编译后的二进制或 Docker 镜像
