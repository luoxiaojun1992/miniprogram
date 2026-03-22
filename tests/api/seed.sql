-- ============================================
-- API test seed data
-- Seeds two users that k6 uses via the debug token endpoint.
--   user_id=1  user_type=3  system admin  (email: admin@example.com / Test@123456)
--   user_id=2  user_type=1  regular user
-- ============================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- Test admin user (user_type=3: system admin)
INSERT INTO `users` (`id`, `nickname`, `user_type`, `created_at`, `updated_at`)
VALUES (1, 'Test Admin', 3, NOW(), NOW());

-- Admin login credentials: admin@example.com / Test@123456
INSERT INTO `admin_users` (`user_id`, `email`, `password_hash`)
VALUES (1, 'admin@example.com', '$2b$10$1dR4uktUZmJfaT8NBnZRGOAxzQLYlnTS3aCUmvUUSnLm96da1jHzK');

-- Assign system-admin role (built-in role id=1) to the admin user
INSERT INTO `user_roles` (`user_id`, `role_id`) VALUES (1, 1);

-- Test regular user (user_type=1: front-end user)
INSERT INTO `users` (`id`, `open_id`, `nickname`, `user_type`, `created_at`, `updated_at`)
VALUES (2, 'test_regular_user_openid', 'Test User', 1, NOW(), NOW());

-- Assign front-user role so permission endpoints return stable role/permission data in tests
INSERT IGNORE INTO `user_roles` (`user_id`, `role_id`) VALUES (2, 4);

-- Front-user defaults used in API/UI tests
INSERT IGNORE INTO `role_permissions` (`role_id`, `permission_id`) VALUES
  (4, 14), -- article:view
  (4, 19), -- course:view
  (4, 25); -- comment:view

SET FOREIGN_KEY_CHECKS = 1;
