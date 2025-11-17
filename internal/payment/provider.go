package payment

import (
	"context"
)

// Provider 支付提供商接口
type Provider interface {
	// GetName 获取提供商名称
	GetName() string

	// CreatePayment 创建支付
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error)

	// QueryPayment 查询支付
	QueryPayment(ctx context.Context, req *QueryPaymentRequest) (*QueryPaymentResponse, error)

	// HandleNotify 处理支付通知
	HandleNotify(ctx context.Context, req *NotifyRequest) (*NotifyResponse, error)

	// RefundPayment 退款
	RefundPayment(ctx context.Context, req *RefundRequest) (*RefundResponse, error)

	// ClosePayment 关闭支付
	ClosePayment(ctx context.Context, req *ClosePaymentRequest) error
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	OutTradeNo  string                 // 商户订单号
	Subject     string                 // 订单标题
	Body        string                 // 订单描述
	Amount      float64                // 订单金额
	Currency    string                 // 货币类型
	NotifyURL   string                 // 异步通知URL
	ReturnURL   string                 // 同步跳转URL
	ClientIP    string                 // 客户端IP
	Config      map[string]interface{} // 支付配置
	ExtraParams map[string]interface{} // 额外参数
}

// CreatePaymentResponse 创建支付响应
type CreatePaymentResponse struct {
	PaymentURL string                 // 支付链接（如果是跳转支付）
	PaymentID  string                 // 支付ID
	TradeNo    string                 // 第三方交易号
	QRCode     string                 // 二维码内容（如果是扫码支付）
	FormData   string                 // 表单数据（如果是表单支付）
	ExtraData  map[string]interface{} // 额外数据
}

// QueryPaymentRequest 查询支付请求
type QueryPaymentRequest struct {
	OutTradeNo string                 // 商户订单号
	TradeNo    string                 // 第三方交易号
	Config     map[string]interface{} // 支付配置
}

// QueryPaymentResponse 查询支付响应
type QueryPaymentResponse struct {
	TradeNo     string  // 第三方交易号
	OutTradeNo  string  // 商户订单号
	Status      string  // 支付状态：pending/success/failed/closed
	Amount      float64 // 订单金额
	PaymentTime string  // 支付时间
	BuyerInfo   string  // 买家信息
}

// NotifyRequest 通知请求
type NotifyRequest struct {
	RawData    []byte                 // 原始数据
	FormData   map[string][]string    // 表单数据
	Config     map[string]interface{} // 支付配置
	RequestURL string                 // 请求URL
}

// NotifyResponse 通知响应
type NotifyResponse struct {
	TradeNo     string  // 第三方交易号
	OutTradeNo  string  // 商户订单号
	Status      string  // 支付状态
	Amount      float64 // 订单金额
	PaymentTime string  // 支付时间
	BuyerInfo   string  // 买家信息
	ReturnData  []byte  // 返回给第三方的数据
}

// RefundRequest 退款请求
type RefundRequest struct {
	OutTradeNo   string                 // 商户订单号
	TradeNo      string                 // 第三方交易号
	RefundNo     string                 // 退款单号
	RefundAmount float64                // 退款金额
	TotalAmount  float64                // 订单总金额
	Reason       string                 // 退款原因
	Config       map[string]interface{} // 支付配置
}

// RefundResponse 退款响应
type RefundResponse struct {
	RefundNo string // 退款单号
	TradeNo  string // 第三方交易号
	Status   string // 退款状态
}

// ClosePaymentRequest 关闭支付请求
type ClosePaymentRequest struct {
	OutTradeNo string                 // 商户订单号
	TradeNo    string                 // 第三方交易号
	Config     map[string]interface{} // 支付配置
}

// PaymentStatus 支付状态
const (
	StatusPending = "pending"
	StatusSuccess = "success"
	StatusFailed  = "failed"
	StatusClosed  = "closed"
)

// ProviderName 提供商名称
const (
	ProviderAlipay = "alipay"
	ProviderWechat = "wechat"
	ProviderStripe = "stripe"
	ProviderPayPal = "paypal"
)
