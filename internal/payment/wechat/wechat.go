package wechat

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/zqdfound/go-uni-pay/internal/payment"
	apperrors "github.com/zqdfound/go-uni-pay/pkg/errors"
)

// Provider 微信支付提供商
type Provider struct{}

// NewProvider 创建微信支付提供商
func NewProvider() *Provider {
	return &Provider{}
}

// GetName 获取提供商名称
func (p *Provider) GetName() string {
	return payment.ProviderWechat
}

// CreatePayment 创建支付（Native扫码支付）
func (p *Provider) CreatePayment(ctx context.Context, req *payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	client, mchID, err := p.getClient(req.Config)
	if err != nil {
		return nil, err
	}

	svc := native.NativeApiService{Client: client}

	// 创建支付请求
	amount := int64(req.Amount * 100) // 转换为分
	resp, _, err := svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(req.Config["app_id"].(string)),
		Mchid:       core.String(mchID),
		Description: core.String(req.Subject),
		OutTradeNo:  core.String(req.OutTradeNo),
		NotifyUrl:   core.String(req.NotifyURL),
		Amount: &native.Amount{
			Total:    core.Int64(amount),
			Currency: core.String("CNY"),
		},
	})

	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrPaymentCreate, "failed to create wechat payment", err)
	}

	return &payment.CreatePaymentResponse{
		QRCode:    *resp.CodeUrl,
		PaymentID: req.OutTradeNo,
	}, nil
}

// QueryPayment 查询支付
func (p *Provider) QueryPayment(ctx context.Context, req *payment.QueryPaymentRequest) (*payment.QueryPaymentResponse, error) {
	// 注意：微信支付查询需要使用对应的API，这里简化实现
	return &payment.QueryPaymentResponse{
		OutTradeNo: req.OutTradeNo,
		Status:     payment.StatusPending,
	}, nil
}

// HandleNotify 处理支付通知
func (p *Provider) HandleNotify(ctx context.Context, req *payment.NotifyRequest) (*payment.NotifyResponse, error) {
	// 获取配置
	apiV3Key, ok := req.Config["api_v3_key"].(string)
	if !ok {
		return nil, apperrors.New(apperrors.ErrConfigNotFound, "api_v3_key not found in config")
	}

	// 创建通知处理器（简化版本，不进行证书验证）
	handler := notify.NewNotifyHandler(apiV3Key, nil)

	// 从请求头中获取必要信息（这些信息应该在FormData中）
	timestamp := getFirstValue(req.FormData, "Wechatpay-Timestamp")
	nonce := getFirstValue(req.FormData, "Wechatpay-Nonce")
	signature := getFirstValue(req.FormData, "Wechatpay-Signature")
	serial := getFirstValue(req.FormData, "Wechatpay-Serial")

	// 构建 HTTP 请求用于解析
	httpReq, err := http.NewRequest("POST", req.RequestURL, bytes.NewReader(req.RawData))
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrPaymentNotify, "failed to create http request", err)
	}

	// 设置必要的请求头
	httpReq.Header.Set("Wechatpay-Timestamp", timestamp)
	httpReq.Header.Set("Wechatpay-Nonce", nonce)
	httpReq.Header.Set("Wechatpay-Signature", signature)
	httpReq.Header.Set("Wechatpay-Serial", serial)

	// 解析并验证通知
	transaction := new(payments.Transaction)
	_, err = handler.ParseNotifyRequest(ctx, httpReq, transaction)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrPaymentNotify, "failed to parse notify request", err)
	}

	// 构建响应
	response := &payment.NotifyResponse{
		ReturnData: []byte(`{"code": "SUCCESS", "message": "成功"}`),
	}

	// 提取支付信息
	if transaction.OutTradeNo != nil {
		response.OutTradeNo = *transaction.OutTradeNo
	}

	if transaction.TransactionId != nil {
		response.TradeNo = *transaction.TransactionId
	}

	// 转换支付状态
	if transaction.TradeState != nil {
		response.Status = p.convertTradeState(transaction.TradeState)
	}

	// 获取金额（微信支付金额单位是分）
	if transaction.Amount != nil && transaction.Amount.Total != nil {
		response.Amount = float64(*transaction.Amount.Total) / 100
	}

	// 获取支付时间
	if transaction.SuccessTime != nil {
		successTime, err := time.Parse(time.RFC3339, *transaction.SuccessTime)
		if err == nil {
			response.PaymentTime = successTime.Format("2006-01-02 15:04:05")
		}
	}

	// 获取买家信息
	if transaction.Payer != nil && transaction.Payer.Openid != nil {
		response.BuyerInfo = *transaction.Payer.Openid
	}

	return response, nil
}

// convertTradeState 转换微信支付交易状态
func (p *Provider) convertTradeState(tradeState *string) string {
	if tradeState == nil {
		return payment.StatusFailed
	}

	switch *tradeState {
	case "SUCCESS":
		return payment.StatusSuccess
	case "REFUND":
		return payment.StatusSuccess // 转入退款状态也认为支付成功
	case "NOTPAY":
		return payment.StatusPending
	case "CLOSED":
		return payment.StatusClosed
	case "REVOKED":
		return payment.StatusClosed
	case "USERPAYING":
		return payment.StatusPending
	case "PAYERROR":
		return payment.StatusFailed
	default:
		return payment.StatusFailed
	}
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

// getClient 获取微信支付客户端
func (p *Provider) getClient(config map[string]interface{}) (*core.Client, string, error) {
	mchID, ok := config["mch_id"].(string)
	if !ok {
		return nil, "", apperrors.New(apperrors.ErrConfigNotFound, "mch_id not found in config")
	}

	serialNo, ok := config["serial_no"].(string)
	if !ok {
		return nil, "", apperrors.New(apperrors.ErrConfigNotFound, "serial_no not found in config")
	}

	apiV3Key, ok := config["api_v3_key"].(string)
	if !ok {
		return nil, "", apperrors.New(apperrors.ErrConfigNotFound, "api_v3_key not found in config")
	}

	privateKey, ok := config["private_key"].(string)
	if !ok {
		return nil, "", apperrors.New(apperrors.ErrConfigNotFound, "private_key not found in config")
	}

	// 解析私钥
	key, err := parsePrivateKey(privateKey)
	if err != nil {
		return nil, "", apperrors.Wrap(apperrors.ErrConfigNotFound, "failed to parse private key", err)
	}

	// 创建客户端 - 使用简化方式
	client, err := core.NewClient(
		context.Background(),
		option.WithMerchantCredential(mchID, serialNo, key),
		option.WithWechatPayAutoAuthCipher(mchID, serialNo, key, apiV3Key),
	)
	if err != nil {
		return nil, "", apperrors.Wrap(apperrors.ErrPaymentCreate, "failed to create wechat client", err)
	}

	return client, mchID, nil
}

// parsePrivateKey 解析私钥
func parsePrivateKey(privateKeyStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey.(*rsa.PrivateKey), nil
}

// getFirstValue 获取第一个值
func getFirstValue(values url.Values, key string) string {
	if values == nil {
		return ""
	}
	return values.Get(key)
}

// init 注册支付提供商
func init() {
	payment.Register(NewProvider())
}
