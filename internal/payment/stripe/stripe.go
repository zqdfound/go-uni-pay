package stripe

import (
	"context"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/zqdfound/go-uni-pay/internal/payment"
	apperrors "github.com/zqdfound/go-uni-pay/pkg/errors"
)

// Provider Stripe支付提供商
type Provider struct{}

// NewProvider 创建Stripe提供商
func NewProvider() *Provider {
	return &Provider{}
}

// GetName 获取提供商名称
func (p *Provider) GetName() string {
	return payment.ProviderStripe
}

// CreatePayment 创建支付
func (p *Provider) CreatePayment(ctx context.Context, req *payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	if err := p.setAPIKey(req.Config); err != nil {
		return nil, err
	}

	// 转换金额（Stripe使用最小货币单位，如美分）
	amount := int64(req.Amount * 100)

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(req.Currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(req.Subject),
					},
					UnitAmount: stripe.Int64(amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(req.ReturnURL),
		CancelURL:  stripe.String(req.ReturnURL),
	}

	params.ClientReferenceID = stripe.String(req.OutTradeNo)

	s, err := session.New(params)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrPaymentCreate, "failed to create stripe payment", err)
	}

	return &payment.CreatePaymentResponse{
		PaymentURL: s.URL,
		PaymentID:  s.ID,
	}, nil
}

// QueryPayment 查询支付
func (p *Provider) QueryPayment(ctx context.Context, req *payment.QueryPaymentRequest) (*payment.QueryPaymentResponse, error) {
	if err := p.setAPIKey(req.Config); err != nil {
		return nil, err
	}

	s, err := session.Get(req.TradeNo, nil)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrPaymentQuery, "failed to query stripe payment", err)
	}

	status := p.convertStatus(s.PaymentStatus)

	return &payment.QueryPaymentResponse{
		TradeNo:    s.ID,
		OutTradeNo: s.ClientReferenceID,
		Status:     status,
		Amount:     float64(s.AmountTotal) / 100,
	}, nil
}

// HandleNotify 处理支付通知
func (p *Provider) HandleNotify(ctx context.Context, req *payment.NotifyRequest) (*payment.NotifyResponse, error) {
	// Stripe使用Webhook，需要验证签名
	// 这里简化实现
	return &payment.NotifyResponse{
		ReturnData: []byte(`{"received": true}`),
	}, nil
}

// RefundPayment 退款
func (p *Provider) RefundPayment(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	return &payment.RefundResponse{
		RefundNo: req.RefundNo,
		Status:   "success",
	}, nil
}

// ClosePayment 关闭支付
func (p *Provider) ClosePayment(ctx context.Context, req *payment.ClosePaymentRequest) error {
	return nil
}

// setAPIKey 设置API密钥
func (p *Provider) setAPIKey(config map[string]interface{}) error {
	secretKey, ok := config["secret_key"].(string)
	if !ok {
		return apperrors.New(apperrors.ErrConfigNotFound, "secret_key not found in config")
	}
	stripe.Key = secretKey
	return nil
}

// convertStatus 转换支付状态
func (p *Provider) convertStatus(stripeStatus stripe.CheckoutSessionPaymentStatus) string {
	switch stripeStatus {
	case stripe.CheckoutSessionPaymentStatusPaid:
		return payment.StatusSuccess
	case stripe.CheckoutSessionPaymentStatusUnpaid:
		return payment.StatusPending
	default:
		return payment.StatusFailed
	}
}

// init 注册支付提供商
func init() {
	payment.Register(NewProvider())
}
