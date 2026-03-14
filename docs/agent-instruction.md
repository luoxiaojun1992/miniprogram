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
- Lint: golangci-lint（包含 gofmt, goimports, vet, staticcheck 等）
- Format: gofmt（标准格式化）

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
│       ├── errors/             # 统一错误处理（新增）
│       └── response/           # 响应封装
├── api/
│   ├── api/docs/swagger.yaml   # OpenAPI 2.0 规范（唯一真相源）
│   └── gen/                    # OpenAPI 生成的代码文件
├── docs/
│   └── docs/agent-instruction.md # 本文件
├── migrations/                 # SQL 迁移文件（直接放这里）
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   └── 000002_add_users.up.sql
├── configs/config.yaml         # 主配置
├── mocks/                      # mockery 生成的 mock 文件
├── .golangci.yml               # lint 配置（新增）
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
- 仅处理：参数绑定 → 调用 Service → 返回响应（或返回错误）
- 禁止：直接操作 DB、业务逻辑、外部调用
- **必须：返回 error，由 ErrorMiddleware 统一处理响应**
- 每个 Handler 上方必须添加 Swagger 注释

func (c *UserController) GetUser(ctx *gin.Context) {
   id := ctx.Param("id")
   user, err := c.service.GetByID(ctx, id)
   if err != nil {
       ctx.Error(err)  // 交给 ErrorMiddleware 处理
       return
   }
   response.Success(ctx, user)
}

func (c *UserController) CreateUser(ctx *gin.Context) {
   var req dto.CreateUserRequest
   if err := ctx.ShouldBindJSON(&req); err != nil {
       ctx.Error(errors.NewBadRequest("参数绑定失败", err))
       return
   }
   if err := req.Validate(); err != nil {
       ctx.Error(errors.NewValidation("参数校验失败", err))
       return
   }
   
   user, err := c.service.Create(ctx, &req)
   if err != nil {
       ctx.Error(err)
       return
   }
   response.Success(ctx, http.StatusCreated, user)
}

### 2. Service 层
- 必须定义接口，方便 Mock 测试
- 处理业务规则、事务协调
- 禁止：直接操作 SQL，必须通过 Repository
- **返回：使用 errors.NewXxx() 包装 domain error**

type UserService interface {
   GetByID(ctx context.Context, id string) (*model.User, error)
   Create(ctx context.Context, req *dto.CreateUserRequest) (*model.User, error)
}

type userService struct {
   repo UserRepository
   log  *logrus.Logger
}

func (s *userService) GetByID(ctx context.Context, id string) (*model.User, error) {
   user, err := s.repo.GetByID(ctx, id)
   if err != nil {
       return nil, errors.NewInternal("数据库查询失败", err)
   }
   if user == nil {
       return nil, errors.NewNotFound("用户不存在", nil)
   }
   return user, nil
}

### 3. Repository 层
- 必须定义接口
- 仅处理单表 CRUD，复杂查询用 Gorm Scopes
- 返回：(*Model, error)，查不到返回 (nil, nil) 而非 error
- **禁止返回具体错误类型，统一返回 errors.NewInternal()**

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
   ctx.Error(errors.NewBadRequest("参数绑定失败", err))
   return
}
if err := req.Validate(); err != nil {
   ctx.Error(errors.NewValidation("参数校验失败", err))
   return
}

### 6. 统一错误处理（errors 包）

#### 错误结构体定义
package errors

type AppError struct {
   Code    int    `json:"code"`     // 业务错误码
   Message string `json:"message"`  // 用户可读消息
   HTTPCode int   `json:"-"`        // HTTP 状态码（不序列化）
   Cause   error  `json:"-"`        // 原始错误（日志用，不暴露）
}

func (e *AppError) Error() string {
   return e.Message
}

// 预定义错误类型
func NewBadRequest(message string, cause error) *AppError {
   return &AppError{Code: 400001, Message: message, HTTPCode: 400, Cause: cause}
}

func NewValidation(message string, cause error) *AppError {
   return &AppError{Code: 400002, Message: message, HTTPCode: 400, Cause: cause}
}

func NewUnauthorized(message string, cause error) *AppError {
   return &AppError{Code: 401001, Message: message, HTTPCode: 401, Cause: cause}
}

func NewForbidden(message string, cause error) *AppError {
   return &AppError{Code: 403001, Message: message, HTTPCode: 403, Cause: cause}
}

