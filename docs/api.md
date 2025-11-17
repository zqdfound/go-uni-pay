# API 文档

## 基础信息

- **Base URL**: `http://localhost:8080`
- **Content-Type**: `application/json`
- **认证方式**: API Key (Header: `X-API-Key`)

## 通用响应格式

### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### 错误响应

```json
{
  "code": 1001,
  "message": "Invalid parameter"
}
```

## 接口列表

### 1. 健康检查

**接口**: `GET /health`

**说明**: 检查服务是否正常运行

**请求示例**:

```bash
curl http://localhost:8080/health
```

**响应示例**:

```json
{
  "status": "ok"
}
```

---

### 2. 创建支付

**接口**: `POST /api/v1/payment/create`

**说明**: 创建支付订单

**认证**: 需要

**请求头**:

```
X-API-Key: your_api_key
Content-Type: application/json
```

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| provider | string | 是 | 支付提供商：alipay/wechat/stripe/paypal |
| out_trade_no | string | 是 | 商户订单号，唯一标识 |
| subject | string | 是 | 订单标题 |
| body | string | 否 | 订单描述 |
| amount | float | 是 | 订单金额，必须大于0 |
| currency | string | 否 | 货币类型，默认CNY |
| notify_url | string | 否 | 异步通知URL |
| return_url | string | 否 | 同步跳转URL |
| extra_params | object | 否 | 额外参数 |

**请求示例**:

```bash
curl -X POST http://localhost:8080/api/v1/payment/create \
  -H "X-API-Key: ak_test_1234567890abcdef1234567890abcdef" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "alipay",
    "out_trade_no": "ORDER_20240101_001",
    "subject": "测试商品",
    "body": "这是一个测试订单",
    "amount": 0.01,
    "currency": "CNY",
    "notify_url": "https://your-domain.com/callback",
    "return_url": "https://your-domain.com/return"
  }'
```

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "order_no": "UNI20240101120000abcd1234",
    "payment_url": "https://openapi.alipay.com/gateway.do?...",
    "payment_id": "ORDER_20240101_001",
    "qr_code": "",
    "extra_data": {}
  }
}
```

**响应字段说明**:

| 字段名 | 类型 | 说明 |
|--------|------|------|
| order_no | string | 系统订单号 |
| payment_url | string | 支付链接（跳转支付） |
| payment_id | string | 支付ID |
| qr_code | string | 二维码内容（扫码支付） |
| extra_data | object | 额外数据 |

---

### 3. 查询支付

**接口**: `GET /api/v1/payment/query/:order_no`

**说明**: 查询支付订单状态

**认证**: 需要

**请求头**:

```
X-API-Key: your_api_key
```

**路径参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| order_no | string | 是 | 系统订单号 |

**请求示例**:

```bash
curl -X GET http://localhost:8080/api/v1/payment/query/UNI20240101120000abcd1234 \
  -H "X-API-Key: ak_test_1234567890abcdef1234567890abcdef"
```

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "order_no": "UNI20240101120000abcd1234",
    "user_id": 1,
    "provider": "alipay",
    "config_id": 1,
    "out_trade_no": "ORDER_20240101_001",
    "trade_no": "2024010122001234567890",
    "subject": "测试商品",
    "body": "这是一个测试订单",
    "amount": 0.01,
    "currency": "CNY",
    "status": "success",
    "notify_url": "https://your-domain.com/callback",
    "return_url": "https://your-domain.com/return",
    "client_ip": "127.0.0.1",
    "payment_time": "2024-01-01T12:00:00Z",
    "created_at": "2024-01-01T11:55:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

**订单状态说明**:

| 状态值 | 说明 |
|--------|------|
| pending | 待支付 |
| processing | 处理中 |
| success | 支付成功 |
| failed | 支付失败 |
| closed | 已关闭 |

---

### 4. 支付通知回调

**接口**: `POST /api/v1/public/notify/:provider`

**说明**: 接收支付平台的异步通知（由支付平台调用）

**认证**: 不需要

**路径参数**:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| provider | string | 是 | 支付提供商：alipay/wechat/stripe/paypal |

**说明**:

- 该接口由第三方支付平台调用
- 不同支付平台的通知格式不同
- 系统会自动验证签名并更新订单状态
- 如果商户配置了 notify_url，系统会将通知转发给商户

**支付宝通知示例**:

```
POST /api/v1/public/notify/alipay
Content-Type: application/x-www-form-urlencoded

notify_time=2024-01-01+12:00:00&
notify_type=trade_status_sync&
notify_id=xxx&
app_id=xxx&
charset=utf-8&
version=1.0&
sign_type=RSA2&
sign=xxx&
trade_no=2024010122001234567890&
out_trade_no=ORDER_20240101_001&
trade_status=TRADE_SUCCESS&
total_amount=0.01
```

---

## 支付流程

### 完整支付流程

```
1. 商户调用「创建支付」接口
   ↓
