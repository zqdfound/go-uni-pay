package alipay

import (
	"context"
	"fmt"

	"github.com/smartwalle/alipay/v3"
	"github.com/zqdfound/go-uni-pay/internal/payment"
	apperrors "github.com/zqdfound/go-uni-pay/pkg/errors"
)

// Provider 支付宝支付提供商
type Provider struct{}

// NewProvider 创建支付宝提供商
func NewProvider() *Provider {
	return &Provider{}
}

// GetName 获取提供商名称
func (p *Provider) GetName() string {
	return payment.ProviderAlipay
}

// CreatePayment 创建支付
func (p *Provider) CreatePayment(ctx context.Context, req *payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	client, err := p.getClient(req.Config)
	if err != nil {
		return nil, err
	}

	// 创建支付请求
	var pay = alipay.TradePagePay{}
	pay.NotifyURL = req.NotifyURL
	pay.ReturnURL = req.ReturnURL
	pay.Subject = req.Subject
	pay.OutTradeNo = req.OutTradeNo
	pay.TotalAmount = fmt.Sprintf("%.2f", req.Amount)
	pay.ProductCode = "FAST_INSTANT_TRADE_PAY"

	// 生成支付URL
	url, err := client.TradePagePay(pay)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrPaymentCreate, "failed to create alipay payment", err)
	}

	return &payment.CreatePaymentResponse{
		PaymentURL: url.String(),
		PaymentID:  req.OutTradeNo,
	}, nil
}

// QueryPayment 查询支付
func (p *Provider) QueryPayment(ctx context.Context, req *payment.QueryPaymentRequest) (*payment.QueryPaymentResponse, error) {
	client, err := p.getClient(req.Config)
	if err != nil {
		return nil, err
	}

	// 查询支付
	var query = alipay.TradeQuery{}
	if req.OutTradeNo != "" {
		query.OutTradeNo = req.OutTradeNo
	}
	if req.TradeNo != "" {
		query.TradeNo = req.TradeNo
	}

	rsp, err := client.TradeQuery(query)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrPaymentQuery, "failed to query alipay payment", err)
	}

	if rsp.IsFailure() {
		return nil, apperrors.New(apperrors.ErrPaymentQuery, rsp.Msg)
	}

	// 转换支付状态
	status := p.convertStatus(string(rsp.TradeStatus))

	return &payment.QueryPaymentResponse{
		TradeNo:     rsp.TradeNo,
		OutTradeNo:  rsp.OutTradeNo,
		Status:      status,
		Amount:      parseAmount(rsp.TotalAmount),
		PaymentTime: rsp.SendPayDate,
		BuyerInfo:   rsp.BuyerLogonId,
	}, nil
}

// HandleNotify 处理支付通知
func (p *Provider) HandleNotify(ctx context.Context, req *payment.NotifyRequest) (*payment.NotifyResponse, error) {
	client, err := p.getClient(req.Config)
	if err != nil {
		return nil, err
	}

	// 解析通知
	notification, err := client.DecodeNotification(req.FormData)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrPaymentNotify, "failed to decode alipay notification", err)
	}

	// 转换支付状态
	status := p.convertStatus(string(notification.TradeStatus))

	response := &payment.NotifyResponse{
		TradeNo:     notification.TradeNo,
		OutTradeNo:  notification.OutTradeNo,
		Status:      status,
		Amount:      parseAmount(notification.TotalAmount),
		PaymentTime: notification.GmtPayment,
		BuyerInfo:   notification.BuyerLogonId,
		ReturnData:  []byte("success"),
	}

	return response, nil
}

// RefundPayment 退款
func (p *Provider) RefundPayment(ctx context.Context, req *payment.RefundRequest) (*payment.RefundResponse, error) {
	client, err := p.getClient(req.Config)
	if err != nil {
		return nil, err
	}

	// 创建退款请求
	var refund = alipay.TradeRefund{}
	if req.OutTradeNo != "" {
		refund.OutTradeNo = req.OutTradeNo
	}
	if req.TradeNo != "" {
		refund.TradeNo = req.TradeNo
	}
	refund.RefundAmount = fmt.Sprintf("%.2f", req.RefundAmount)
	refund.OutRequestNo = req.RefundNo
	refund.RefundReason = req.Reason

	rsp, err := client.TradeRefund(refund)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrPaymentRefund, "failed to refund alipay payment", err)
	}

	if rsp.IsFailure() {
		return nil, apperrors.New(apperrors.ErrPaymentRefund, rsp.Msg)
	}

	return &payment.RefundResponse{
		RefundNo: req.RefundNo,
		TradeNo:  rsp.TradeNo,
		Status:   "success",
	}, nil
}

// ClosePayment 关闭支付
func (p *Provider) ClosePayment(ctx context.Context, req *payment.ClosePaymentRequest) error {
	client, err := p.getClient(req.Config)
	if err != nil {
		return err
	}

	// 关闭支付
	var close = alipay.TradeClose{}
	if req.OutTradeNo != "" {
		close.OutTradeNo = req.OutTradeNo
	}
	if req.TradeNo != "" {
		close.TradeNo = req.TradeNo
	}

	rsp, err := client.TradeClose(close)
	if err != nil {
		return apperrors.Wrap(apperrors.ErrPaymentCancel, "failed to close alipay payment", err)
	}

	if rsp.IsFailure() {
		return apperrors.New(apperrors.ErrPaymentCancel, rsp.Msg)
	}

	return nil
}

// getClient 获取支付宝客户端
func (p *Provider) getClient(config map[string]interface{}) (*alipay.Client, error) {
	appID, ok := config["app_id"].(string)
	if !ok {
		return nil, apperrors.New(apperrors.ErrConfigNotFound, "app_id not found in config")
	}

	privateKey, ok := config["private_key"].(string)
	if !ok {
		return nil, apperrors.New(apperrors.ErrConfigNotFound, "private_key not found in config")
	}

	publicKey, ok := config["public_key"].(string)
	if !ok {
		publicKey = "" // 可选
	}

	isProduction, _ := config["is_production"].(bool)

	// 创建客户端
	client, err := alipay.New(appID, privateKey, isProduction)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrPaymentCreate, "failed to create alipay client", err)
	}

	// 加载支付宝公钥
	if publicKey != "" {
		if err := client.LoadAliPayPublicKey(publicKey); err != nil {
			return nil, apperrors.Wrap(apperrors.ErrPaymentCreate, "failed to load alipay public key", err)
		}
	}

	return client, nil
}

// convertStatus 转换支付状态
func (p *Provider) convertStatus(alipayStatus string) string {
	switch alipayStatus {
	case "WAIT_BUYER_PAY":
		return payment.StatusPending
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		return payment.StatusSuccess
	case "TRADE_CLOSED":
		return payment.StatusClosed
	default:
		return payment.StatusFailed
	}
}

// parseAmount 解析金额
func parseAmount(amount string) float64 {
	var result float64
	fmt.Sscanf(amount, "%f", &result)
	return result
}

// init 注册支付提供商
func init() {
	payment.Register(NewProvider())
}
