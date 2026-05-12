# 流程图索引（PlantUML）

本目录用于沉淀知识库后端项目的 **原生 PlantUML** 图。

## 目录约定

- 每个流程独立一个 `.puml` 文件，便于后续增量补充。
- 文件名建议：`api-<path>-<method>.puml` 或 `miniprogram-db-er.puml`。
- 图中标题包含 HTTP 方法与路径，便于与 Swagger 对齐。

## 当前流程图

### 数据库 ER 图

- `miniprogram-db-er.puml`

### API 时序图

#### 认证 Auth

- `api-auth-admin-login-post.puml` - POST /auth/admin-login
- `api-auth-refresh-post.puml` - POST /auth/refresh
- `api-auth-wechat-login-post.puml` - POST /auth/wechat-login

#### 用户 Users

- `api-users-permissions-get.puml` - GET /users/permissions
- `api-users-profile-get.puml` - GET /users/profile
- `api-users-profile-put.puml` - PUT /users/profile

#### 模块 Modules

- `api-modules-get.puml` - GET /modules

#### 轮播图 Banners

- `api-banners-get.puml` - GET /banners

#### 文章 Articles

- `api-articles-get.puml` - GET /articles
- `api-articles-id-get.puml` - GET /articles/{id}
- `api-articles-id-attachments-get.puml` - GET /articles/{id}/attachments

#### 课程 Courses

- `api-courses-get.puml` - GET /courses
- `api-courses-id-get.puml` - GET /courses/{id}
- `api-courses-id-attachments-get.puml` - GET /courses/{id}/attachments
- `api-courses-id-units-get.puml` - GET /courses/{id}/units
- `api-courses-id-units-unit_id-attachments-get.puml` - GET /courses/{id}/units/{unit_id}/attachments

#### 学习记录 Study Records

- `api-study-records-get.puml` - GET /study-records
- `api-study-records-post.puml` - POST /study-records

#### 收藏 Collections

- `api-collections-get.puml` - GET /collections
- `api-collections-content_type-content_id-delete.puml` - DELETE /collections/{content_type}/{content_id}
- `api-collections-content_type-content_id-post.puml` - POST /collections/{content_type}/{content_id}

#### 点赞 Likes

- `api-likes-content_type-content_id-delete.puml` - DELETE /likes/{content_type}/{content_id}
- `api-likes-content_type-content_id-post.puml` - POST /likes/{content_type}/{content_id}

#### 关注 Follows

- `api-follows-user_id-delete.puml` - DELETE /follows/{user_id}
- `api-follows-user_id-post.puml` - POST /follows/{user_id}

#### 评论 Comments

- `api-comments-content_type-content_id-get.puml` - GET /comments/{content_type}/{content_id}
- `api-comments-content_type-content_id-post.puml` - POST /comments/{content_type}/{content_id}

#### 通知 Notifications

- `api-notifications-get.puml` - GET /notifications
- `api-notifications-read-all-put.puml` - PUT /notifications/read-all
- `api-notifications-id-read-put.puml` - PUT /notifications/{id}/read

#### 上传 Upload

- `api-upload-avatar-post.puml` - POST /upload/avatar

#### 下载 Download

- `api-download-article-attachment-file_id-get.puml` - GET /download/article/attachment/{file_id}
- `api-download-banner-media-file_id-get.puml` - GET /download/banner/media/{file_id}
- `api-download-course-attachment-file_id-get.puml` - GET /download/course/attachment/{file_id}
- `api-download-course-unit-attachment-file_id-get.puml` - GET /download/course/unit/attachment/{file_id}
- `api-download-course-video-file_id-get.puml` - GET /download/course/video/{file_id}
- `api-download-static-file_id-get.puml` - GET /download/static/{file_id}

#### 管理端-用户

- `api-admin-users-get.puml` - GET /admin/users
- `api-admin-users-post.puml` - POST /admin/users
- `api-admin-users-id-delete.puml` - DELETE /admin/users/{id}
- `api-admin-users-id-get.puml` - GET /admin/users/{id}
- `api-admin-users-id-put.puml` - PUT /admin/users/{id}
- `api-admin-users-id-attributes-delete.puml` - DELETE /admin/users/{id}/attributes
- `api-admin-users-id-attributes-get.puml` - GET /admin/users/{id}/attributes
- `api-admin-users-id-attributes-post.puml` - POST /admin/users/{id}/attributes
- `api-admin-users-id-roles-put.puml` - PUT /admin/users/{id}/roles
- `api-admin-users-id-tags-delete.puml` - DELETE /admin/users/{id}/tags
- `api-admin-users-id-tags-post.puml` - POST /admin/users/{id}/tags

#### 管理端-角色