2. 系统创建订单并调用支付平台API
   ↓
3. 系统返回支付链接/二维码给商户
   ↓
4. 用户完成支付
   ↓
5. 支付平台发送异步通知到系统
   ↓
6. 系统验证通知并更新订单状态
   ↓
7. 系统将通知转发给商户（如果配置了notify_url）
   ↓
8. 商户可通过「查询支付」接口主动查询订单状态
```

### 建议的集成方式

1. **创建支付订单**
   ```javascript
   fetch('http://localhost:8080/api/v1/payment/create', {
     method: 'POST',
     headers: {
       'X-API-Key': 'your_api_key',
       'Content-Type': 'application/json'
     },
     body: JSON.stringify({
       provider: 'alipay',
       out_trade_no: 'ORDER_' + Date.now(),
       subject: '商品标题',
       amount: 0.01,
       notify_url: 'https://your-domain.com/notify'
     })
   })
   .then(res => res.json())
   .then(data => {
     // 跳转到支付链接
     window.location.href = data.data.payment_url;
   });
   ```

2. **处理支付通知**
   ```javascript
   // 商户服务端接收通知
   app.post('/notify', (req, res) => {
     const { order_no, status } = req.body;

     // 验证通知的真实性（建议）
     // 更新本地订单状态

     if (status === 'success') {
       // 支付成功，处理业务逻辑
     }

     // 返回成功响应
     res.json({ code: 0, message: 'success' });
   });
   ```

3. **主动查询订单状态**
   ```javascript
   // 定时查询订单状态
   setInterval(() => {
     fetch('http://localhost:8080/api/v1/payment/query/' + order_no, {
       headers: {
         'X-API-Key': 'your_api_key'
       }
     })
     .then(res => res.json())
     .then(data => {
       if (data.data.status === 'success') {
         // 支付成功
       }
     });
   }, 3000);
   ```

## 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1000 | 内部服务器错误 |
| 1001 | 参数错误 |
| 1002 | 未授权（API Key无效） |
| 1003 | 禁止访问（用户被禁用） |
| 1004 | 资源未找到 |
| 1005 | 冲突 |
| 1006 | 请求过于频繁 |
| 2000 | 创建支付失败 |
| 2001 | 查询支付失败 |
| 2002 | 处理支付通知失败 |
| 2003 | 退款失败 |
| 2004 | 取消支付失败 |
| 2005 | 支付提供商未找到 |
| 2006 | 支付配置未找到 |
| 2007 | 订单未找到 |
| 2008 | 订单状态错误 |
| 2009 | 金额无效 |

## 注意事项

1. **API Key 安全**
   - 不要在前端代码中暴露 API Key
   - 定期更换 API Key
   - 使用 HTTPS 传输

2. **订单号唯一性**
   - `out_trade_no` 必须在商户系统中唯一
   - 建议格式：`前缀_日期_随机数`

3. **金额精度**
   - 金额使用浮点数，保留两位小数
   - 人民币最小单位为分（0.01元）

4. **异步通知**
   - 必须正确处理支付平台的异步通知
   - 通知可能会重复发送，需要做幂等性处理
   - 返回正确的响应格式，否则支付平台会重复通知

5. **主动查询**
   - 建议在支付页面定时查询订单状态
   - 不要过于频繁查询，建议间隔3-5秒

6. **超时处理**
   - 订单创建后如果长时间未支付，建议关闭订单
   - 可以通过定时任务自动关闭超时订单

## 测试建议

1. **使用沙箱环境**
   - 支付宝：使用沙箱应用和沙箱账号
   - 微信：使用测试商户号
   - Stripe：使用测试密钥
   - PayPal：使用 sandbox 模式

2. **测试用例**
   - 正常支付流程
   - 支付失败场景
   - 网络超时场景
   - 重复通知处理
   - 订单状态查询

## 支持的支付方式

### 支付宝

- 网页支付（PC）
- 手机网站支付（H5）
- APP 支付（需要额外配置）

### 微信支付

- 扫码支付（Native）
- 公众号支付（JSAPI，需要额外配置）
- H5 支付（需要额外配置）

### Stripe

- Checkout Session（网页支付）
- 支持信用卡支付

### PayPal

- 标准支付流程
- 支持多种货币

## 常见问题

**Q: 如何获取 API Key？**

A: 目前需要直接在数据库中创建用户并获取 API Key。后续版本会提供用户注册接口。

**Q: 支持退款吗？**

A: 支持。退款接口正在开发中，敬请期待。

**Q: 如何测试支付功能？**

A: 建议使用各支付平台提供的沙箱环境进行测试。

**Q: 系统如何保证支付安全？**

A: 系统会验证支付通知的签名，确保通知来自官方支付平台。同时建议商户也要验证金额等关键信息。

**Q: 支持哪些货币？**

A: 支持所有主流货币，具体取决于选择的支付平台。
