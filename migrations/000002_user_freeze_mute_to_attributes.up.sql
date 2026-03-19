ALTER TABLE `users`
    DROP COLUMN `status`,
    DROP COLUMN `freeze_end_time`,
    DROP INDEX `idx_type_status`,
    ADD INDEX `idx_user_type` (`user_type`);
