package wechat

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
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
	// 解析通知（需要根据微信支付的通知格式进行解析）
	// 这里简化实现，实际需要验签和解密
	return &payment.NotifyResponse{
		ReturnData: []byte(`{"code": "SUCCESS", "message": "成功"}`),
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

// init 注册支付提供商
func init() {
	payment.Register(NewProvider())
}
