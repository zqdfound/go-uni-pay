# Go-Uni-Pay - 统一支付网关平台

## 项目简介

Go-Uni-Pay 是一个基于 Go 语言和 Gin 框架开发的多平台支付集成平台。它提供统一的支付接口，支持支付宝、微信、Stripe、PayPal 等多种支付方式，让开发者只需配置相应的支付平台参数即可快速集成支付功能。

## 核心特性

- ✅ **多支付平台支持**：支付宝、微信支付、Stripe、PayPal
- ✅ **统一接口设计**：提供统一的 API 接口，简化集成流程
- ✅ **异步通知处理**：支持异步通知回调及失败重试机制
- ✅ **支付记录查询**：完整的支付订单查询和日志追踪
- ✅ **API 认证鉴权**：基于 API Key 的认证系统
- ✅ **调用记录追踪**：完整记录每次 API 调用
- ✅ **分布式支持**：支持 Redis 缓存和分布式锁
- ✅ **多节点部署**：可水平扩展，支持高并发场景
- ✅ **清晰的代码结构**：遵循 Clean Architecture 设计原则

## 技术栈

- **语言**：Go 1.21+
- **Web 框架**：Gin
- **数据库**：MySQL 8.0+
- **缓存**：Redis 6.0+
- **ORM**：GORM
- **日志**：Zap + Lumberjack
- **配置管理**：Viper

## 项目结构

```
go-uni-pay/
├── cmd/
│   └── server/           # 程序入口
├── internal/
│   ├── api/              # API 层
│   │   ├── handler/      # 请求处理器
│   │   ├── middleware/   # 中间件
│   │   └── router/       # 路由配置
│   ├── domain/           # 领域层
│   │   ├── entity/       # 实体定义
│   │   └── repository/   # 仓储接口
│   ├── service/          # 业务逻辑层
│   │   ├── payment/      # 支付服务
│   │   ├── notify/       # 通知服务
│   │   └── auth/         # 认证服务
│   ├── infrastructure/   # 基础设施层
│   │   ├── database/     # 数据库
│   │   ├── cache/        # 缓存
│   │   ├── lock/         # 分布式锁
│   │   └── config/       # 配置管理
│   └── payment/          # 支付提供商实现
│       ├── alipay/       # 支付宝
│       ├── wechat/       # 微信支付
│       ├── stripe/       # Stripe
│       └── paypal/       # PayPal
├── pkg/                  # 公共库
│   ├── logger/           # 日志
│   ├── errors/           # 错误处理
│   └── utils/            # 工具函数
├── configs/              # 配置文件
├── scripts/              # 脚本文件
└── docs/                 # 文档
```

## 快速开始

### 1. 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### 2. 克隆项目

```bash
git clone https://github.com/zqdfound/go-uni-pay.git
cd go-uni-pay
```

### 3. 安装依赖

```bash
go mod tidy
```

### 4. 初始化数据库

```bash
# 创建数据库并导入表结构
mysql -u root -p < scripts/init.sql
```

### 5. 配置文件

修改 `configs/config.yaml` 文件，配置数据库和 Redis 连接信息：

```yaml
database:
  host: localhost
  port: 3306
  username: root
  password: your_password
  database: uni_pay

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
```

### 6. 启动服务

```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

## API 使用指南

### 认证

所有需要认证的 API 请求都需要在 Header 中携带 API Key：

```
X-API-Key: your_api_key
```

### 创建支付

**请求：**

```bash
POST /api/v1/payment/create
Content-Type: application/json
X-API-Key: your_api_key

{
  "provider": "alipay",
  "out_trade_no": "ORDER_20240101_001",
  "subject": "商品标题",
  "body": "商品描述",
  "amount": 0.01,
  "currency": "CNY",
  "notify_url": "https://your-domain.com/callback",
  "return_url": "https://your-domain.com/return"
}
```

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "order_no": "UNI20240101120000abcd1234",
    "payment_url": "https://openapi.alipay.com/...",
    "payment_id": "ORDER_20240101_001",
    "qr_code": ""
  }
}
```

### 查询支付

**请求：**

