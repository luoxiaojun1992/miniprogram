-- ============================================
-- 知识库后端系统 - 数据库初始化脚本
-- 版本: 1.0.0
-- 创建时间: 2026-03-09
-- 数据库: MySQL 8.0+
-- 字符集: utf8mb4
-- ============================================

-- 设置字符集
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================
-- 1. 基础配置表
-- ============================================

-- 微信配置表
CREATE TABLE IF NOT EXISTS `wechat_configs` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `app_id` VARCHAR(32) NOT NULL,
    `app_secret` VARCHAR(64) NOT NULL,
    `api_token` VARCHAR(255) COMMENT '微信API Token',
    `js_api_ticket` VARCHAR(512),
    `ticket_expires_at` DATETIME,
    `access_token` VARCHAR(512),
    `token_expires_at` DATETIME,
    `status` TINYINT DEFAULT 1,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='微信配置表';

-- 日志配置表
CREATE TABLE IF NOT EXISTS `log_configs` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `retention_days` INT DEFAULT 90 COMMENT '日志保留天数',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='日志配置表';

-- 插入默认日志配置
INSERT INTO `log_configs` (`retention_days`) VALUES (90);

-- ============================================
-- 2. 用户与权限模块
-- ============================================

-- 用户表
CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `open_id` VARCHAR(64) UNIQUE COMMENT '微信openid',
    `union_id` VARCHAR(64) COMMENT '微信unionid',
    `nickname` VARCHAR(64) COMMENT '用户昵称',
    `avatar_url` VARCHAR(255) COMMENT '头像URL',
    `user_type` TINYINT DEFAULT 1 COMMENT '1前台用户 2普通管理员 3系统管理员',
    `status` TINYINT DEFAULT 1 COMMENT '0冻结 1正常',
    `freeze_end_time` DATETIME COMMENT '冻结结束时间',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME NULL COMMENT '软删除时间',
    INDEX `idx_open_id` (`open_id`),
    INDEX `idx_type_status` (`user_type`, `status`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 管理员扩展表
CREATE TABLE IF NOT EXISTS `admin_users` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `email` VARCHAR(128) UNIQUE,
    `password_hash` VARCHAR(255),
    `last_login_at` DATETIME,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员扩展表';

-- 角色表
CREATE TABLE IF NOT EXISTS `roles` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(64) NOT NULL,
    `description` VARCHAR(255),
    `parent_id` INT UNSIGNED DEFAULT 0 COMMENT '父角色ID，0为顶级',
    `level` TINYINT DEFAULT 1 COMMENT '层级',
    `is_builtin` TINYINT DEFAULT 0 COMMENT '是否内置角色',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_parent` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';

-- 权限表
CREATE TABLE IF NOT EXISTS `permissions` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(64) NOT NULL COMMENT '权限名称',
    `code` VARCHAR(128) NOT NULL UNIQUE COMMENT '权限编码，如 article:create',
    `type` TINYINT DEFAULT 1 COMMENT '1菜单 2按钮 3接口',
    `parent_id` INT UNSIGNED DEFAULT 0,
    `level` TINYINT DEFAULT 1,
    `is_builtin` TINYINT DEFAULT 0 COMMENT '是否内置权限',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_parent` (`parent_id`),
    INDEX `idx_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限表';

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS `user_roles` (
    `user_id` BIGINT UNSIGNED,
    `role_id` INT UNSIGNED,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`user_id`, `role_id`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`role_id`) REFERENCES `roles`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联表';

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS `role_permissions` (
    `role_id` INT UNSIGNED,
    `permission_id` INT UNSIGNED,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`role_id`, `permission_id`),
    FOREIGN KEY (`role_id`) REFERENCES `roles`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`permission_id`) REFERENCES `permissions`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限关联表';

-- 用户标签表
CREATE TABLE IF NOT EXISTS `user_tags` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED,
    `tag_name` VARCHAR(32),
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    INDEX `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户标签表';

-- ============================================
-- 3. 内容管理模块
-- ============================================

-- 模块表（课程/文章分类）
CREATE TABLE IF NOT EXISTS `modules` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `title` VARCHAR(128) NOT NULL,
    `description` TEXT,
    `sort_order` INT DEFAULT 0,
    `status` TINYINT DEFAULT 1 COMMENT '0禁用 1启用',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_sort` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='模块表';

-- 模块页面表（富文本内容）
CREATE TABLE IF NOT EXISTS `module_pages` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `module_id` INT UNSIGNED,
    `title` VARCHAR(128),
    `content` LONGTEXT COMMENT '富文本内容',
    `content_type` TINYINT DEFAULT 1 COMMENT '1富文本 2HTML',
    `sort_order` INT DEFAULT 0,
    `status` TINYINT DEFAULT 1,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (`module_id`) REFERENCES `modules`(`id`) ON DELETE CASCADE,
    INDEX `idx_module_sort` (`module_id`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='模块页面表';

-- 文章表
CREATE TABLE IF NOT EXISTS `articles` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `title` VARCHAR(200) NOT NULL,
    `summary` VARCHAR(500),
    `content` LONGTEXT,
    `content_type` TINYINT DEFAULT 1 COMMENT '1富文本 2HTML 3Markdown',
    `cover_image` VARCHAR(255),
    `author_id` BIGINT UNSIGNED,
    `module_id` INT UNSIGNED,
    `status` TINYINT DEFAULT 0 COMMENT '0草稿 1已发布 2定时发布',
    `publish_time` DATETIME COMMENT '定时发布时间',
    `view_count` INT UNSIGNED DEFAULT 0,
    `like_count` INT UNSIGNED DEFAULT 0,
    `collect_count` INT UNSIGNED DEFAULT 0,
    `sort_order` INT DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (`author_id`) REFERENCES `users`(`id`),
    FOREIGN KEY (`module_id`) REFERENCES `modules`(`id`),
    INDEX `idx_status_time` (`status`, `publish_time`),
    INDEX `idx_module` (`module_id`),
    INDEX `idx_sort` (`sort_order`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章表';

-- 课程表
CREATE TABLE IF NOT EXISTS `courses` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `title` VARCHAR(200) NOT NULL,
    `description` TEXT,
    `cover_image` VARCHAR(255),
    `video_url` VARCHAR(255),
    `duration` INT UNSIGNED COMMENT '总课时(分钟)',
    `author_id` BIGINT UNSIGNED,
    `module_id` INT UNSIGNED,
    `status` TINYINT DEFAULT 0 COMMENT '0草稿 1已发布 2定时发布',
    `publish_time` DATETIME,
    `price` DECIMAL(10,2) DEFAULT 0.00 COMMENT '价格，0为免费',
    `view_count` INT UNSIGNED DEFAULT 0,
    `like_count` INT UNSIGNED DEFAULT 0,
    `collect_count` INT UNSIGNED DEFAULT 0,
    `study_count` INT UNSIGNED DEFAULT 0 COMMENT '学习人数',
    `sort_order` INT DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (`author_id`) REFERENCES `users`(`id`),
    FOREIGN KEY (`module_id`) REFERENCES `modules`(`id`),
    INDEX `idx_status_time` (`status`, `publish_time`),
    INDEX `idx_module` (`module_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='课程表';

-- 课程单元表
CREATE TABLE IF NOT EXISTS `course_units` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `course_id` BIGINT UNSIGNED,
    `title` VARCHAR(200),
    `video_url` VARCHAR(255),
    `duration` INT UNSIGNED COMMENT '课时(分钟)',
    `sort_order` INT DEFAULT 0,
    `status` TINYINT DEFAULT 1,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`course_id`) REFERENCES `courses`(`id`) ON DELETE CASCADE,
    INDEX `idx_course_sort` (`course_id`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='课程单元表';

-- 内容权限关联表（查看权限控制）
CREATE TABLE IF NOT EXISTS `content_permissions` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `content_type` TINYINT COMMENT '1文章 2课程',
    `content_id` BIGINT UNSIGNED,
    `role_id` INT UNSIGNED COMMENT 'null表示公开',
    `permission_type` TINYINT DEFAULT 1 COMMENT '1查看 2编辑',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_content` (`content_type`, `content_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='内容权限关联表';

-- ============================================
-- 4. 互动与收藏模块
-- ============================================

-- 用户学习记录表
CREATE TABLE IF NOT EXISTS `user_study_records` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED,
    `course_id` BIGINT UNSIGNED,
    `unit_id` BIGINT UNSIGNED,
    `progress` INT UNSIGNED DEFAULT 0 COMMENT '学习进度(秒)',
    `status` TINYINT DEFAULT 0 COMMENT '0未开始 1学习中 2已完成',
    `last_study_at` DATETIME,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`course_id`) REFERENCES `courses`(`id`) ON DELETE CASCADE,
    UNIQUE KEY `uk_user_unit` (`user_id`, `unit_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户学习记录表';

-- 收藏表
CREATE TABLE IF NOT EXISTS `collections` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED,
    `content_type` TINYINT COMMENT '1文章 2课程',
    `content_id` BIGINT UNSIGNED,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    UNIQUE KEY `uk_user_content` (`user_id`, `content_type`, `content_id`),
    INDEX `idx_content` (`content_type`, `content_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='收藏表';

-- 点赞表
CREATE TABLE IF NOT EXISTS `likes` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED,
    `content_type` TINYINT COMMENT '1文章 2课程',
    `content_id` BIGINT UNSIGNED,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    UNIQUE KEY `uk_user_content` (`user_id`, `content_type`, `content_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='点赞表';

-- 评论表
CREATE TABLE IF NOT EXISTS `comments` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED,
    `content_type` TINYINT COMMENT '1文章 2课程',
    `content_id` BIGINT UNSIGNED,
    `parent_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '回复评论ID',
    `content` TEXT NOT NULL,
    `status` TINYINT DEFAULT 1 COMMENT '0待审核 1通过 2拒绝',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    INDEX `idx_content` (`content_type`, `content_id`, `status`),
    INDEX `idx_parent` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='评论表';

-- ============================================
-- 5. 消息通知与日志模块
-- ============================================

-- 消息通知表
CREATE TABLE IF NOT EXISTS `notifications` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED COMMENT 'null为全站广播',
    `type` TINYINT DEFAULT 1 COMMENT '1系统通知 2评论回复 3学习提醒',
    `title` VARCHAR(128),
    `content` TEXT,
    `is_read` TINYINT DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_user_read` (`user_id`, `is_read`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息通知表';

-- 审计日志表
CREATE TABLE IF NOT EXISTS `audit_logs` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED,
    `username` VARCHAR(64) COMMENT '操作人昵称',
    `action` VARCHAR(64) COMMENT '操作类型',
    `module` VARCHAR(64) COMMENT '操作模块',
    `description` TEXT,
    `ip_address` VARCHAR(45),
    `user_agent` VARCHAR(255),
    `request_data` JSON,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_user_time` (`user_id`, `created_at`),
    INDEX `idx_module_action` (`module`, `action`),
    INDEX `idx_time` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审计日志表';

-- ============================================
-- 6. 初始化内置数据
-- ============================================

-- 初始化系统管理员角色（内置）
INSERT INTO `roles` (`id`, `name`, `description`, `is_builtin`, `level`) VALUES 
(1, '系统管理员', '拥有所有权限', 1, 1),
(2, '普通管理员', '基础管理权限', 1, 1),
(3, '内容编辑', '文章课程管理权限', 1, 1),
(4, '前台用户', '普通用户', 1, 1);

-- 初始化基础权限（内置）
INSERT INTO `permissions` (`id`, `name`, `code`, `type`, `is_builtin`, `parent_id`, `level`) VALUES 
-- 用户管理权限
(1, '用户管理', 'user:management', 1, 1, 0, 1),
(2, '用户查看', 'user:view', 3, 1, 1, 2),
(3, '用户编辑', 'user:edit', 3, 1, 1, 2),
(4, '用户删除', 'user:delete', 3, 1, 1, 2),
(5, '用户冻结', 'user:freeze', 3, 1, 1, 2),
-- 角色权限管理
(6, '角色管理', 'role:management', 1, 1, 0, 1),
(7, '角色查看', 'role:view', 3, 1, 6, 2),
(8, '角色编辑', 'role:edit', 3, 1, 6, 2),
(9, '角色删除', 'role:delete', 3, 1, 6, 2),
(10, '权限查看', 'permission:view', 3, 1, 6, 2),
(11, '权限分配', 'permission:assign', 3, 1, 6, 2),
-- 内容管理权限
(12, '内容管理', 'content:management', 1, 1, 0, 1),
(13, '模块管理', 'module:management', 3, 1, 12, 2),
(14, '文章查看', 'article:view', 3, 1, 12, 2),
(15, '文章创建', 'article:create', 3, 1, 12, 2),
(16, '文章编辑', 'article:edit', 3, 1, 12, 2),
(17, '文章删除', 'article:delete', 3, 1, 12, 2),
(18, '文章发布', 'article:publish', 3, 1, 12, 2),
(19, '课程查看', 'course:view', 3, 1, 12, 2),
(20, '课程创建', 'course:create', 3, 1, 12, 2),
(21, '课程编辑', 'course:edit', 3, 1, 12, 2),
(22, '课程删除', 'course:delete', 3, 1, 12, 2),
(23, '课程发布', 'course:publish', 3, 1, 12, 2),
-- 评论审核权限
(24, '评论管理', 'comment:management', 1, 1, 0, 1),
(25, '评论查看', 'comment:view', 3, 1, 24, 2),
(26, '评论审核', 'comment:audit', 3, 1, 24, 2),
(27, '评论删除', 'comment:delete', 3, 1, 24, 2),
-- 系统设置权限
(28, '系统设置', 'system:management', 1, 1, 0, 1),
(29, '微信配置', 'wechat:config', 3, 1, 28, 2),
(30, '日志查看', 'log:view', 3, 1, 28, 2),
(31, '日志配置', 'log:config', 3, 1, 28, 2);

-- 为系统管理员角色分配所有权限
INSERT INTO `role_permissions` (`role_id`, `permission_id`)
SELECT 1, id FROM `permissions` WHERE `is_builtin` = 1;

-- 为普通管理员分配基础权限（不含系统设置和角色删除）
INSERT INTO `role_permissions` (`role_id`, `permission_id`) VALUES 
(2, 2), (2, 3), (2, 5), -- 用户管理（不含删除）
(2, 7), (2, 10),       -- 角色查看、权限查看
(2, 13), (2, 14), (2, 15), (2, 16), (2, 18), -- 内容管理
(2, 19), (2, 20), (2, 21), (2, 23),
(2, 25), (2, 26),      -- 评论管理
(2, 30);               -- 日志查看

-- 为内容编辑分配内容相关权限
INSERT INTO `role_permissions` (`role_id`, `permission_id`) VALUES 
(3, 14), (3, 15), (3, 16), (3, 18),
(3, 19), (3, 20), (3, 21), (3, 23),
(3, 25), (3, 26);

SET FOREIGN_KEY_CHECKS = 1;