- `api-admin-roles-get.puml` - GET /admin/roles
- `api-admin-roles-post.puml` - POST /admin/roles
- `api-admin-roles-id-delete.puml` - DELETE /admin/roles/{id}
- `api-admin-roles-id-get.puml` - GET /admin/roles/{id}
- `api-admin-roles-id-put.puml` - PUT /admin/roles/{id}

#### 管理端-权限

- `api-admin-permissions-get.puml` - GET /admin/permissions

#### 管理端-属性

- `api-admin-attributes-get.puml` - GET /admin/attributes
- `api-admin-attributes-post.puml` - POST /admin/attributes
- `api-admin-attributes-id-delete.puml` - DELETE /admin/attributes/{id}
- `api-admin-attributes-id-put.puml` - PUT /admin/attributes/{id}

#### 管理端-模块

- `api-admin-modules-post.puml` - POST /admin/modules
- `api-admin-modules-id-delete.puml` - DELETE /admin/modules/{id}
- `api-admin-modules-id-put.puml` - PUT /admin/modules/{id}
- `api-admin-modules-id-pages-get.puml` - GET /admin/modules/{id}/pages
- `api-admin-modules-id-pages-post.puml` - POST /admin/modules/{id}/pages
- `api-admin-modules-id-pages-page_id-delete.puml` - DELETE /admin/modules/{id}/pages/{page_id}
- `api-admin-modules-id-pages-page_id-put.puml` - PUT /admin/modules/{id}/pages/{page_id}

#### 管理端-轮播图

- `api-admin-banners-get.puml` - GET /admin/banners
- `api-admin-banners-post.puml` - POST /admin/banners
- `api-admin-banners-id-delete.puml` - DELETE /admin/banners/{id}
- `api-admin-banners-id-put.puml` - PUT /admin/banners/{id}

#### 管理端-文章

- `api-admin-articles-get.puml` - GET /admin/articles
- `api-admin-articles-post.puml` - POST /admin/articles
- `api-admin-articles-id-delete.puml` - DELETE /admin/articles/{id}
- `api-admin-articles-id-get.puml` - GET /admin/articles/{id}
- `api-admin-articles-id-put.puml` - PUT /admin/articles/{id}
- `api-admin-articles-id-copy-post.puml` - POST /admin/articles/{id}/copy
- `api-admin-articles-id-pin-post.puml` - POST /admin/articles/{id}/pin
- `api-admin-articles-id-publish-post.puml` - POST /admin/articles/{id}/publish

#### 管理端-课程

- `api-admin-courses-get.puml` - GET /admin/courses
- `api-admin-courses-post.puml` - POST /admin/courses
- `api-admin-courses-id-delete.puml` - DELETE /admin/courses/{id}
- `api-admin-courses-id-get.puml` - GET /admin/courses/{id}
- `api-admin-courses-id-put.puml` - PUT /admin/courses/{id}
- `api-admin-courses-id-copy-post.puml` - POST /admin/courses/{id}/copy
- `api-admin-courses-id-pin-post.puml` - POST /admin/courses/{id}/pin
- `api-admin-courses-id-publish-post.puml` - POST /admin/courses/{id}/publish
- `api-admin-courses-id-units-get.puml` - GET /admin/courses/{id}/units
- `api-admin-courses-id-units-post.puml` - POST /admin/courses/{id}/units
- `api-admin-courses-id-units-unit_id-delete.puml` - DELETE /admin/courses/{id}/units/{unit_id}
- `api-admin-courses-id-units-unit_id-put.puml` - PUT /admin/courses/{id}/units/{unit_id}

#### 管理端-评论

- `api-admin-comments-get.puml` - GET /admin/comments
- `api-admin-comments-id-delete.puml` - DELETE /admin/comments/{id}
- `api-admin-comments-id-audit-put.puml` - PUT /admin/comments/{id}/audit

#### 管理端-系统

- `api-admin-audit-logs-get.puml` - GET /admin/audit-logs
- `api-admin-log-config-get.puml` - GET /admin/log-config
- `api-admin-log-config-put.puml` - PUT /admin/log-config
- `api-admin-wechat-config-get.puml` - GET /admin/wechat-config
- `api-admin-wechat-config-put.puml` - PUT /admin/wechat-config

#### 管理端-上传

- `api-admin-upload-article-attachment-presign-get.puml` - GET /admin/upload/article/attachment/presign
- `api-admin-upload-banner-media-presign-get.puml` - GET /admin/upload/banner/media/presign
- `api-admin-upload-course-attachment-presign-get.puml` - GET /admin/upload/course/attachment/presign
- `api-admin-upload-course-unit-attachment-presign-get.puml` - GET /admin/upload/course/unit/attachment/presign
- `api-admin-upload-course-video-presign-get.puml` - GET /admin/upload/course/video/presign
- `api-admin-upload-files-presign-get.puml` - GET /admin/upload/files/presign

#### 调试 Debug

- `api-debug-token-post.puml` - POST /debug/token

#### 健康检查

- `api-health-get.puml` - GET /health
