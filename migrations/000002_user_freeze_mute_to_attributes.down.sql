ALTER TABLE `users`
    ADD COLUMN `status` TINYINT DEFAULT 1 COMMENT '0冻结 1正常' AFTER `user_type`,
    ADD COLUMN `freeze_end_time` DATETIME COMMENT '冻结结束时间' AFTER `status`,
    DROP INDEX `idx_user_type`,
    ADD INDEX `idx_type_status` (`user_type`, `status`);
