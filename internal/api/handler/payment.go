package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/zqdfound/go-uni-pay/internal/payment"
	paymentService "github.com/zqdfound/go-uni-pay/internal/service/payment"
	apperrors "github.com/zqdfound/go-uni-pay/pkg/errors"
)

// PaymentServiceInterface 支付服务接口，用于依赖注入和测试
type PaymentServiceInterface interface {
	CreatePayment(ctx context.Context, req *paymentService.CreatePaymentRequest) (*paymentService.CreatePaymentResponse, error)
	QueryPayment(ctx context.Context, orderNo string) (interface{}, error)
	HandleNotify(ctx context.Context, provider string, req *payment.NotifyRequest) ([]byte, error)
	GetConfigByID(ctx context.Context, configID uint64) (map[string]interface{}, error)
}

// PaymentHandler 支付处理器
type PaymentHandler struct {
	paymentService PaymentServiceInterface
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler(paymentService PaymentServiceInterface) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	Provider    string                 `json:"provider" binding:"required"`
	OutTradeNo  string                 `json:"out_trade_no" binding:"required"`
	Subject     string                 `json:"subject" binding:"required"`
	Body        string                 `json:"body"`
	Amount      float64                `json:"amount" binding:"required,gt=0"`
	Currency    string                 `json:"currency"`
	NotifyURL   string                 `json:"notify_url"`
	ReturnURL   string                 `json:"return_url"`
	ExtraParams map[string]interface{} `json:"extra_params"`
}

// CreatePayment 创建支付
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    apperrors.ErrInvalidParam,
			"message": err.Error(),
		})
		return
	}

	// 获取用户ID
	userID, _ := c.Get("user_id")

	// 设置默认值
	if req.Currency == "" {
		req.Currency = "CNY"
	}

	// 创建支付
	resp, err := h.paymentService.CreatePayment(c.Request.Context(), &paymentService.CreatePaymentRequest{
		UserID:      userID.(uint64),
		Provider:    req.Provider,
		OutTradeNo:  req.OutTradeNo,
		Subject:     req.Subject,
		Body:        req.Body,
		Amount:      req.Amount,
		Currency:    req.Currency,
		NotifyURL:   req.NotifyURL,
		ReturnURL:   req.ReturnURL,
		ClientIP:    c.ClientIP(),
		ExtraParams: req.ExtraParams,
	})

	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(400, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}

		c.JSON(500, gin.H{
			"code":    apperrors.ErrInternalServer,
			"message": "internal server error",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    apperrors.ErrSuccess,
		"message": "success",
		"data":    resp,
	})
}

// QueryPayment 查询支付
func (h *PaymentHandler) QueryPayment(c *gin.Context) {
	orderNo := c.Param("order_no")
	if orderNo == "" {
		c.JSON(400, gin.H{
			"code":    apperrors.ErrInvalidParam,
			"message": "order_no is required",
		})
		return
	}

	order, err := h.paymentService.QueryPayment(c.Request.Context(), orderNo)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(400, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}

		c.JSON(500, gin.H{
			"code":    apperrors.ErrInternalServer,
			"message": "internal server error",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    apperrors.ErrSuccess,
		"message": "success",
		"data":    order,
	})
}

// HandleNotify 处理支付通知
func (h *PaymentHandler) HandleNotify(c *gin.Context) {
	provider := c.Param("provider")
	configIDStr := c.Param("config_id")

	if provider == "" {
		c.String(400, "provider is required")
		return
	}

	if configIDStr == "" {
		c.String(400, "config_id is required")
		return
	}

	// 解析 config_id
	var configID uint64
	if _, err := fmt.Sscanf(configIDStr, "%d", &configID); err != nil {
		c.String(400, "invalid config_id")
		return
	}

	// 获取支付配置
	config, err := h.paymentService.GetConfigByID(c.Request.Context(), configID)
	if err != nil {
		c.String(400, "config not found")
		return
	}

	// 读取请求数据
	bodyBytes, _ := c.GetRawData()

	// 构造通知请求
	notifyReq := &payment.NotifyRequest{
		RawData:    bodyBytes,
		FormData:   c.Request.Form,
		RequestURL: c.Request.URL.String(),
		Config:     config, // 从数据库获取的配置
	}

	// 处理通知
	returnData, err := h.paymentService.HandleNotify(c.Request.Context(), provider, notifyReq)
	if err != nil {
		c.String(500, "error")
		return
	}

	c.Data(200, "text/plain", returnData)
}
