package stripe

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"
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
	if err := p.setAPIKey(req.Config); err != nil {
		return nil, err
	}

	// 验证Webhook签名(可选,取决于配置中是否提供了webhook_secret)
	var event stripe.Event
	if webhookSecret, ok := req.Config["webhook_secret"].(string); ok && webhookSecret != "" {
		// 从请求头中获取签名
		signature := getFirstValue(req.FormData, "Stripe-Signature")
		if signature == "" {
			return nil, apperrors.New(apperrors.ErrPaymentNotify, "missing Stripe-Signature header")
		}

		// 验证签名并构造事件
		var err error
		event, err = webhook.ConstructEvent(req.RawData, signature, webhookSecret)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.ErrPaymentNotify, "failed to verify stripe webhook signature", err)
		}
	} else {
		// 如果没有配置webhook_secret,直接解析事件(不推荐用于生产环境)
		if err := json.Unmarshal(req.RawData, &event); err != nil {
			return nil, apperrors.Wrap(apperrors.ErrPaymentNotify, "failed to parse stripe webhook event", err)
		}
	}

	// 初始化响应
	response := &payment.NotifyResponse{
		ReturnData: []byte(`{"received": true}`),
	}

	// 根据事件类型处理
	switch event.Type {
	case "checkout.session.completed":
		// 结账会话完成
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			return nil, apperrors.Wrap(apperrors.ErrPaymentNotify, "failed to parse checkout.session.completed event", err)
		}

		response.TradeNo = sess.ID
		response.OutTradeNo = sess.ClientReferenceID
		response.Status = p.convertStatus(sess.PaymentStatus)
		response.Amount = float64(sess.AmountTotal) / 100

		// 获取支付时间
		if sess.Created > 0 {
			response.PaymentTime = fmt.Sprintf("%d", sess.Created)
		}

		// 获取买家信息
		if sess.CustomerDetails != nil {
			response.BuyerInfo = sess.CustomerDetails.Email
		}

	case "payment_intent.succeeded":
		// 支付意图成功
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return nil, apperrors.Wrap(apperrors.ErrPaymentNotify, "failed to parse payment_intent.succeeded event", err)
		}

		response.TradeNo = pi.ID
		response.Status = payment.StatusSuccess
		response.Amount = float64(pi.Amount) / 100

		// 获取支付时间
		if pi.Created > 0 {
			response.PaymentTime = fmt.Sprintf("%d", pi.Created)
		}

		// 从metadata中获取商户订单号(如果有)
		if pi.Metadata != nil {
			if outTradeNo, ok := pi.Metadata["out_trade_no"]; ok {
				response.OutTradeNo = outTradeNo
			}
		}

		// 获取买家信息
		if pi.ReceiptEmail != "" {
			response.BuyerInfo = pi.ReceiptEmail
		}

	case "payment_intent.payment_failed":
		// 支付意图失败
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return nil, apperrors.Wrap(apperrors.ErrPaymentNotify, "failed to parse payment_intent.payment_failed event", err)
		}

		response.TradeNo = pi.ID
		response.Status = payment.StatusFailed
		response.Amount = float64(pi.Amount) / 100

		// 从metadata中获取商户订单号(如果有)
		if pi.Metadata != nil {
			if outTradeNo, ok := pi.Metadata["out_trade_no"]; ok {
				response.OutTradeNo = outTradeNo
			}
		}

	case "charge.succeeded":
		// 收费成功
		var charge stripe.Charge
		if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
			return nil, apperrors.Wrap(apperrors.ErrPaymentNotify, "failed to parse charge.succeeded event", err)
		}

		response.TradeNo = charge.ID
		response.Status = payment.StatusSuccess
		response.Amount = float64(charge.Amount) / 100

		// 获取支付时间
		if charge.Created > 0 {
			response.PaymentTime = fmt.Sprintf("%d", charge.Created)
		}

		// 获取买家信息
		if charge.ReceiptEmail != "" {
			response.BuyerInfo = charge.ReceiptEmail
		}

		// 从metadata中获取商户订单号(如果有)
		if charge.Metadata != nil {
			if outTradeNo, ok := charge.Metadata["out_trade_no"]; ok {
				response.OutTradeNo = outTradeNo
			}
		}

	case "charge.failed":
		// 收费失败
		var charge stripe.Charge
		if err := json.Unmarshal(event.Data.Raw, &charge); err != nil {
			return nil, apperrors.Wrap(apperrors.ErrPaymentNotify, "failed to parse charge.failed event", err)
		}

		response.TradeNo = charge.ID
		response.Status = payment.StatusFailed
		response.Amount = float64(charge.Amount) / 100

		// 从metadata中获取商户订单号(如果有)
		if charge.Metadata != nil {
			if outTradeNo, ok := charge.Metadata["out_trade_no"]; ok {
				response.OutTradeNo = outTradeNo
			}
		}

	default:
		// 其他事件类型,只记录但不更新状态
		response.Status = ""
	}

	return response, nil
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

// getFirstValue 从表单数据中获取第一个值
func getFirstValue(formData map[string][]string, key string) string {
	if values, ok := formData[key]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}

// init 注册支付提供商
func init() {
	payment.Register(NewProvider())
}
