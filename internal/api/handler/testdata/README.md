# 支付通知测试数据说明

本目录包含不同支付渠道的 Webhook 通知测试数据，用于测试支付通知处理功能。

## 文件列表

### 1. alipay_notify.txt
**支付宝支付通知测试数据**

- **格式**: URL 编码的表单数据 (application/x-www-form-urlencoded)
- **字符集**: UTF-8
- **签名方式**: RSA2

**关键字段说明**:
- `out_trade_no`: 商户订单号 (TEST20240101000001)
- `trade_no`: 支付宝交易号 (2024010122001400001234567890)
- `trade_status`: 交易状态 (TRADE_SUCCESS - 支付成功)
- `total_amount`: 订单金额 (100.00 CNY)
- `buyer_logon_id`: 买家支付宝账号 (test***@alipay.com)
- `gmt_payment`: 支付时间 (2024-01-01 10:05:30)
- `sign`: RSA2 签名

**使用场景**:
- 支付成功通知
- 订单状态更新测试

---

### 2. wechat_notify.xml
**微信支付通知测试数据**

- **格式**: XML
- **字符集**: UTF-8
- **签名方式**: MD5/HMAC-SHA256

**关键字段说明**:
- `out_trade_no`: 商户订单号 (TEST20240101000001)
- `transaction_id`: 微信支付订单号 (4200001234567890123456789012)
- `return_code`: 返回状态码 (SUCCESS)
- `result_code`: 业务结果 (SUCCESS)
- `total_fee`: 订单金额，单位：分 (10000 = 100.00 CNY)
- `time_end`: 支付完成时间 (20240101100530)
- `openid`: 用户标识 (oUpF8uMuAJO_M2pxb1Q9zNjWeS6o)
- `sign`: 签名

**使用场景**:
- 支付成功通知
- Native 支付回调测试

---

### 3. stripe_notify.json
**Stripe 支付通知测试数据**

- **格式**: JSON
- **事件类型**: checkout.session.completed
- **API 版本**: 2023-10-16

**关键字段说明**:
- `id`: 事件 ID (evt_1OTestEvent123456)
- `type`: 事件类型 (checkout.session.completed)
- `data.object.id`: 会话 ID (cs_test_a1b2c3d4e5f6g7h8i9j0)
- `data.object.client_reference_id`: 商户订单号 (TEST20240101000001)
- `data.object.amount_total`: 订单金额，单位：分 (10000 = $100.00)
- `data.object.payment_status`: 支付状态 (paid)
- `data.object.customer_details.email`: 客户邮箱
- `created`: 事件创建时间戳 (1704096330)

**请求头要求**:
- `Content-Type`: application/json
- `Stripe-Signature`: Webhook 签名头

**使用场景**:
- 结账会话完成通知
- 订单支付成功测试

---

### 4. paypal_notify.json
**PayPal 支付通知测试数据**

- **格式**: JSON
- **事件类型**: PAYMENT.CAPTURE.COMPLETED
- **资源版本**: 2.0

**关键字段说明**:
- `id`: Webhook 事件 ID (WH-1AB23456CD789012E-3FG45678HI901234J)
- `event_type`: 事件类型 (PAYMENT.CAPTURE.COMPLETED)
- `resource.id`: 捕获 ID (2AB34567CD890123E)
- `resource.custom_id`: 商户订单号 (TEST20240101000001)
- `resource.supplementary_data.related_ids.order_id`: PayPal 订单 ID
- `resource.amount.value`: 金额 ($100.00)
- `resource.status`: 状态 (COMPLETED)
- `resource.payer.email_address`: 付款人邮箱
- `create_time`: 创建时间 (2024-01-01T10:05:00Z)

**请求头要求**:
- `Content-Type`: application/json
- `Paypal-Transmission-Id`: 传输 ID
- `Paypal-Transmission-Time`: 传输时间
- `Paypal-Transmission-Sig`: 传输签名
- `Paypal-Cert-Url`: 证书 URL
- `Paypal-Auth-Algo`: 签名算法 (SHA256withRSA)

**使用场景**:
- 支付捕获完成通知
- 订单支付成功测试

---

## 测试用例说明

### 运行测试

```bash
# 运行所有测试
go test -v ./internal/api/handler

# 运行特定测试
go test -v ./internal/api/handler -run TestHandleNotify_Alipay

# 运行基准测试
go test -bench=. ./internal/api/handler

# 查看测试覆盖率
go test -cover ./internal/api/handler
```

### 测试覆盖范围

1. **正常流程测试**
   - 支付宝支付成功通知
   - 微信支付成功通知
   - Stripe 支付成功通知
   - PayPal 支付成功通知

2. **异常流程测试**
   - 无效的支付提供商
   - 服务层错误处理

3. **性能测试**
   - 支付通知处理的基准测试

## 注意事项

1. **签名验证**: 测试数据中的签名是示例值，实际生产环境中需要使用真实的签名验证
2. **金额单位**:
   - 支付宝/微信: 人民币元
   - Stripe: 美分 (需除以 100)
   - PayPal: 美元
3. **时间格式**:
   - 支付宝: yyyy-MM-dd HH:mm:ss
   - 微信: yyyyMMddHHmmss
   - Stripe: Unix 时间戳
   - PayPal: ISO 8601 格式
4. **编码格式**: 支付宝使用 URL 编码，其他使用 JSON/XML

## 扩展测试

如需添加更多测试场景，可以参考以下示例：

### 支付失败通知
修改 `trade_status` 或 `result_code` 为失败状态

### 退款通知
使用对应的退款事件类型和数据结构

### 部分支付通知
调整金额字段测试部分支付场景

## 相关文档

- [支付宝开放平台文档](https://opendocs.alipay.com/)
- [微信支付开发文档](https://pay.weixin.qq.com/wiki/doc/api/)
- [Stripe API 文档](https://stripe.com/docs/api)
- [PayPal Developer](https://developer.paypal.com/docs/)