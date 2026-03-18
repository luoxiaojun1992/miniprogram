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

-- 敏感词表
CREATE TABLE IF NOT EXISTS `sensitive_words` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `word` VARCHAR(128) NOT NULL COMMENT '敏感词',
    `status` TINYINT DEFAULT 1 COMMENT '0禁用 1启用',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_word` (`word`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='敏感词表';

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

-- 属性表
CREATE TABLE IF NOT EXISTS `attributes` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(64) NOT NULL COMMENT '属性名称',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='属性表';

-- 用户属性表
CREATE TABLE IF NOT EXISTS `user_attributes` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `attribute_id` INT UNSIGNED NOT NULL COMMENT '属性ID',
    `value` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '属性值',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`attribute_id`) REFERENCES `attributes`(`id`) ON DELETE CASCADE,
    UNIQUE INDEX `idx_user_attribute` (`user_id`, `attribute_id`),
    INDEX `idx_attribute` (`attribute_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户属性表';

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

-- 轮播图表
CREATE TABLE IF NOT EXISTS `banners` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `title` VARCHAR(128),
    `image_file_id` BIGINT UNSIGNED,
    `link_url` VARCHAR(255),
    `sort_order` INT DEFAULT 0,
    `status` TINYINT DEFAULT 1 COMMENT '0禁用 1启用',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX `idx_image_file` (`image_file_id`),
    INDEX `idx_status_sort` (`status`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='轮播图表';

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
    `comment_count` INT UNSIGNED DEFAULT 0,
    `share_count` INT UNSIGNED DEFAULT 0,
    `sort_order` INT DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (`author_id`) REFERENCES `users`(`id`),
    FOREIGN KEY (`module_id`) REFERENCES `modules`(`id`),
    INDEX `idx_status_time` (`status`, `publish_time`),
    INDEX `idx_module` (`module_id`),
    INDEX `idx_sort` (`sort_order`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章表';

-- 文件表
CREATE TABLE IF NOT EXISTS `files` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `key` VARCHAR(255) NOT NULL,
    `filename` VARCHAR(255) NOT NULL,
    `usage` VARCHAR(32) NOT NULL,
    `category` VARCHAR(32) NOT NULL,
    `business` VARCHAR(64),
    `static_url` VARCHAR(512),
    `created_by` BIGINT UNSIGNED,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_key` (`key`),
    INDEX `idx_usage_category` (`usage`, `category`),
    FOREIGN KEY (`created_by`) REFERENCES `users`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件表';

-- 课程表
CREATE TABLE IF NOT EXISTS `courses` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `title` VARCHAR(200) NOT NULL,
    `description` TEXT,
    `cover_image` VARCHAR(255),
    `duration` INT UNSIGNED COMMENT '总课时(分钟)',
    `author_id` BIGINT UNSIGNED,
    `module_id` INT UNSIGNED,
    `status` TINYINT DEFAULT 0 COMMENT '0草稿 1已发布 2定时发布',
    `publish_time` DATETIME,
    `price` DECIMAL(10,2) DEFAULT 0.00 COMMENT '价格，0为免费',
    `view_count` INT UNSIGNED DEFAULT 0,
    `like_count` INT UNSIGNED DEFAULT 0,
    `collect_count` INT UNSIGNED DEFAULT 0,
    `comment_count` INT UNSIGNED DEFAULT 0,
    `share_count` INT UNSIGNED DEFAULT 0,
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
    `video_file_id` BIGINT UNSIGNED,
    `duration` INT UNSIGNED COMMENT '课时(分钟)',
    `sort_order` INT DEFAULT 0,
    `status` TINYINT DEFAULT 1,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`course_id`) REFERENCES `courses`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`video_file_id`) REFERENCES `files`(`id`),
    INDEX `idx_course_sort` (`course_id`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='课程单元表';

-- 文章附件表
CREATE TABLE IF NOT EXISTS `article_attachments` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `article_id` BIGINT UNSIGNED NOT NULL,
    `file_id` BIGINT UNSIGNED NOT NULL,
    `sort_order` INT DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`article_id`) REFERENCES `articles`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`file_id`) REFERENCES `files`(`id`) ON DELETE CASCADE,
    UNIQUE KEY `uk_article_file` (`article_id`, `file_id`),
    INDEX `idx_article_sort` (`article_id`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章附件关联表';

-- 课程附件表
CREATE TABLE IF NOT EXISTS `course_attachments` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `course_id` BIGINT UNSIGNED NOT NULL,
    `file_id` BIGINT UNSIGNED NOT NULL,
    `sort_order` INT DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`course_id`) REFERENCES `courses`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`file_id`) REFERENCES `files`(`id`) ON DELETE CASCADE,
    UNIQUE KEY `uk_course_file` (`course_id`, `file_id`),
    INDEX `idx_course_sort` (`course_id`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='课程附件关联表';

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
    `type` TINYINT DEFAULT 1 COMMENT '1系统通知 2评论回复 3学习提醒 4点赞通知',
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
(32, '轮播图管理', 'banner:management', 3, 1, 12, 2),
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
(2, 19), (2, 20), (2, 21), (2, 23), (2, 32),
(2, 25), (2, 26),      -- 评论管理
(2, 30);               -- 日志查看

-- 为内容编辑分配内容相关权限
INSERT INTO `role_permissions` (`role_id`, `permission_id`) VALUES 
(3, 14), (3, 15), (3, 16), (3, 18),
(3, 19), (3, 20), (3, 21), (3, 23),
(3, 25), (3, 26);

-- 初始化敏感词（可在后台继续维护）
-- 数据参考：
-- 1) https://github.com/fwwdn/sensitive-stop-words （Apache-2.0，中文词库）
-- 2) https://github.com/zacanger/profane-words （WTFPL，英文词库）
-- 3) https://github.com/houbb/sensitive-word （Apache-2.0，扩展中文敏感词库）
INSERT INTO `sensitive_words` (`word`, `status`) VALUES
('色情', 1),
('赌博', 1),
('诈骗', 1),
('暴力', 1),
('违禁', 1),
('爱液', 1),
('按摩棒', 1),
('爆草', 1),
('暴奸', 1),
('被操', 1),
('被插', 1),
('被干', 1),
('逼奸', 1),
('操逼', 1),
('肏你', 1),
('肏死', 1),
('操死', 1),
('插逼', 1),
('插阴', 1),
('潮吹', 1),
('成人色情', 1),
('成人网站', 1),
('春药', 1),
('荡妇', 1),
('盗撮', 1),
('肥逼', 1),
('干穴', 1),
('肛交', 1),
('肛门', 1),
('龟头', 1),
('国产av', 1),
('黑逼', 1),
('后庭', 1),
('黄片', 1),
('鸡巴', 1),
('鸡奸', 1),
('妓女', 1),
('叫床', 1),
('精液', 1),
('巨屌', 1),
('菊花洞', 1),
('口爆', 1),
('口交', 1),
('口淫', 1),
('狂操', 1),
('浪女', 1),
('凌辱', 1),
('露b', 1),
('乱交', 1),
('乱伦', 1),
('轮奸', 1),
('买春', 1),
('奶子', 1),
('内射', 1),
('嫩逼', 1),
('嫩穴', 1),
('女优', 1),
('炮友', 1),
('喷精', 1),
('屁眼', 1),
('强暴', 1),
('强奸处女', 1),
('情色', 1),
('群交', 1),
('人兽', 1),
('日逼', 1),
('肉棒', 1),
('肉洞', 1),
('肉欲', 1),
('乳房', 1),
('乳交', 1),
('乳头', 1),
('骚逼', 1),
('骚女', 1),
('色情网站', 1),
('色欲', 1),
('手淫', 1),
('兽交', 1),
('兽欲', 1),
('熟女', 1),
('丝袜', 1),
('舔阴', 1),
('调教', 1),
('偷欢', 1),
('脱内裤', 1),
('吸精', 1),
('小穴', 1),
('性交', 1),
('性奴', 1),
('性虐', 1),
('性欲', 1),
('颜射', 1),
('阳具', 1),
('阴部', 1),
('阴唇', 1),
('阴道', 1),
('淫荡', 1),
('淫妇', 1),
('淫贱', 1),
('淫女', 1),
('淫水', 1),
('淫液', 1),
('应召', 1),
('幼交', 1),
('欲女', 1),
('援交', 1),
('援助交际', 1),
('招妓', 1),
('抓胸', 1),
('自慰', 1),
('作爱', 1),
('a片', 1),
('g点', 1),
('h动画', 1),
('失身粉', 1),
('淫荡自慰器', 1),
('anal', 1),
('anus', 1),
('arse', 1),
('arsehole', 1),
('ass', 1),
('asshole', 1),
('bastard', 1),
('bdsm', 1),
('bitch', 1),
('blowjob', 1),
('bollocks', 1),
('boner', 1),
('boob', 1),
('boobs', 1),
('butt', 1),
('butthole', 1),
('clit', 1),
('cock', 1),
('cum', 1),
('cumshot', 1),
('cunt', 1),
('deepthroat', 1),
('dick', 1),
('dildo', 1),
('ejaculation', 1),
('erotic', 1),
('fellatio', 1),
('fingering', 1),
('fisting', 1),
('fuck', 1),
('fucking', 1),
('gangbang', 1),
('handjob', 1),
('hentai', 1),
('horny', 1),
('incest', 1),
('jerk off', 1),
('jizz', 1),
('kinky', 1),
('masturbation', 1),
('milf', 1),
('motherfucker', 1),
('naked', 1),
('nigger', 1),
('nude', 1),
('nudity', 1),
('orgasm', 1),
('orgy', 1),
('penis', 1),
('porn', 1),
('porno', 1),
('pornography', 1),
('pussy', 1),
('rape', 1),
('raping', 1),
('rapist', 1),
('rimjob', 1),
('rimming', 1),
('sex', 1),
('sexcam', 1),
('sexy', 1),
('shemale', 1),
('shit', 1),
('slut', 1),
('sodomy', 1),
('strip club', 1),
('suck', 1),
('threesome', 1),
('tit', 1),
('tits', 1),
('titties', 1),
('twat', 1),
('vagina', 1),
('vibrator', 1),
('vulva', 1),
('whore', 1),
('xxx', 1),
('nsfw', 1),
('suicide girls', 1),
('习近平', 1),
('胡锦涛', 1),
('江泽民', 1),
('温家宝', 1),
('周永康', 1),
('薄熙来', 1),
('紫阳', 1),
('中南海', 1),
('共产党', 1),
('中共', 1),
('共匪', 1),
('共产专制', 1),
('北京当局', 1),
('中国当局', 1),
('法轮功', 1),
('李洪志', 1),
('党产共', 1),
('共残主义', 1),
('裆中央', 1),
('六四', 1),
('天安门事件', 1),
('八九学运', 1),
('学潮', 1),
('民运', 1),
('民主墙', 1),
('一党专政', 1),
('专制政权', 1),
('独裁政权', 1),
('政治审查', 1),
('言论审查', 1),
('新闻封锁', 1),
('网络审查', 1),
('维稳', 1),
('上访', 1),
('访民', 1),
('强拆', 1),
('官商勾结', 1),
('贪官污吏', 1),
('太子党', 1),
('红二代', 1),
('政治局常委', 1),
('党禁', 1),
('报禁', 1),
('军管', 1),
('戒严', 1),
('异见人士', 1),
('政治犯', 1),
('良心犯', 1),
('反革命', 1),
('颠覆国家政权', 1),
('煽动颠覆', 1),
('境外势力', 1),
('达赖喇嘛', 1),
('班禅', 1),
('东突', 1),
('疆独', 1),
('藏独', 1),
('台独', 1),
('港独', 1),
('新疆集中营', 1),
('西藏独立', 1);

SET FOREIGN_KEY_CHECKS = 1;
