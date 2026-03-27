# AI Instruction（项目协作指引）

本文件用于指导 AI/开发者在本仓库内进行一致、可落地的开发与维护工作。

## 1. 项目概述

- 项目：`luoxiaojun1992/miniprogram`
- 类型：Go + Gin 的小程序/管理后台后端服务
- 架构：`controller -> service -> repository -> model`
- 入口：`/cmd/server/main.go`
- 路由：`/internal/app/router.go`
- API 文档：`/api/docs/swagger.yaml`

---

## 2. 技术栈与关键依赖

- Go 版本：`go 1.25.0`（见 `go.mod`）
- Web 框架：Gin
- 配置：Viper（支持 `CONFIG_PATH` + `APP_*` 环境变量）
- 日志：Logrus（JSONFormatter）
- 数据库：MySQL + GORM
- 缓存/限流存储：Redis
- 鉴权：JWT
- 测试：
  - Go 单元/集成：`go test`
  - API 压测/回归：k6（`docker-compose.api-test.yml`）
  - UI 自动化：Playwright（`docker-compose.ui-test.yml`）

---

## 3. 目录结构（当前仓库真实结构）

```text
cmd/server/main.go                # 服务启动入口
internal/
  app/                            # 配置、Provider、路由装配
  controller/                     # HTTP 层（参数绑定/响应）
  service/                        # 业务逻辑层
  repository/                     # 数据访问层
  model/
    entity/                       # 领域实体（DB 模型）
    dto/                          # 请求/响应 DTO
  middleware/                     # Gin 中间件
  pkg/                            # 内部工具（errors/response/wechat/cos 等）
api/docs/swagger.yaml             # OpenAPI 2.0 文档
migrations/                       # SQL 迁移
tests/
  api/                            # k6 API 测试
  ui/                             # Playwright UI 测试
```

---

## 4. 配置与启动约定

### 4.1 配置加载规则

1. 程序读取环境变量 `CONFIG_PATH`
2. 若 `CONFIG_PATH` 非空，先加载配置文件
3. `APP_*` 环境变量始终可覆盖配置项（Viper 规则）

关键代码位置：
- `cmd/server/main.go`
- `internal/app/config.go`

### 4.2 关键配置项

- `server`: 端口与模式
- `database`: MySQL 连接信息
- `redis`: Redis 连接信息
- `jwt`: 密钥与过期时间
- `upload`: 上传存储（local/cos）
- `rate_limit`: 频控开关与阈值
- `debug.enable_test_token`: 是否启用 `/v1/debug/token`（仅非生产）

---

## 5. 架构与分层职责（必须遵守）

### Controller 层（`internal/controller`）
- 负责：参数解析、鉴权上下文读取、调用 Service、返回统一响应
- 不负责：业务规则编排、DB 访问

### Service 层（`internal/service`）
- 负责：核心业务逻辑、跨仓储协作、事务语义与规则校验
- 不负责：HTTP 细节、SQL 拼接

### Repository 层（`internal/repository`）
- 负责：持久化与查询实现（基于 GORM）
- 不负责：业务策略判断

### Middleware 层（`internal/middleware`）
- 全局中间件在 `router.go` 注册
- 当前包括：Recovery、Error、CORS、RequestID、Logger、可选限流

---

## 6. API 与路由约定

- 健康检查：`GET /health`
- API 前缀：`/v1`
- 鉴权模型：
  - `optionalJWT`：可匿名访问，但可识别用户态
  - `requiredJWT`：必须登录
  - 管理端：`/v1/admin/*` + `RequireAdmin()` + 审计日志中间件
- 调试接口：
  - `POST /v1/debug/token`
  - 仅当 `debug.enable_test_token=true` 时启用

开发时新增/修改路由，需同时更新：
1. `internal/app/router.go`
2. `api/docs/swagger.yaml`

---

## 7. 开发命令（以仓库现有 Taskfile 为准）

```bash
task dev              # air 热重载
task build            # go build -o bin/server ./cmd/server
task test             # go test ./...
task test-unit        # go test -short ./...
task lint             # golangci-lint run ./...
task ui-test          # UI 自动化测试（docker compose）
task ui-test-down     # 清理 UI 测试环境
```

如果本地没有 `task`，可直接执行对应原生命令。

---

## 8. 测试与质量门禁

### 本地建议顺序
1. `go test -short ./...`
2. `go build -o /tmp/miniprogram-server ./cmd/server`
3. （可用时）`golangci-lint run ./...`

### CI（`.github/workflows/ci.yml`）
按顺序执行：
1. Unit Tests（覆盖率阈值校验）
2. API Tests（k6）
3. UI Tests（Playwright）

其中 UI Job 依赖 API Job 成功后才会运行。

---

## 9. 文档维护规则

当你修改以下内容时，必须同步更新文档：

- 路由/API：更新 `api/docs/swagger.yaml`
- 架构/流程：更新 `README.md` 与本文件
- 开发命令：更新 `Taskfile.yml` 后同步到 README/本文件
- 配置项：更新 `configs/config.yaml.sample` 与 README

---

## 10. 安全与发布注意事项

- 生产环境必须关闭：`debug.enable_test_token`
- 生产环境必须替换：
  - `APP_JWT_SECRET`
  - 数据库与 Redis 默认密码
- 上传使用 COS 时，避免将密钥写入仓库文件，优先使用环境变量注入
- 新增依赖时，优先使用成熟库并评估安全风险

---

## 11. AI 变更执行要求（面向自动化代理）

1. 先阅读相关代码与文档，再改动
2. 保持“最小且完整”的改动，避免无关重构
3. 涉及行为变化时补充或更新测试
4. 提交前至少通过：
   - `go test -short ./...`
   - `go build -o /tmp/miniprogram-server ./cmd/server`
5. 变更路由/协议时，确保 Swagger 与实现一致

---

## 12. 常见任务检查清单

### 新增接口
- [ ] 在 `router.go` 注册路由
- [ ] 在 controller/service/repository 完成分层实现
- [ ] 更新 `api/docs/swagger.yaml`
- [ ] 补充对应测试

### 调整鉴权/权限
- [ ] 检查 `requiredJWT` / `optionalJWT` 使用位置
- [ ] 检查 `RequireAdmin()` 与角色权限影响范围
- [ ] 回归管理端相关 API 与 UI 流程

### 调整配置项
- [ ] 更新 `internal/app/config.go` 默认值与映射
- [ ] 更新 `configs/config.yaml.sample`
- [ ] 更新 `README.md` 配置说明

