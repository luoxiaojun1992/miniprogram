# AI Instruction（Harness 协作指南）

> 本文档为本仓库 AI instruction 唯一真源（SSOT）。
> 入口与索引文件：`.github/copilot-instructions.md`、`.workbuddy/AI_AGENT_INSTRUCTIONS.md`、`docs/agent-instruction.md`

本文件用于指导 AI/开发者在本仓库内以“可控输入—标准流程—可验证输出”的方式完成改动。

## 1. 目标与范围

- 目标：以最小且完整的改动交付可验证结果，保持代码与文档一致。
- 适用范围：`luoxiaojun1992/miniprogram` 后端服务（Go + Gin）。
- 非目标：不做无关重构，不引入与需求无关的新依赖。

---

## 2. 项目事实卡（不可偏离）

- 语言与框架：Go 1.25 / Gin / GORM / MySQL / Redis
- 架构：`controller -> service -> repository -> model`
- 入口：`/cmd/server/main.go`
- 路由装配：`/internal/app/router.go`
- API 文档：`/api/docs/swagger.yaml`
- 配置：Viper（`CONFIG_PATH` + `APP_*` 覆盖）
- 鉴权：JWT
- 日志：Logrus（JSONFormatter）
- 测试工具：`go test`、k6（API）、Playwright（UI）

---

## 3. 系统地图（当前仓库结构）

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

## 4. 配置与启动规则（Harness 输入）

### 4.1 加载顺序
1. 读取环境变量 `CONFIG_PATH`
2. 若 `CONFIG_PATH` 非空，先加载配置文件
3. `APP_*` 环境变量始终覆盖配置项（Viper 规则）

### 4.2 关键配置项
- `server`: 端口与模式
- `database`: MySQL 连接信息
- `redis`: Redis 连接信息
- `jwt`: 密钥与过期时间
- `upload`: 上传存储（local/cos）
- `rate_limit`: 频控开关与阈值
- `debug.enable_test_token`: 是否启用 `/v1/debug/token`（仅非生产）

---

## 5. API / 路由 / 鉴权约定

- 健康检查：`GET /health`
- API 前缀：`/v1`
- 鉴权模型：
  - `optionalJWT`：可匿名访问，但可识别用户态
  - `requiredJWT`：必须登录
  - 管理端：`/v1/admin/*` + `RequireAdmin()` + 审计日志中间件
- 调试接口：`POST /v1/debug/token`（仅 `debug.enable_test_token=true` 启用）

新增/修改路由必须同步更新：
1. `internal/app/router.go`
2. `api/docs/swagger.yaml`

---

## 6. 分层职责与变更边界

### Controller（`internal/controller`）
- 负责：参数解析、鉴权上下文读取、调用 Service、返回统一响应
- 禁止：业务规则编排、DB 访问

### Service（`internal/service`）
- 负责：核心业务逻辑、跨仓储协作、事务语义与规则校验
- 禁止：HTTP 细节、SQL 拼接

### Repository（`internal/repository`）
- 负责：持久化与查询实现（基于 GORM）
- 禁止：业务策略判断

### Middleware（`internal/middleware`）
- 全局中间件在 `router.go` 注册
- 当前包括：Recovery、Error、CORS、RequestID、Logger、可选限流

---

## 7. 执行 Harness（输入 → 流程 → 输出）

### 7.1 输入信号
- 需求说明 / Issue / PR 变更描述
- 相关代码、配置、文档
- 既有测试与 CI 约束

### 7.2 标准流程
1. 先阅读相关代码与文档，再动手
2. 保持“最小且完整”的改动，避免无关重构
3. 涉及行为变化时补充或更新测试
4. 变更路由/协议时，确保 Swagger 与实现一致
5. 提交前通过质量门禁（见第 8 节）

### 7.3 输出要求
- 明确说明完成内容与影响面
- 列出已运行的命令与结果
- 标注需要人工关注的风险点

---

## 8. 质量门禁（可验证输出）

### 本地建议顺序
1. `go test -short ./...`
2. `go build -o /tmp/miniprogram-server ./cmd/server`
3. `golangci-lint run ./...`（可用时）

### Taskfile（与仓库保持一致）

```bash
task dev
task build
task test
task test-unit
task test-coverage
task lint
task lint-fix
task ui-test
task ui-test-down
```

### CI（`.github/workflows/ci.yml`）
1. Unit Tests（覆盖率阈值校验）
2. API Tests（k6）
3. UI Tests（Playwright，依赖 API Tests 成功后执行）

---

## 9. 变更触发器（同步文档）

当你修改以下内容时，必须同步更新文档：

- 路由/API：更新 `api/docs/swagger.yaml`
- 架构/流程：更新 `README.md` 与本文件
- 开发命令：更新 `Taskfile.yml` 后同步到 README/本文件
- 配置项：更新 `configs/config.yaml.sample` 与 README

---

## 10. 安全与发布 Guardrails

- 生产环境必须关闭：`debug.enable_test_token`
- 生产环境必须替换：
  - `APP_JWT_SECRET`
  - 数据库与 Redis 默认密码
- 上传使用 COS 时，避免将密钥写入仓库文件，优先使用环境变量注入
- 新增依赖时，优先使用成熟库并评估安全风险

---

## 11. 常见任务检查清单

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