func NewNotFound(message string, cause error) *AppError {
   return &AppError{Code: 404001, Message: message, HTTPCode: 404, Cause: cause}
}

func NewInternal(message string, cause error) *AppError {
   return &AppError{Code: 500001, Message: message, HTTPCode: 500, Cause: cause}
}

// 错误转换工具
func ToResponse(err error) (int, map[string]interface{}) {
   if appErr, ok := err.(*AppError); ok {
       return appErr.HTTPCode, map[string]interface{}{
           "code":    appErr.Code,
           "message": appErr.Message,
           "data":    nil,
       }
   }
   // 未知错误类型，统一包装为 500
   return http.StatusInternalServerError, map[string]interface{}{
       "code":    500000,
       "message": "服务器内部错误",
       "data":    nil,
   }
}

#### ErrorMiddleware 实现
package middleware

import (
   "github.com/gin-gonic/gin"
   "project/internal/pkg/errors"
   "github.com/sirupsen/logrus"
)

func ErrorMiddleware(log *logrus.Logger) gin.HandlerFunc {
   return func(ctx *gin.Context) {
       ctx.Next() // 先执行后续 handler
       
       // 检查是否有错误
       if len(ctx.Errors) > 0 {
           err := ctx.Errors.Last().Err
           
           // 记录日志（包含原始错误）
           if appErr, ok := err.(*errors.AppError); ok && appErr.Cause != nil {
               log.WithError(appErr.Cause).WithField("code", appErr.Code).Error(appErr.Message)
           } else {
               log.WithError(err).Error("request error")
           }
           
           // 统一转换为 response
           httpCode, resp := errors.ToResponse(err)
           ctx.JSON(httpCode, resp)
           ctx.Abort()
       }
   }
}

#### response 包更新
package response

import (
   "net/http"
   "github.com/gin-gonic/gin"
)

// Success 统一成功响应
func Success(ctx *gin.Context, data interface{}) {
   ctx.JSON(http.StatusOK, gin.H{
       "code":    0,
       "message": "success",
       "data":    data,
   })
}

// SuccessWithStatus 指定状态码的成功响应
func SuccessWithStatus(ctx *gin.Context, status int, data interface{}) {
   ctx.JSON(status, gin.H{
       "code":    0,
       "message": "success",
       "data":    data,
   })
}

### 7. 代码规范与 Lint（强制）

#### .golangci.yml 配置
run:
 timeout: 5m
 issues-exit-code: 1
 tests: true

output:
 format: colored-line-number
 print-issued-lines: true
 print-linter-name: true

linters-settings:
 gofmt:
   simplify: true
 goimports:
   local-prefixes: github.com/your/project
 govet:
   check-shadowing: true
   enable-all: true
 staticcheck:
   checks: ["all"]
 errcheck:
   check-type-assertions: true
   check-blank: true
 gocyclo:
   min-complexity: 15
 misspell:
   locale: US

linters:
 enable:
   - gofmt        # 标准格式化
   - goimports    # 自动导入管理
   - govet        # 标准 vet 工具
   - staticcheck  # 静态分析
   - errcheck     # 检查未处理错误
   - gosimple     # 简化代码建议
   - ineffassign  # 无效赋值检查
   - typecheck    # 类型检查
   - unused       # 未使用代码
   - misspell     # 拼写检查
   - gocyclo      # 圈复杂度
 disable:
   - deadcode     # 已弃用，由 unused 替代
   - structcheck  # 已弃用

issues:
 exclude-use-default: false
 max-issues-per-linter: 50
 max-same-issues: 3

#### Lint 相关 Task 命令
task lint              # 运行 golangci-lint 检查
task lint-fix          # 自动修复（gofmt, goimports）
task format            # 仅运行 gofmt 格式化

