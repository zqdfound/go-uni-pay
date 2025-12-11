# 快速开始指南

## 🚀 管理后台快速体验

### 默认管理员账号

系统已经预设了一个默认管理员账号，可以直接使用：

**登录信息：**
- 🔑 用户名：`admin`
- 🔒 密码：`admin123`
- 📧 邮箱：`admin@example.com`

**访问地址：**
- 后端API：http://localhost:8080
- 前端管理后台：http://localhost:3000

### 初始化管理员账号

如果数据库中还没有管理员账号，执行以下SQL：

```sql
-- 创建管理员表（如果还没有创建）
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

-- 插入默认管理员（用户名：admin，密码：admin123）
INSERT INTO `admins` (`username`, `password`, `nickname`, `email`, `status`)
VALUES ('admin', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z2ELoYkSWLbk1cN5lZfRUBEu', '超级管理员', 'admin@example.com', 1);
```

### 生成自定义密码

如果需要创建其他管理员或修改密码，使用密码生成工具：

```bash
cd /Users/zqd/Desktop/zqd/my/go-uni-pay

# 生成密码hash
go run tools/gen_password.go your_password_here

# 示例：生成密码为 "mypassword" 的hash
go run tools/gen_password.go mypassword
```

输出示例：
```
========================================
密码生成成功!
========================================

原始密码: mypassword
加密密码: $2a$10$xxx...

SQL插入语句:
----------------------------------------
INSERT INTO `admins` (`username`, `password`, `nickname`, `email`, `status`)
VALUES ('your_username', '$2a$10$xxx...', '管理员', 'admin@example.com', 1);
----------------------------------------
```

### 启动管理后台

**1. 启动后端服务：**
```bash
cd /Users/zqd/Desktop/zqd/my/go-uni-pay
go run cmd/server/main.go
```

**2. 启动前端服务：**
```bash
cd /Users/zqd/Desktop/zqd/my/admin-frontend
npm install
npm run dev
```

**3. 访问管理后台：**

打开浏览器访问 http://localhost:3000，使用默认账号登录即可。

---

## 5分钟快速体验

### 1. 环境准备

确保你已安装：
- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### 2. 启动项目

```bash
# 克隆项目
git clone https://github.com/zqdfound/go-uni-pay.git
cd go-uni-pay

# 安装依赖
go mod tidy

# 初始化数据库
mysql -u root -p < scripts/init.sql

# 修改配置文件（更新数据库和Redis连接信息）
vim configs/config.yaml

# 启动服务
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

### 3. 测试 API

#### 健康检查

```bash
curl http://localhost:8080/health
```

#### 创建支付（支付宝示例）

**前提条件**：需要先在数据库中配置支付宝支付信息

```sql
-- 1. 使用默认测试用户（已在init.sql中创建）
-- API Key: ak_test_1234567890abcdef1234567890abcdef

-- 2. 添加支付宝配置
INSERT INTO payment_configs (user_id, provider, config_name, config_data, status)
VALUES (1, 'alipay', '测试配置', '{
  "app_id": "your_alipay_app_id",
  "private_key": "your_private_key",
  "public_key": "alipay_public_key",
  "is_production": false
}', 1);
```

**创建支付订单**：

```bash
curl -X POST http://localhost:8080/api/v1/payment/create \
  -H "X-API-Key: ak_test_1234567890abcdef1234567890abcdef" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "alipay",
    "out_trade_no": "TEST_ORDER_001",
    "subject": "测试商品",
    "amount": 0.01,
    "currency": "CNY",
    "notify_url": "http://your-domain.com/notify",
    "return_url": "http://your-domain.com/return"
  }'
```

**响应示例**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "order_no": "UNI20240101120000abcd1234",
    "payment_url": "https://openapi.alipay.com/gateway.do?...",
    "payment_id": "TEST_ORDER_001"
  }
}
```

#### 查询支付状态

```bash
curl -X GET "http://localhost:8080/api/v1/payment/query/UNI20240101120000abcd1234" \
  -H "X-API-Key: ak_test_1234567890abcdef1234567890abcdef"
```

## Docker 快速启动

如果你想使用 Docker 快速体验：

```bash
# 启动所有服务（MySQL + Redis + App）
docker-compose up -d

# 查看日志
docker-compose logs -f app

# 停止服务
docker-compose down
```

## 支付平台配置说明

### 支付宝沙箱环境

1. 访问：https://open.alipay.com/develop/sandbox/app
2. 获取以下信息：
   - AppID
   - 应用私钥
   - 支付宝公钥
3. 将配置信息插入数据库

### 微信支付测试

1. 申请微信支付商户号
2. 获取：
   - AppID
   - 商户号
   - API证书
   - APIv3密钥
3. 配置到数据库

### Stripe 测试

1. 注册 Stripe 账号：https://stripe.com
2. 获取测试密钥（以 `sk_test_` 开头）
3. 配置到数据库

### PayPal 测试

1. 访问：https://developer.paypal.com
2. 创建沙箱应用
3. 获取 Client ID 和 Secret
4. 配置到数据库（mode设置为"sandbox"）

## 常见问题

### Q1: 数据库连接失败

**错误**：`failed to connect database`

**解决**：
1. 检查 MySQL 是否启动：`sudo systemctl status mysql`
2. 检查配置文件中的数据库连接信息
3. 确认数据库已创建：`mysql -u root -p -e "SHOW DATABASES;"`

### Q2: Redis 连接失败

**错误**：`failed to connect redis`

**解决**：
1. 检查 Redis 是否启动：`sudo systemctl status redis`
2. 检查配置文件中的 Redis 连接信息
3. 测试连接：`redis-cli ping`（应返回 PONG）

### Q3: 如何创建新用户

目前需要手动在数据库中创建：

```sql
-- 生成随机 API Key 和 Secret
INSERT INTO users (username, email, api_key, api_secret, status)
VALUES (
  'new_user',
  'user@example.com',
  'ak_your_generated_key',
  '$2a$10$...hashed_secret',
  1
);
```

或使用认证服务的 CreateUser 方法。

### Q4: 支付配置如何管理

每个用户可以为每个支付平台配置多个配置，系统会使用状态为"启用"的第一个配置。

```sql
-- 查看用户的所有配置
SELECT * FROM payment_configs WHERE user_id = 1;

-- 禁用某个配置
UPDATE payment_configs SET status = 0 WHERE id = 1;

-- 启用某个配置
UPDATE payment_configs SET status = 1 WHERE id = 2;
```

## 开发建议

### 1. 使用 Makefile 命令

```bash
make help          # 查看所有可用命令
make build         # 编译项目
make run           # 运行项目
make test          # 运行测试
make clean         # 清理编译文件
```

### 2. 开启日志调试

修改 `configs/config.yaml`:

```yaml
logger:
  level: debug  # 设置为 debug 级别
```

### 3. 查看日志文件

```bash
# 实时查看日志
tail -f logs/app.log

# 查看错误日志
grep "ERROR" logs/app.log
```

## 下一步

1. 阅读完整的 [API 文档](docs/api.md)
2. 了解 [部署指南](docs/deployment.md)
3. 查看 [README](README.md) 了解更多功能

## 技术支持

- 提交 Issue: https://github.com/zqdfound/go-uni-pay/issues
- 查看文档: https://github.com/zqdfound/go-uni-pay/wiki
