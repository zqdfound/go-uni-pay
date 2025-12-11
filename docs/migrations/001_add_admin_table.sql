-- 管理员表创建脚本
-- 版本: 001
-- 描述: 添加管理员表用于后台管理系统
-- 日期: 2025-12-11

CREATE TABLE IF NOT EXISTS `admins` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '管理员ID',
  `username` varchar(50) NOT NULL COMMENT '用户名',
  `password` varchar(128) NOT NULL COMMENT '密码（加密）',
  `nickname` varchar(50) DEFAULT NULL COMMENT '昵称',
  `email` varchar(100) DEFAULT NULL COMMENT '邮箱',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：0-禁用，1-启用',
  `last_login` datetime DEFAULT NULL COMMENT '最后登录时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员表';

-- 插入默认管理员账号
-- 用户名: admin
-- 密码: admin123 (使用bcrypt加密，这里需要在应用中生成实际的hash值)
-- 注意：在实际部署时，应该通过应用程序生成密码hash并插入
INSERT INTO `admins` (`username`, `password`, `nickname`, `email`, `status`)
VALUES ('admin', '$2a$10$placeholder', '超级管理员', 'admin@example.com', 1);

-- 注意：上面的password字段值'$2a$10$placeholder'只是占位符
-- 需要使用实际的bcrypt hash值替换
-- Go代码生成示例：
-- import "golang.org/x/crypto/bcrypt"
-- hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