```bash
GET /api/v1/payment/query/:order_no
X-API-Key: your_api_key
```

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "order_no": "UNI20240101120000abcd1234",
    "out_trade_no": "ORDER_20240101_001",
    "trade_no": "2024010122001234567890",
    "status": "success",
    "amount": 0.01,
    "currency": "CNY",
    "payment_time": "2024-01-01T12:00:00Z"
  }
}
```

## 支付配置

### 支付宝配置

在数据库中插入支付配置：

```sql
INSERT INTO payment_configs (user_id, provider, config_name, config_data, status)
VALUES (1, 'alipay', '默认配置', '{
  "app_id": "your_app_id",
  "private_key": "your_private_key",
  "public_key": "alipay_public_key",
  "is_production": false
}', 1);
```

### 微信支付配置

```sql
INSERT INTO payment_configs (user_id, provider, config_name, config_data, status)
VALUES (1, 'wechat', '默认配置', '{
  "app_id": "your_app_id",
  "mch_id": "your_mch_id",
  "serial_no": "your_serial_no",
  "api_v3_key": "your_api_v3_key",
  "private_key": "your_private_key"
}', 1);
```

### Stripe 配置

```sql
INSERT INTO payment_configs (user_id, provider, config_name, config_data, status)
VALUES (1, 'stripe', '默认配置', '{
  "secret_key": "sk_test_...",
  "webhook_secret": "whsec_..."
}', 1);
```

### PayPal 配置

```sql
INSERT INTO payment_configs (user_id, provider, config_name, config_data, status)
VALUES (1, 'paypal', '默认配置', '{
  "client_id": "your_client_id",
  "secret": "your_secret",
  "mode": "sandbox"
}', 1);
```

## 分布式部署

### Redis 配置

项目使用 Redis 实现：
- 缓存：存储临时数据，提高访问速度
- 分布式锁：确保多节点下的数据一致性

### 多节点部署

1. 确保所有节点连接到同一个 MySQL 和 Redis 实例
2. 启动多个服务实例
3. 使用负载均衡器（如 Nginx）分发请求

```nginx
upstream go_uni_pay {
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
    server 127.0.0.1:8082;
}

server {
    listen 80;
    location / {
        proxy_pass http://go_uni_pay;
    }
}
```

## 异步通知与重试

系统自动处理支付通知的异步推送：

- 失败自动重试，支持指数退避策略
- 最多重试 5 次
- 重试间隔：1分钟、2分钟、5分钟、10分钟、30分钟

## 开发指南

### 添加新的支付提供商

1. 在 `internal/payment/` 下创建新的提供商目录
2. 实现 `Provider` 接口
3. 在 `init()` 函数中注册提供商
4. 在 `main.go` 中导入提供商包

示例：

```go
package newprovider

import (
    "github.com/zqdfound/go-uni-pay/internal/payment"
)

type Provider struct{}

func (p *Provider) GetName() string {
    return "newprovider"
}

// 实现其他接口方法...

func init() {
    payment.Register(&Provider{})
}
```

## 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1000 | 内部服务器错误 |
| 1001 | 参数错误 |
| 1002 | 未授权 |
| 1003 | 禁止访问 |
| 1004 | 资源未找到 |
| 2000 | 创建支付失败 |
| 2001 | 查询支付失败 |
| 2005 | 支付提供商未找到 |
| 2006 | 支付配置未找到 |
| 2007 | 订单未找到 |

## 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/service/payment

# 运行测试并显示覆盖率
go test -cover ./...
```

## 性能优化建议

1. **数据库优化**
   - 为常用查询字段添加索引
   - 使用连接池
   - 定期清理历史数据

2. **缓存策略**
   - 缓存支付配置
   - 缓存用户信息
   - 设置合理的过期时间

3. **并发控制**
   - 使用分布式锁防止重复支付
   - 限制 API 调用频率

## 安全建议

1. **API 密钥安全**
   - 定期轮换 API 密钥
   - 不要在代码中硬编码密钥
   - 使用 HTTPS 传输

2. **支付安全**
   - 验证支付通知签名
   - 检查订单金额
   - 防止重复支付

3. **数据安全**
   - 敏感信息加密存储
   - 定期备份数据库
   - 限制数据库访问权限

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 联系方式

- 作者：zqdfound
- Email: your-email@example.com
- GitHub: https://github.com/zqdfound/go-uni-pay

## 更新日志

### v1.0.0 (2024-01-01)

- 初始版本发布
- 支持支付宝、微信、Stripe、PayPal 四种支付方式
- 实现异步通知和失败重试机制
- 支持分布式部署
