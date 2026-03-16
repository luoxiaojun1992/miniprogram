-- ============================================
-- 属性表与用户属性表
-- ============================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

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

SET FOREIGN_KEY_CHECKS = 1;
