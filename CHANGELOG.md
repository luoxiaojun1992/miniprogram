# Changelog

All notable changes to this project will be documented in this file.

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
