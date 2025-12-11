# 数据库迁移文档

本文档记录了支付管理系统的所有数据库表结构。

## 版本历史

- **版本 001**: 2025-12-11 - 初始版本，创建所有基础表
- **版本 002**: 2025-12-11 - 添加管理员表

---

## 表结构说明

### 1. users - 用户表

存储系统用户信息。

```sql
CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` varchar(50) NOT NULL COMMENT '用户名',
  `email` varchar(100) NOT NULL COMMENT '邮箱',
  `api_key` varchar(64) NOT NULL COMMENT 'API密钥',
  `api_secret` varchar(128) NOT NULL COMMENT 'API密钥（加密）',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：0-禁用，1-启用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_email` (`email`),
  UNIQUE KEY `uk_api_key` (`api_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
```

### 2. payment_configs - 支付配置表

存储各支付渠道的配置信息。

```sql
CREATE TABLE IF NOT EXISTS `payment_configs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '配置ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `provider` varchar(20) NOT NULL COMMENT '支付渠道：alipay/wechat/paypal/stripe',
  `config_name` varchar(50) NOT NULL COMMENT '配置名称',
  `config_data` json NOT NULL COMMENT '配置数据',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：0-禁用，1-启用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_provider` (`provider`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付配置表';
```

### 3. payment_orders - 支付订单表

存储所有支付订单信息。

```sql
CREATE TABLE IF NOT EXISTS `payment_orders` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `order_no` varchar(64) NOT NULL COMMENT '系统订单号',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `provider` varchar(20) NOT NULL COMMENT '支付渠道',
  `config_id` bigint unsigned NOT NULL COMMENT '配置ID',
  `out_trade_no` varchar(64) NOT NULL COMMENT '商户订单号',
  `trade_no` varchar(64) DEFAULT NULL COMMENT '第三方订单号',
  `subject` varchar(256) NOT NULL COMMENT '订单标题',
  `body` text COMMENT '订单描述',
  `amount` decimal(10,2) NOT NULL COMMENT '订单金额',
  `currency` varchar(10) NOT NULL DEFAULT 'CNY' COMMENT '币种',
  `status` varchar(20) NOT NULL DEFAULT 'pending' COMMENT '订单状态：pending/processing/success/failed/closed',
  `notify_url` varchar(512) DEFAULT NULL COMMENT '异步通知URL',
  `return_url` varchar(512) DEFAULT NULL COMMENT '同步跳转URL',
  `client_ip` varchar(45) DEFAULT NULL COMMENT '客户端IP',
  `extra_data` json DEFAULT NULL COMMENT '额外数据',
  `payment_time` datetime DEFAULT NULL COMMENT '支付时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_provider` (`provider`),
  KEY `idx_out_trade_no` (`out_trade_no`),
  KEY `idx_trade_no` (`trade_no`),
  KEY `idx_status` (`status`),
  KEY `idx_payment_time` (`payment_time`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付订单表';
```

### 4. payment_logs - 支付日志表

记录支付相关的所有操作日志。

```sql
CREATE TABLE IF NOT EXISTS `payment_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `order_id` bigint unsigned NOT NULL COMMENT '订单ID',
  `order_no` varchar(64) NOT NULL COMMENT '订单号',
  `action` varchar(50) NOT NULL COMMENT '操作类型：create/query/notify等',
  `provider` varchar(20) NOT NULL COMMENT '支付渠道',
  `request_data` json DEFAULT NULL COMMENT '请求数据',
  `response_data` json DEFAULT NULL COMMENT '响应数据',
  `status` varchar(20) NOT NULL COMMENT '状态',
  `error_msg` text COMMENT '错误信息',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_order_no` (`order_no`),
  KEY `idx_action` (`action`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付日志表';
```

### 5. api_logs - API调用日志表

记录所有API调用信息。

```sql
CREATE TABLE IF NOT EXISTS `api_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `api_key` varchar(64) NOT NULL COMMENT 'API密钥',
  `method` varchar(10) NOT NULL COMMENT 'HTTP方法',
  `path` varchar(255) NOT NULL COMMENT '请求路径',
  `query` text COMMENT '查询参数',
  `request_body` text COMMENT '请求体',
  `response_status` int NOT NULL COMMENT '响应状态码',
  `response_body` text COMMENT '响应体',
  `ip` varchar(45) DEFAULT NULL COMMENT '客户端IP',
  `user_agent` varchar(512) DEFAULT NULL COMMENT 'User Agent',
  `duration` int DEFAULT NULL COMMENT '请求耗时(毫秒)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_api_key` (`api_key`),
  KEY `idx_path` (`path`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='API调用日志表';
```

### 6. notify_queue - 通知队列表

管理支付通知的重试队列。

```sql
CREATE TABLE IF NOT EXISTS `notify_queue` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '队列ID',
  `order_id` bigint unsigned NOT NULL COMMENT '订单ID',
  `order_no` varchar(64) NOT NULL COMMENT '订单号',
  `notify_url` varchar(512) NOT NULL COMMENT '通知URL',
  `notify_data` json NOT NULL COMMENT '通知数据',
  `retry_count` int NOT NULL DEFAULT 0 COMMENT '重试次数',
  `max_retry` int NOT NULL DEFAULT 5 COMMENT '最大重试次数',
  `status` varchar(20) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/processing/success/failed',
  `last_error` text COMMENT '最后一次错误信息',
  `next_retry_time` datetime DEFAULT NULL COMMENT '下次重试时间',
  `success_time` datetime DEFAULT NULL COMMENT '成功时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_order_no` (`order_no`),
  KEY `idx_status` (`status`),
  KEY `idx_next_retry_time` (`next_retry_time`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知队列表';
```

### 7. admins - 管理员表

存储管理后台的管理员信息。

```sql
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
-- 注意：password字段需要使用bcrypt加密后的值
-- 默认密码：admin123
INSERT INTO `admins` (`username`, `password`, `nickname`, `email`, `status`)
VALUES ('admin', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z2ELoYkSWLbk1cN5lZfRUBEu', '超级管理员', 'admin@example.com', 1);
```

---

## 索引说明

所有表都包含以下标准索引：
- 主键索引：`id`
- 唯一索引：根据业务需求设置（如 `username`, `email`, `api_key`, `order_no`等）
- 普通索引：常用查询字段（如 `user_id`, `status`, `created_at`等）

## 字符集和排序规则

- 字符集：`utf8mb4`
- 排序规则：`utf8mb4_unicode_ci`
- 存储引擎：`InnoDB`

## 备注

1. 所有时间字段使用 `datetime` 类型
2. 金额字段使用 `decimal(10,2)` 类型确保精度
3. JSON字段用于存储灵活的配置和数据
4. 所有表都包含 `created_at` 和 `updated_at` 字段，使用MySQL自动更新

## 初始化脚本

可以使用以下命令初始化数据库：

```bash
mysql -u root -p < /path/to/migrations/init.sql
```

## 密码生成

管理员密码使用bcrypt加密，可以使用以下Go代码生成：

```go
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    password := "admin123"
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    fmt.Println(string(hashedPassword))
}
```