#### 编码规范检查清单
- [ ] 运行 `task lint` 无错误
- [ ] 所有文件通过 `gofmt -s` 简化格式化
- [ ] 导入分组正确：标准库 → 第三方 → 项目内部（goimports）
- [ ] 无未处理的 error（errcheck）
- [ ] 无无效赋值（ineffassign）
- [ ] 函数圈复杂度 < 15（gocyclo）
- [ ] 无拼写错误（misspell）

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
   "project/internal/pkg/errors"
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
           It("应返回 NotFound 错误", func() {
               mockRepo.On("GetByID", ctx, uint(999)).Return(nil, nil)

               user, err := svc.GetByID(ctx, "999")

               Expect(err).To(HaveOccurred())
               Expect(err.(*errors.AppError).Code).To(Equal(404001))
               Expect(user).To(BeNil())
           })
       })

       Context("当数据库出错时", func() {
           It("应返回 Internal 错误", func() {
               mockRepo.On("GetByID", ctx, uint(1)).Return(nil, errors.New("db error"))

               user, err := svc.GetByID(ctx, "1")

               Expect(err).To(HaveOccurred())
               Expect(err.(*errors.AppError).HTTPCode).To(Equal(500))
               Expect(user).To(BeNil())
           })
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
   "project/internal/middleware"
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
       // 注册 ErrorMiddleware 用于测试
       router.Use(middleware.ErrorMiddleware(logrus.New()))
       rec = httptest.NewRecorder()
   })

   Describe("POST /users", func() {
       Context("当参数有效时", func() {
           It("应创建用户并返回 201", func() {
               mockSvc.On("Create", mock.Anything, mock.Anything).Return(&model.User{ID: 1}, nil)

               body, _ := json.Marshal(map[string]string{
                   "nickname": "test",
                   "email":    "test@example.com",
               })
               req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
               req.Header.Set("Content-Type", "application/json")

               router.POST("/users", ctrl.CreateUser)
               router.ServeHTTP(rec, req)

               Expect(rec.Code).To(Equal(http.StatusCreated))
           })
       })

       Context("当参数校验失败时", func() {
           It("应返回 400 和错误码", func() {
               body, _ := json.Marshal(map[string]string{
                   "nickname": "",
               })
               req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
               req.Header.Set("Content-Type", "application/json")

               router.POST("/users", ctrl.CreateUser)
               router.ServeHTTP(rec, req)

               Expect(rec.Code).To(Equal(http.StatusBadRequest))
               // 验证错误码格式
               var resp map[string]interface{}
               json.Unmarshal(rec.Body.Bytes(), &resp)
               Expect(resp["code"]).To(BeNumerically(">=", 400000))
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

1. 先写 OpenAPI：在 api/docs/swagger.yaml 定义接口
2. 生成/更新文档：task generate-swagger
3. 生成 DTO：根据 swagger 定义创建 dto 结构体，实现 Validate()
4. 实现 Repository：先写接口，再实现
5. 实现 Service：组合 Repository，编写业务逻辑
6. 实现 Controller：绑定路由，调用 Service，调用 Validate()
7. 注册路由：在 InitRouter() 中添加
8. 数据库迁移：task migrate-create -- <name> 创建 SQL
9. 生成 mock：task mock-generate
10. 编写测试：Ginkgo BDD 风格，覆盖成功/失败场景
11. **运行 lint：task lint 确保无代码规范问题**
12. 构建镜像：task docker-build 验证打包

## Task 命令规范
所有操作必须通过 Task，禁止手写 Makefile：

task dev                  # 热重载开发 (air)
task build                # 构建二进制
task test                 # 运行所有测试（Ginkgo）
task test-unit            # 仅单元测试
task test-coverage        # 覆盖率报告
task test-watch           # 监听模式
task lint                 # 运行 golangci-lint 检查
task lint-fix             # 自动修复代码规范（gofmt, goimports）
task format               # 仅运行 gofmt 格式化
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
- **ErrorMiddleware：统一错误转换（必须在其他中间件之前注册）**
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

## 错误处理（统一错误体系）
- **统一错误包：internal/pkg/errors（必须按此规范实现）**
- **错误结构：{code, message, http_code}，其中 code 为业务码，http_code 为 HTTP 状态码**
- **错误传递：Controller/Service/Repository 统一返回 error，由 ErrorMiddleware 统一转换**
- **错误码规范：**
 - 400xxx: 参数错误（400001=通用参数错误，400002=校验失败）
 - 401xxx: 未认证（401001=Token无效）
 - 403xxx: 无权限（403001=禁止访问）
 - 404xxx: 资源不存在（404001=记录不存在）
 - 500xxx: 服务器错误（500001=内部错误）
- **错误创建：使用 errors.NewXxx() 系列函数，禁止直接构造 AppError**
- **错误响应：统一格式 {"code": 400001, "message": "错误描述", "data": null}**

## 微信小程序特定
- 登录流程：前端 code → 后端调微信 auth.code2Session → 返回自定义 token
- Token 使用 JWT，存储在 Redis（支持多端登录控制）
- 敏感数据解密使用微信提供的算法

## 禁止事项
- 禁止在 Controller 写业务逻辑
- **禁止 Repository/Service 返回具体错误类型（必须包装为 *errors.AppError）**
- 禁止在代码中硬编码配置（必须用 Viper）
- 禁止直接 import 外部包到 Controller/Service（通过接口解耦）
- 禁止修改已发布的 Migration 文件（只能新增）
- 禁止在 Controller 直接使用 gin 的 binding 验证器（必须用 ozzo-validation）
- **禁止在 Controller 中直接调用 response.Error()（必须 ctx.Error() 抛出）**
- 禁止在单元测试中连接真实数据库（必须用 mock）
- 禁止在测试中使用 testify/suite（必须用 Ginkgo）
- 禁止单阶段 Dockerfile（必须用多阶段构建）
- 禁止在最终镜像中保留 Go 编译器（必须 copy 到 alpine）
- 禁止以 root 用户运行容器（必须创建 appuser）
- 禁止为 migrations 创建独立 Dockerfile（直接使用官方 migrate/migrate 镜像或应用内迁移）
- **禁止提交未通过 lint 检查的代码（CI 必须配置 lint 步骤）**

## Git 提交规范
- feat: 新功能
- fix: 修复
- refactor: 重构
- docs: 文档
- chore: 构建/工具
- test: 测试相关

## 示例指令（给 Agent 的 Prompt 模板）

帮我实现用户模块：
1. 在 api/docs/swagger.yaml 添加 /users 的 CRUD 接口定义
2. 生成对应的 DTO 结构体（实现 ozzo-validation 的 Validate 方法）
3. 创建 UserRepository 接口和实现（Gorm）
4. 创建 UserService 接口和实现，使用 errors.NewInternal/NewNotFound 包装错误
5. 创建 UserController 绑定路由，使用 ctx.Error() 传递错误
6. 创建 000002_add_users_table 迁移文件（放在 migrations/ 根目录）
7. 运行 task mock-generate 生成 mock
8. 创建 service/user_service_test.go 编写 Ginkgo BDD 测试（验证错误类型）
9. 创建 controller/user_controller_test.go 编写 HTTP 层测试（验证错误码）
10. **运行 task lint 确保代码规范通过**
11. 验证 task docker-build 成功构建多阶段镜像
12. 验证 task docker-compose-up 能自动执行迁移并启动服务

## 上下文记忆（Agent 必须记住）
- 所有数据库操作必须经过 Repository 接口
- 所有外部调用（微信、OSS）必须封装在 pkg/ 下
- 所有请求参数必须通过 ozzo-validation 校验
- 所有接口必须定义在 Service/Repository 层用于 mock
- **所有错误统一使用 internal/pkg/errors 包创建，由 ErrorMiddleware 统一处理**
- 所有测试使用 Ginkgo + Gomega 风格，mock 用 mockery 生成
- 所有构建必须通过多阶段 Dockerfile（golang:alpine → alpine）
- **所有代码必须通过 golangci-lint 检查（包含 gofmt, goimports 等）**
- 迁移文件直接放在 migrations/ 根目录（无需子目录，无需独立 Dockerfile）
- Docker 环境使用官方 migrate/migrate 镜像执行迁移
- 配置变更后必须重启（Viper 不支持热重载，除非显式实现）
- 开发环境用 task dev，生产用编译后的二进制或 Docker 镜像

## 业务流程规范（基于 OpenAPI 2.0 定义）

### 1. 认证模块业务流程

#### 1.1 微信登录（/auth/wechat-login）
1. 小程序前端调用 wx.login() 获取临时登录凭证 code
2. 前端将 code 发送到后端 /auth/wechat-login
3. 后端用 code + appid + secret 请求微信接口获取 openid
4. 查询数据库：openid 是否存在？
   - 存在 → 直接生成 JWT Token，返回用户信息
   - 不存在 → 创建新用户（user_type=1前台用户）→ 生成 JWT Token，返回用户信息
5. 可选：如传 encrypted_data 和 iv，解密获取手机号绑定

#### 1.2 管理员登录（/auth/admin-login）
1. 前端获取图形验证码（需单独接口）
2. 用户输入邮箱、密码、验证码
3. 后端验证：
   - 验证码是否正确（Redis/Session 比对）
   - 邮箱是否存在且 user_type ∈ {2,3}（管理员）
   - 密码是否正确（bcrypt 比对）
4. 验证通过 → 生成 JWT Token，更新 last_login_at
5. 记录审计日志（admin:login）

#### 1.3 Token 刷新（/auth/refresh）
1. 前端在 Token 即将过期时调用（需携带有效 Token）
2. 后端解析 Token 获取 user_id
3. 查询用户状态（是否被冻结/删除）
4. 生成新的 Access Token（延长有效期）
5. 返回新 Token（Refresh Token 机制可在此扩展）

### 2. 用户模块业务流程

#### 2.1 获取/更新当前用户信息（/users/profile）
- 获取流程：
  1. 从 JWT Token 解析 user_id
  2. 查询 users 表获取基础信息
  3. 查询 user_tags 表获取用户标签列表
  4. 组装 UserInfo 返回（openid 仅自己可见）
- 更新流程：
  1. 验证 JWT Token
  2. 校验参数：nickname（长度、敏感词过滤）、avatar_url（格式校验、域名白名单）
  3. 更新 users 表对应字段
  4. 返回成功响应

#### 2.2 获取用户权限列表（/users/permissions）
1. 从 JWT 获取 user_id
2. 查询 user_roles 表获取用户所有 role_id
3. 查询 roles 表获取角色编码列表
4. 查询 role_permissions 表获取所有 permission_id
5. 去重后查询 permissions 表获取权限编码列表
6. 返回 {roles: [...], permissions: [...]}
7. 【优化】：权限可缓存到 Redis，用户登录时写入，变更时清除

#### 2.3 管理员用户管理（/admin/users）
- 列表查询：
  1. 权限校验：检查当前用户是否有 "user:view" 权限
  2. 参数处理：page/page_size/keyword/user_type/status
  3. 构建动态 SQL 查询 users 表
  4. 联查 user_tags（聚合为数组）
  5. 分页返回，敏感字段（openid）脱敏
- 创建管理员：
  1. 权限校验："user:edit"
  2. 校验：email 唯一性、password 强度、user_type 合法性
  3. bcrypt 加密密码
  4. 插入 users 表（user_type=2或3）
  5. 记录审计日志（user:create）
  6. 返回 201 + Location 头
- 更新/删除/分配角色/标签管理关键检查：
  - 不能修改自己的 user_type（防止降权后无法操作）
  - 删除前检查：是否有创建的内容、是否为唯一系统管理员
  - 标签操作：去重、限制单个用户标签数量（如最多10个）

### 3. 权限管理（RBAC）业务流程

#### 3.1 角色管理（/admin/roles）
- 创建角色：
  1. 权限校验："role:edit"
  2. 校验：name 唯一性、parent_id 是否存在（层级校验）
  3. 计算 level：parent_id=0 则为1，否则父角色 level+1
  4. 检查层级深度限制（如最多5层）
  5. 插入 roles 表
  6. 如传 permission_ids，批量插入 role_permissions
  7. 清除角色权限缓存
- 删除角色：
  1. 检查 is_builtin=1 → 返回 403 禁止删除
  2. 检查是否有子角色 → 如有，需先删除子角色或转移
  3. 检查是否已分配给用户 → 如有，提示先解除关联
  4. 事务删除：role_permissions → user_roles → roles
  5. 清除相关用户权限缓存

#### 3.2 权限树获取（/admin/permissions）
1. 查询所有 permissions（type=1菜单,2按钮,3接口）
2. 按 parent_id 构建树形结构（递归或层级遍历）
3. 返回树形 JSON，用于前端菜单渲染和权限配置界面
4. 【数据初始化】：系统启动时检查内置权限是否完整，缺失自动插入

### 4. 内容管理 - 模块业务流程

#### 4.1 模块管理（/modules, /admin/modules）
- 列表查询（前台）：只返回 status=1 的模块，按 sort_order 排序，可用于小程序首页模块导航
- 创建/更新（管理端）：
  1. 权限："module:management"
  2. 检查 title 唯一性
  3. 更新时检查是否有关联文章/课程（影响删除决策）
- 模块页面管理（/admin/modules/{id}/pages）：
  1. 支持富文本/HTML 内容存储
  2. 应用场景：模块介绍页、帮助文档等静态页面
  3. 内容安全：XSS 过滤（HTML 类型需严格白名单标签）

### 5. 文章管理业务流程

#### 5.1 前台文章浏览
- 文章列表（/articles）：
  1. 参数处理：分页、keyword（标题/摘要模糊搜索）、module_id（筛选）、sort（排序规则）
  2. 基础过滤：status=1（已发布）且 publish_time <= now()
  3. 权限过滤：如文章有 role_permissions，检查当前用户角色
     - 无登录 → 只返回公开文章
     - 有登录 → 返回公开 + 用户角色匹配的文章
  4. 查询 users 表获取作者信息
  5. 如已登录，联查 likes/collections 表标记 is_liked 等
  6. 返回列表（不含 content 全文，减轻传输）
- 文章详情（/articles/{id}）：
  1. 查询文章，检查存在性和状态
  2. 权限校验（同列表逻辑）
  3. 异步/延迟：view_count +1（避免阻塞，可用 Redis 累加）
  4. 如已登录，查询 like/collection 状态
  5. 返回完整内容（含 permissions 配置）

#### 5.2 管理端文章管理（/admin/articles）
- 创建文章：
  1. 权限校验："article:create"
  2. 参数校验：
     - title：必填，长度 1-200，敏感词检测
     - content：必填，XSS 过滤（根据 content_type）
     - cover_image：URL 格式校验
     - status：0草稿/1立即发布/2定时发布
     - publish_time：status=2 时必填且必须未来时间
  3. 处理 role_permissions：空数组 → 公开文章；指定角色 → 插入 article_permissions 表
  4. 插入 articles 表，获取 id
  5. 如 status=1，记录发布日志
  6. 返回 201 + Location: /articles/{id}
- 发布/取消发布（/admin/articles/{id}/publish）：
  1. 权限："article:publish"
  2. 检查文章存在性
  3. status=1（发布）：检查 content 非空、设置 publish_time=now
  4. status=0（草稿）：直接更新状态
  5. 发送通知：如文章从草稿变为发布，通知收藏该模块的用户（可选）
- 【定时发布实现方案】：
  - 方案A：定时任务（每分钟扫描 status=2 and publish_time<=now）
  - 方案B：延迟队列（Redis ZSET / RabbitMQ 延迟消息）
  - 建议：简单实现用方案A，大规模用方案B

### 6. 课程管理业务流程

#### 6.1 课程结构与购买逻辑
- 课程单元管理（/admin/courses/{id}/units）：
  1. 单元是课程的子资源，一个课程含多个单元（视频课）
  2. 创建单元：上传视频 → 获取 duration（ffmpeg 解析或前端传）
  3. 课程总时长 = SUM(units.duration)
  4. 单元排序：sort_order 字段，支持拖拽排序
  5. 删除单元：同步更新 courses.duration
- 课程学习权限检查（/courses/{id}）：
  1. 检查课程状态（已发布）
  2. 检查价格：
     - price=0 → 允许访问
     - price>0 → 检查用户是否已购买（orders 表查询）
       - 已购买 → 返回完整信息 + units
       - 未购买 → 返回基础信息 + 第一个单元预览
  3. 如已购买，查询 study_records 返回学习进度

### 7. 互动功能业务流程

#### 7.1 学习记录（/study-records）
1. 前端定时上报（如每30秒）：当前观看 unit_id + progress(秒)
2. 后端：INSERT INTO study_records ... ON DUPLICATE KEY UPDATE progress=VALUES(progress), status=..., last_study_at=NOW()
3. 状态计算：
   - progress < duration-10 → status=1（学习中）
   - progress >= duration-10 → status=2（已完成）
4. 课程学习人数统计：COUNT(DISTINCT user_id WHERE status>=1)
5. 【防刷】：progress 增长幅度合理性校验（如单次不超过60秒）

#### 7.2 收藏/点赞（/collections/{type}/{id}, /likes/{type}/{id}）
- 收藏流程：
  1. 幂等性检查：SELECT * FROM collections WHERE user_id=? AND content_type=? AND content_id=?
     - 存在 → 返回 409 Conflict（已收藏）
     - 不存在 → 继续
  2. 检查内容存在性（articles/courses 表）
  3. INSERT collections
  4. 异步：UPDATE articles/courses SET collect_count=collect_count+1
  5. 返回 201
- 点赞流程（类似收藏，但支持取消后重新点赞）：
  - 取消点赞：DELETE likes 记录，like_count-1
  - 重新点赞：INSERT 新记录（允许重复操作，不报错）

#### 7.3 评论系统（/comments/{type}/{id}, /admin/comments）
- 发表评论：
  1. 内容校验：长度 1-1000、敏感词过滤、防SQL注入
  2. 检查 parent_id：如不为0，检查父评论是否存在且属于同一内容
  3. 插入 comments（status=0 待审核 / 1 通过）
     【审核策略】：
     - 用户等级高/历史记录好 → 自动通过（status=1）
     - 含敏感词/新用户 → 待审核（status=0）
  4. 如审核通过，通知被回复用户（如 parent_id≠0）
- 评论审核（/admin/comments/{id}/audit）：
  1. 权限："comment:audit"
  2. status=1（通过）：更新状态，发送通知给评论者
  3. status=2（拒绝）：更新状态，可选发送拒绝原因
  4. 前端展示：只查询 status=1 的评论

### 8. 消息通知业务流程（/notifications）

#### 8.1 通知生成与拉取
- 通知触发场景：
  1. 系统通知：管理员后台群发 → 写入所有用户 notifications
  2. 评论回复：用户A回复用户B → 给B写入 type=2 通知
  3. 学习提醒：定时任务扫描 study_records WHERE last_study_at < DATE_SUB(NOW(), INTERVAL 3 DAY) → 发送学习提醒通知
- 通知列表：
  1. 查询当前用户通知，支持 is_read 筛选
  2. 返回 unread_count（未读总数，用于小红点）
  3. 标记已读：可单条（PUT /{id}/read）或全部（PUT /read-all）
  4. 【性能优化】：未读数可缓存到 Redis，变更时更新

### 9. 系统管理业务流程

#### 9.1 微信配置管理（/admin/wechat-config）
1. 权限："wechat:config"（仅系统管理员）
2. 存储字段：app_id, app_secret（加密存储）, api_token
3. 自动刷新机制：
   - access_token：缓存7200秒，过期前自动刷新
   - jsapi_ticket：缓存7200秒，用于前端 SDK 签名
4. 安全：app_secret 返回时脱敏（只显示后4位）
5. 变更后：清除 Redis 中的 token 缓存，强制重新获取

#### 9.2 审计日志（/admin/audit-logs）
- 自动记录场景（AOP/中间件实现）：
  - 管理员登录/登出
  - 内容创建/修改/删除（记录变更前后数据）
  - 用户冻结/角色变更
  - 配置修改
- 日志清理：
  - 定时任务：每天删除 retention_days 之前的日志
  - 或归档到冷存储（OSS/S3）后删除

### 10. 文件上传业务流程（/upload/image, /upload/video）

#### 10.1 图片上传
1. 验证：文件大小（<5MB）、MIME类型（image/*）、文件头魔数（防伪装）
2. 根据 type 参数决定存储路径：
   - avatar：用户头像（压缩至 200x200）
   - article/course：内容图片（压缩至最大宽度1200）
   - cover：封面图（压缩至 750x400）
3. 生成文件名：{type}/{date}/{uuid}.{ext}
4. 上传至对象存储（OSS/COS/S3）
5. 返回访问 URL 和存储 Key

#### 10.2 视频上传
- 【大文件处理方案】：
  - 方案A（当前）：后端接收 → 转存对象存储（适合小文件）
  - 方案B（推荐）：后端生成预签名 URL → 前端直传对象存储
- 方案B流程：
  1. 前端请求预签名 URL（/upload/presign?filename=xxx.mp4）
  2. 后端生成带过期时间的 PUT URL（如15分钟）
  3. 前端直传对象存储
  4. 前端通知后端上传完成，后端获取 duration 和封面

### 关键工程实现建议

| 模块 | 关键技术点 |
|------|-----------|
| **认证** | JWT（access_token）+ 刷新（refresh_token）+ 注销|
| **权限** | RBAC 数据模型 + 中间件权限校验 + 缓存 |
| **文件存储** | 对象存储（OSS）+ CDN 加速 |
| **安全** | SQL注入防护、XSS过滤、限流（ratelimit） |
