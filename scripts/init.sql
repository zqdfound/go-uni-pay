-- 多平台支付集成平台数据库初始化脚本
-- 创建数据库
CREATE DATABASE IF NOT EXISTS uni_pay DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE uni_pay;

-- 用户表（API调用方）
CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '用户ID',
    `username` VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    `email` VARCHAR(100) NOT NULL UNIQUE COMMENT '邮箱',
    `api_key` VARCHAR(64) NOT NULL UNIQUE COMMENT 'API密钥',
    `api_secret` VARCHAR(128) NOT NULL COMMENT 'API秘钥',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1-启用 0-禁用',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX `idx_api_key` (`api_key`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 支付配置表
CREATE TABLE IF NOT EXISTS `payment_configs` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '配置ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `provider` VARCHAR(20) NOT NULL COMMENT '支付提供商：alipay/wechat/stripe/paypal',
    `config_name` VARCHAR(50) NOT NULL COMMENT '配置名称',
    `config_data` JSON NOT NULL COMMENT '配置数据（JSON格式）',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1-启用 0-禁用',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_provider` (`provider`),
    INDEX `idx_status` (`status`),
    UNIQUE KEY `uk_user_provider_name` (`user_id`, `provider`, `config_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付配置表';

-- 支付订单表
CREATE TABLE IF NOT EXISTS `payment_orders` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '订单ID',
    `order_no` VARCHAR(64) NOT NULL UNIQUE COMMENT '订单号',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `provider` VARCHAR(20) NOT NULL COMMENT '支付提供商',
    `config_id` BIGINT UNSIGNED NOT NULL COMMENT '支付配置ID',
    `out_trade_no` VARCHAR(64) NOT NULL COMMENT '商户订单号',
    `trade_no` VARCHAR(64) DEFAULT NULL COMMENT '第三方交易号',
    `subject` VARCHAR(256) NOT NULL COMMENT '订单标题',
    `body` TEXT COMMENT '订单描述',
    `amount` DECIMAL(10, 2) NOT NULL COMMENT '订单金额',
    `currency` VARCHAR(10) NOT NULL DEFAULT 'CNY' COMMENT '货币类型',
    `status` VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '订单状态：pending/processing/success/failed/closed',
    `notify_url` VARCHAR(512) COMMENT '异步通知URL',
    `return_url` VARCHAR(512) COMMENT '同步跳转URL',
    `client_ip` VARCHAR(45) COMMENT '客户端IP',
    `extra_data` JSON COMMENT '额外数据',
    `payment_time` TIMESTAMP NULL COMMENT '支付时间',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX `idx_order_no` (`order_no`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_out_trade_no` (`out_trade_no`),
    INDEX `idx_trade_no` (`trade_no`),
    INDEX `idx_status` (`status`),
    INDEX `idx_provider` (`provider`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付订单表';

-- 支付日志表
CREATE TABLE IF NOT EXISTS `payment_logs` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '日志ID',
    `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    `order_no` VARCHAR(64) NOT NULL COMMENT '订单号',
    `action` VARCHAR(50) NOT NULL COMMENT '操作类型：create/pay/query/refund/notify',
    `provider` VARCHAR(20) NOT NULL COMMENT '支付提供商',
    `request_data` JSON COMMENT '请求数据',
    `response_data` JSON COMMENT '响应数据',
    `status` VARCHAR(20) NOT NULL COMMENT '状态：success/failed',
    `error_msg` TEXT COMMENT '错误信息',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX `idx_order_id` (`order_id`),
    INDEX `idx_order_no` (`order_no`),
    INDEX `idx_action` (`action`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付日志表';

-- API调用记录表
CREATE TABLE IF NOT EXISTS `api_logs` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '日志ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `api_key` VARCHAR(64) NOT NULL COMMENT 'API密钥',
    `method` VARCHAR(10) NOT NULL COMMENT 'HTTP方法',
    `path` VARCHAR(255) NOT NULL COMMENT '请求路径',
    `query` TEXT COMMENT '查询参数',
    `request_body` TEXT COMMENT '请求体',
    `response_status` INT NOT NULL COMMENT '响应状态码',
    `response_body` TEXT COMMENT '响应体',
    `ip` VARCHAR(45) COMMENT '客户端IP',
    `user_agent` VARCHAR(512) COMMENT 'User-Agent',
    `duration` INT COMMENT '请求耗时（毫秒）',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_api_key` (`api_key`),
    INDEX `idx_path` (`path`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='API调用记录表';

-- 通知队列表（用于异步通知和重试）
CREATE TABLE IF NOT EXISTS `notify_queue` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '队列ID',
    `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    `order_no` VARCHAR(64) NOT NULL COMMENT '订单号',
    `notify_url` VARCHAR(512) NOT NULL COMMENT '通知URL',
    `notify_data` JSON NOT NULL COMMENT '通知数据',
    `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
    `max_retry` INT NOT NULL DEFAULT 5 COMMENT '最大重试次数',
    `status` VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/processing/success/failed',
    `last_error` TEXT COMMENT '最后错误信息',
    `next_retry_time` TIMESTAMP NULL COMMENT '下次重试时间',
    `success_time` TIMESTAMP NULL COMMENT '成功时间',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX `idx_order_id` (`order_id`),
    INDEX `idx_order_no` (`order_no`),
    INDEX `idx_status` (`status`),
    INDEX `idx_next_retry_time` (`next_retry_time`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知队列表';

-- 插入测试用户
INSERT INTO `users` (`username`, `email`, `api_key`, `api_secret`, `status`)
VALUES ('test_user', 'test@example.com', 'ak_test_1234567890abcdef1234567890abcdef',
        '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 1);

-- 创建视图：订单统计
CREATE OR REPLACE VIEW `v_order_stats` AS
SELECT
    user_id,
    provider,
    status,
    COUNT(*) as order_count,
    SUM(amount) as total_amount,
    DATE(created_at) as order_date
FROM payment_orders
GROUP BY user_id, provider, status, DATE(created_at);
