# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased] - 2026-03-19

### Added
- **CHANGELOG 文档** - 新增项目变更日志文档

## [Unreleased] - 2026-03-18

### Added
- **数据模型重构** - 头像及文章/课程计数器迁移到类型化属性系统（user/article/course attribute 表）
- **互动通知** - 点赞/评论后发送系统通知，补全计数器更新链路
- **删除关联保护** - 实体存在关联时禁止删除，防止孤数据
- **敏感词过滤** - 对文章内容和评论统一过滤，扩充中文词库
- **管理员内容管理** - 新增文章/课程的置顶、复制 API
- **评论安全** - 拒绝含 HTML 标签的评论写入

### Changed
- **文件类型验证** - 下载签名前通过 COS HeadObject 校验真实 MIME 类型

### Fixed
- 修复多处 Playwright UI 测试和 k6 API 测试的脆性问题

## [Unreleased] - 2026-03-17

### Added
- **互动通知与计数补齐** - 点赞/评论后发送通知，文章/课程计数器更新能力补全
- **删除关联保护** - 模块、文章、课程、用户等实体删除前关联校验
- **敏感词库扩充** - 新增中文政治敏感词，添加第三词库来源
- **管理员内容管理 API** - 文章/课程置顶、复制功能
- **评论安全** - 拒绝包含 HTML 标签的评论写入
- **点赞通知类型** - 通知系统新增类型 4（点赞通知）
- **内容计数器** - 文章/课程新增评论数、分享数字段

### Changed
- **数据库迁移整合** - 属性相关迁移合并到初始迁移文件
- **敏感词数据库驱动** - 从数据库加载敏感词进行内容过滤

### Fixed
- 修复用户核心路径测试覆盖率问题

## [Unreleased] - 2026-03-14

### Added
- **UI 测试稳定化** - 修复 Playwright 严格模式违规问题，解决 v-show 标签页切换导致的元素定位错误

### Changed
- **文档结构调整** - OpenAPI 规范移至 api/docs/，代理指南移至 docs/

### Fixed
- 修复页面标题定位匹配隐藏元素问题
- 修复搜索输入框定位重复元素问题
- 修复 HTML 报告器阻塞容器退出问题

## [Unreleased] - 2026-03-13

### Fixed
- 修复 Viper AutomaticEnv 忽略嵌套 APP_* 环境变量问题（缺少 key replacer）
