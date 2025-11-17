package paypal

import (
	"context"
	"fmt"

	"github.com/plutov/paypal/v4"
	"github.com/zqdfound/go-uni-pay/internal/payment"
	apperrors "github.com/zqdfound/go-uni-pay/pkg/errors"
)

// Provider PayPal支付提供商
type Provider struct{}

// NewProvider 创建PayPal提供商
func NewProvider() *Provider {
	return &Provider{}
}

// GetName 获取提供商名称
func (p *Provider) GetName() string {
	return payment.ProviderPayPal
}

// CreatePayment 创建支付
func (p *Provider) CreatePayment(ctx context.Context, req *payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	client, err := p.getClient(req.Config)
	if err != nil {
		return nil, err
	}

	// 创建订单
	order, err := client.CreateOrder(ctx, paypal.OrderIntentCapture, []paypal.PurchaseUnitRequest{
		{
			ReferenceID: req.OutTradeNo,
			Amount: &paypal.PurchaseUnitAmount{
				Currency: req.Currency,
				Value:    fmt.Sprintf("%.2f", req.Amount),
			},
			Description: req.Subject,
		},
	}, nil, &paypal.ApplicationContext{
		ReturnURL: req.ReturnURL,
		CancelURL: req.ReturnURL,
	})

	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrPaymentCreate, "failed to create paypal order", err)
	}

	// 获取批准链接
	var approveURL string
	for _, link := range order.Links {
		if link.Rel == "approve" {
			approveURL = link.Href
			break
		}
	}

	return &payment.CreatePaymentResponse{
		PaymentURL: approveURL,
		PaymentID:  order.ID,
		TradeNo:    order.ID,
	}, nil
}

// QueryPayment 查询支付
func (p *Provider) QueryPayment(ctx context.Context, req *payment.QueryPaymentRequest) (*payment.QueryPaymentResponse, error) {
	client, err := p.getClient(req.Config)
	if err != nil {
		return nil, err
	}

	order, err := client.GetOrder(ctx, req.TradeNo)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrPaymentQuery, "failed to query paypal order", err)
	}

	status := p.convertStatus(order.Status)

	var amount float64
	if len(order.PurchaseUnits) > 0 {
		fmt.Sscanf(order.PurchaseUnits[0].Amount.Value, "%f", &amount)
	}

	return &payment.QueryPaymentResponse{
		TradeNo:    order.ID,
		OutTradeNo: order.PurchaseUnits[0].ReferenceID,
		Status:     status,
		Amount:     amount,
	}, nil
}

// HandleNotify 处理支付通知
func (p *Provider) HandleNotify(ctx context.Context, req *payment.NotifyRequest) (*payment.NotifyResponse, error) {
	// PayPal使用Webhook进行通知
	// 这里简化实现
	return &payment.NotifyResponse{
		ReturnData: []byte(`{"status": "success"}`),
	}, nil
}

// RefundPayment 退款
func (p *Provider) RefundPayment(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	client, err := p.getClient(req.Config)
	if err != nil {
		return nil, err
	}

	// 首先需要获取capture ID
	// 这里简化处理
	_, err = client.RefundCapture(ctx, req.TradeNo, paypal.RefundCaptureRequest{
		Amount: &paypal.Money{
			Currency: "USD",
			Value:    fmt.Sprintf("%.2f", req.RefundAmount),
		},
	})

	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrPaymentRefund, "failed to refund paypal payment", err)
	}

	return &payment.RefundResponse{
		RefundNo: req.RefundNo,
		Status:   "success",
	}, nil
}

// ClosePayment 关闭支付
func (p *Provider) ClosePayment(ctx context.Context, req *payment.ClosePaymentRequest) error {
	return nil
}

// getClient 获取PayPal客户端
func (p *Provider) getClient(config map[string]interface{}) (*paypal.Client, error) {
	clientID, ok := config["client_id"].(string)
	if !ok {
		return nil, apperrors.New(apperrors.ErrConfigNotFound, "client_id not found in config")
	}

	secret, ok := config["secret"].(string)
	if !ok {
		return nil, apperrors.New(apperrors.ErrConfigNotFound, "secret not found in config")
	}

	mode, _ := config["mode"].(string)
	if mode == "" {
		mode = "sandbox"
	}

	var apiBase string
	if mode == "live" {
		apiBase = paypal.APIBaseLive
	} else {
		apiBase = paypal.APIBaseSandBox
	}

	client, err := paypal.NewClient(clientID, secret, apiBase)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrPaymentCreate, "failed to create paypal client", err)
	}

	// 获取访问令牌
	_, err = client.GetAccessToken(context.Background())
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrPaymentCreate, "failed to get paypal access token", err)
	}

	return client, nil
}

// convertStatus 转换支付状态
func (p *Provider) convertStatus(paypalStatus string) string {
	switch paypalStatus {
	case "CREATED", "SAVED", "APPROVED", "PAYER_ACTION_REQUIRED":
		return payment.StatusPending
	case "COMPLETED":
		return payment.StatusSuccess
	case "VOIDED", "CANCELLED":
		return payment.StatusClosed
	default:
		return payment.StatusFailed
	}
}

// init 注册支付提供商
func init() {
	payment.Register(NewProvider())
}
