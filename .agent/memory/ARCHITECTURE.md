# 项目架构

## 项目概述
- 后端服务：Go 1.25 + Gin + GORM
- 数据存储：MySQL、Redis
- 鉴权：JWT
- 入口：`/cmd/server/main.go`
- 路由装配：`/internal/app/router.go`

## 分层职责
- Controller：参数解析、鉴权上下文读取、调用 Service、返回统一响应
- Service：核心业务逻辑、跨仓储协作、事务语义与规则校验
- Repository：数据访问与查询实现（基于 GORM）
- Model：领域实体与 DTO
- Middleware：Recovery、Error、CORS、RequestID、Logger、限流

## 关键路径
- 路由装配：`internal/app/router.go`
- 服务入口：`cmd/server/main.go`
- 配置加载：`internal/app/config.go`
- Swagger：`api/docs/swagger.yaml`

## 外部依赖
- WeChat：登录与用户态相关能力
- COS：对象存储（上传/下载）

## 相关图示
- 架构图：`.agent/diagrams/miniprogram-architecture.puml`
- 数据库 ER：`.agent/diagrams/miniprogram-db-er.puml`
