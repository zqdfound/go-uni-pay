package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/zqdfound/go-uni-pay/internal/payment"
	paymentService "github.com/zqdfound/go-uni-pay/internal/service/payment"
	apperrors "github.com/zqdfound/go-uni-pay/pkg/errors"
)

// PaymentHandler 支付处理器
type PaymentHandler struct {
	paymentService *paymentService.Service
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler(paymentService *paymentService.Service) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	Provider   string                 `json:"provider" binding:"required"`
	OutTradeNo string                 `json:"out_trade_no" binding:"required"`
	Subject    string                 `json:"subject" binding:"required"`
	Body       string                 `json:"body"`
	Amount     float64                `json:"amount" binding:"required,gt=0"`
	Currency   string                 `json:"currency"`
	NotifyURL  string                 `json:"notify_url"`
	ReturnURL  string                 `json:"return_url"`
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
	if provider == "" {
		c.String(400, "provider is required")
		return
	}

	// 读取请求数据
	bodyBytes, _ := c.GetRawData()

	// 构造通知请求
	notifyReq := &payment.NotifyRequest{
		RawData:    bodyBytes,
		FormData:   c.Request.Form,
		RequestURL: c.Request.URL.String(),
		Config:     make(map[string]interface{}), // 需要从数据库获取配置
	}

	// 处理通知
	returnData, err := h.paymentService.HandleNotify(c.Request.Context(), provider, notifyReq)
	if err != nil {
		c.String(500, "error")
		return
	}

	c.Data(200, "text/plain", returnData)
}
