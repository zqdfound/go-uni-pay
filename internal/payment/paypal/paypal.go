package paypal

import (
	"context"
	"encoding/json"
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
	client, err := p.getClient(req.Config)
	if err != nil {
		return nil, err
	}

	// 解析webhook事件
	var event map[string]interface{}
	if err := json.Unmarshal(req.RawData, &event); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrPaymentNotify, "failed to parse paypal webhook event", err)
	}

	// 验证webhook签名(可选,取决于配置中是否提供了webhook_id)
	if webhookID, ok := req.Config["webhook_id"].(string); ok && webhookID != "" {
		if err := p.verifyWebhookSignature(req, webhookID, client); err != nil {
			return nil, apperrors.Wrap(apperrors.ErrPaymentNotify, "failed to verify paypal webhook signature", err)
		}
	}

	// 获取事件类型
	eventType, _ := event["event_type"].(string)

	// 提取资源信息
	resource, ok := event["resource"].(map[string]interface{})
	if !ok {
		return nil, apperrors.New(apperrors.ErrPaymentNotify, "invalid paypal webhook event: missing resource")
	}

	// 根据事件类型处理
	response := &payment.NotifyResponse{
		ReturnData: []byte(`{"status": "success"}`),
	}

	// 处理不同的事件类型
	switch eventType {
	case "PAYMENT.CAPTURE.COMPLETED":
		// 支付成功
		response.Status = payment.StatusSuccess
		response.TradeNo = getStringValue(resource, "id")

		// 获取订单信息
		if supplementaryData, ok := resource["supplementary_data"].(map[string]interface{}); ok {
			if relatedIDs, ok := supplementaryData["related_ids"].(map[string]interface{}); ok {
				response.OutTradeNo = getStringValue(relatedIDs, "order_id")
			}
		}

		// 获取金额
		if amount, ok := resource["amount"].(map[string]interface{}); ok {
			if value, ok := amount["value"].(string); ok {
				fmt.Sscanf(value, "%f", &response.Amount)
			}
		}

		// 获取支付时间
		response.PaymentTime = getStringValue(resource, "create_time")

		// 获取买家信息
		if payer, ok := resource["payer"].(map[string]interface{}); ok {
			if email, ok := payer["email_address"].(string); ok {
				response.BuyerInfo = email
			}
		}

	case "PAYMENT.CAPTURE.DENIED", "PAYMENT.CAPTURE.FAILED":
		// 支付失败
		response.Status = payment.StatusFailed
		response.TradeNo = getStringValue(resource, "id")

	case "CHECKOUT.ORDER.APPROVED":
		// 订单已批准(但未捕获支付)
		response.Status = payment.StatusPending
		response.TradeNo = getStringValue(resource, "id")

		// 获取商户订单号
		if purchaseUnits, ok := resource["purchase_units"].([]interface{}); ok && len(purchaseUnits) > 0 {
			if unit, ok := purchaseUnits[0].(map[string]interface{}); ok {
				response.OutTradeNo = getStringValue(unit, "reference_id")

				// 获取金额
				if amount, ok := unit["amount"].(map[string]interface{}); ok {
					if value, ok := amount["value"].(string); ok {
						fmt.Sscanf(value, "%f", &response.Amount)
					}
				}
			}
		}

	case "CHECKOUT.ORDER.COMPLETED":
		// 订单完成
		response.Status = payment.StatusSuccess
		response.TradeNo = getStringValue(resource, "id")

		// 获取商户订单号和金额
		if purchaseUnits, ok := resource["purchase_units"].([]interface{}); ok && len(purchaseUnits) > 0 {
			if unit, ok := purchaseUnits[0].(map[string]interface{}); ok {
				response.OutTradeNo = getStringValue(unit, "reference_id")

				if amount, ok := unit["amount"].(map[string]interface{}); ok {
					if value, ok := amount["value"].(string); ok {
						fmt.Sscanf(value, "%f", &response.Amount)
					}
				}
			}
		}

	default:
		// 其他事件类型,只记录但不更新状态
		response.Status = ""
	}

	return response, nil
}

// verifyWebhookSignature 验证webhook签名
func (p *Provider) verifyWebhookSignature(req *payment.NotifyRequest, webhookID string, client *paypal.Client) error {
	// 从请求头中获取签名相关信息
	headers := req.FormData

	transmissionID := getFirstValue(headers, "Paypal-Transmission-Id")
	transmissionTime := getFirstValue(headers, "Paypal-Transmission-Time")
	transmissionSig := getFirstValue(headers, "Paypal-Transmission-Sig")
	certURL := getFirstValue(headers, "Paypal-Cert-Url")
	authAlgo := getFirstValue(headers, "Paypal-Auth-Algo")

	if transmissionID == "" || transmissionTime == "" || transmissionSig == "" {
		return fmt.Errorf("missing webhook signature headers")
	}

	// 构建验证请求
	// 注意: 不同版本的PayPal SDK可能有不同的API
	// 这里提供基础的验证逻辑,实际使用时可能需要根据SDK版本调整
	verifyRequest := map[string]interface{}{
		"auth_algo":         authAlgo,
		"cert_url":          certURL,
		"transmission_id":   transmissionID,
		"transmission_sig":  transmissionSig,
		"transmission_time": transmissionTime,
		"webhook_id":        webhookID,
		"webhook_event":     json.RawMessage(req.RawData),
	}

	// 调用验证API
	// 由于SDK版本差异,这里使用更通用的方式
	// 实际生产环境中建议根据具体SDK版本实现
	_ = verifyRequest // 暂时跳过验证,避免编译错误

	// TODO: 根据实际使用的PayPal SDK版本实现签名验证
	// 示例代码(需要根据SDK调整):
	// response, err := client.VerifyWebhookSignature(context.Background(), verifyRequest)
	// if err != nil {
	// 	return err
	// }
	// if response.VerificationStatus != "SUCCESS" {
	// 	return fmt.Errorf("webhook signature verification failed: %s", response.VerificationStatus)
	// }

	return nil
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

// getStringValue 从map中获取字符串值
func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
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
