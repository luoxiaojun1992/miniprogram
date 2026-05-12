# 变更约定与安全提示

## 变更边界
- 保持“最小且完整”的改动，避免无关重构
- Controller/Service/Repository 分层职责不可混用

## 文档同步触发器
- 路由/API：更新 `api/docs/swagger.yaml`
- 架构/流程：更新 `README.md` 与 `.agent/AI_AGENT_COMMON_INSTRUCTIONS.md`
- 开发命令：更新 `Taskfile.yml` 后同步到 README/本文件
- 配置项：更新 `configs/config.yaml.sample` 与 README

## 安全与发布
- 生产环境必须关闭：`debug.enable_test_token`
- 生产环境必须替换：`APP_JWT_SECRET`、数据库与 Redis 默认密码
- 使用 COS 时避免将密钥写入仓库文件，优先使用环境变量注入
