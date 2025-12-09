package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zqdfound/go-uni-pay/internal/domain/entity"
	"github.com/zqdfound/go-uni-pay/internal/domain/repository"
	"github.com/zqdfound/go-uni-pay/internal/payment"
	"github.com/zqdfound/go-uni-pay/pkg/logger"
	"go.uber.org/zap"
)

// NotifyService 通知服务接口
type NotifyService interface {
	AddNotify(ctx context.Context, orderID uint64, orderNo, notifyURL string, notifyData map[string]interface{}) error
}

// Service 支付服务
type Service struct {
	orderRepo     repository.PaymentOrderRepository
	configRepo    repository.PaymentConfigRepository
	logRepo       repository.PaymentLogRepository
	notifyService NotifyService
}

// NewService 创建支付服务
func NewService(
	orderRepo repository.PaymentOrderRepository,
	configRepo repository.PaymentConfigRepository,
	logRepo repository.PaymentLogRepository,
	notifyService NotifyService,
) *Service {
	return &Service{
		orderRepo:     orderRepo,
		configRepo:    configRepo,
		logRepo:       logRepo,
		notifyService: notifyService,
	}
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	UserID      uint64
	Provider    string
	OutTradeNo  string
	Subject     string
	Body        string
	Amount      float64
	Currency    string
	NotifyURL   string
	ReturnURL   string
	ClientIP    string
	ExtraParams map[string]interface{}
}

// CreatePaymentResponse 创建支付响应
type CreatePaymentResponse struct {
	OrderNo    string
	PaymentURL string
	PaymentID  string
	QRCode     string
	ExtraData  map[string]interface{}
}

// CreatePayment 创建支付
func (s *Service) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// 获取支付配置
	config, err := s.configRepo.GetActiveByUserAndProvider(ctx, req.UserID, req.Provider)
	if err != nil {
		return nil, err
	}

	// 生成订单号
	orderNo := s.generateOrderNo()

	// 创建订单记录
	order := &entity.PaymentOrder{
		OrderNo:    orderNo,
		UserID:     req.UserID,
		Provider:   req.Provider,
		ConfigID:   config.ID,
		OutTradeNo: req.OutTradeNo,
		Subject:    req.Subject,
		Body:       req.Body,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Status:     entity.OrderStatusPending,
		NotifyURL:  req.NotifyURL,
		ReturnURL:  req.ReturnURL,
		ClientIP:   req.ClientIP,
		ExtraData:  req.ExtraParams,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// 获取支付提供商
	provider, err := payment.GetProvider(req.Provider)
	if err != nil {
		return nil, err
	}

	// 创建支付请求
	payReq := &payment.CreatePaymentRequest{
		OutTradeNo:  req.OutTradeNo,
		Subject:     req.Subject,
		Body:        req.Body,
		Amount:      req.Amount,
		Currency:    req.Currency,
		NotifyURL:   req.NotifyURL,
		ReturnURL:   req.ReturnURL,
		ClientIP:    req.ClientIP,
		Config:      config.ConfigData,
		ExtraParams: req.ExtraParams,
	}

	// 调用支付提供商创建支付
	payResp, err := provider.CreatePayment(ctx, payReq)
	if err != nil {
		// 记录错误日志
		s.logPayment(ctx, order.ID, orderNo, "create", req.Provider, payReq, nil, "failed", err.Error())

		// 更新订单状态为失败
		order.Status = entity.OrderStatusFailed
		s.orderRepo.Update(ctx, order)

		return nil, err
	}

	// 记录成功日志
	s.logPayment(ctx, order.ID, orderNo, "create", req.Provider, payReq, payResp, "success", "")

	// 更新订单信息
	if payResp.TradeNo != "" {
		order.TradeNo = payResp.TradeNo
		order.Status = entity.OrderStatusProcessing
		s.orderRepo.Update(ctx, order)
	}

	return &CreatePaymentResponse{
		OrderNo:    orderNo,
		PaymentURL: payResp.PaymentURL,
		PaymentID:  payResp.PaymentID,
		QRCode:     payResp.QRCode,
		ExtraData:  payResp.ExtraData,
	}, nil
}

// QueryPayment 查询支付
func (s *Service) QueryPayment(ctx context.Context, orderNo string) (interface{}, error) {
	// 查询订单
	order, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}

	// 如果订单已经是最终状态，直接返回
	if order.Status == entity.OrderStatusSuccess || order.Status == entity.OrderStatusClosed {
		return order, nil
	}

	// 获取支付配置
	config, err := s.configRepo.GetByID(ctx, order.ConfigID)
	if err != nil {
		return nil, err
	}

	// 获取支付提供商
	provider, err := payment.GetProvider(order.Provider)
	if err != nil {
		return nil, err
	}

	// 查询支付状态
	queryReq := &payment.QueryPaymentRequest{
		OutTradeNo: order.OutTradeNo,
		TradeNo:    order.TradeNo,
		Config:     config.ConfigData,
	}

	queryResp, err := provider.QueryPayment(ctx, queryReq)
	if err != nil {
		s.logPayment(ctx, order.ID, orderNo, "query", order.Provider, queryReq, nil, "failed", err.Error())
		return order, nil
	}

	// 记录日志
	s.logPayment(ctx, order.ID, orderNo, "query", order.Provider, queryReq, queryResp, "success", "")

	// 更新订单状态
	if queryResp.Status != order.Status {
		oldStatus := order.Status
		order.Status = queryResp.Status
		if queryResp.Status == entity.OrderStatusSuccess {
			now := time.Now()
			order.PaymentTime = &now
		}
		if queryResp.TradeNo != "" {
			order.TradeNo = queryResp.TradeNo
		}
		if err := s.orderRepo.Update(ctx, order); err != nil {
			logger.Error("failed to update order", zap.Error(err))
			return order, err
		}

		// 如果订单状态变为成功，且有通知URL，添加通知任务
		if order.Status == entity.OrderStatusSuccess && oldStatus != entity.OrderStatusSuccess && order.NotifyURL != "" {
			notifyData := map[string]interface{}{
				"order_no":     order.OrderNo,
				"out_trade_no": order.OutTradeNo,
				"trade_no":     order.TradeNo,
				"amount":       order.Amount,
				"currency":     order.Currency,
				"status":       order.Status,
				"payment_time": order.PaymentTime,
				"subject":      order.Subject,
			}

			if err := s.notifyService.AddNotify(ctx, order.ID, order.OrderNo, order.NotifyURL, notifyData); err != nil {
				logger.Error("failed to add notify task",
					zap.Uint64("order_id", order.ID),
					zap.String("order_no", order.OrderNo),
					zap.Error(err))
				// 不影响主流程，继续返回
			} else {
				logger.Info("notify task added",
					zap.Uint64("order_id", order.ID),
					zap.String("order_no", order.OrderNo),
					zap.String("notify_url", order.NotifyURL))
			}
		}
	}

	return order, nil
}

// HandleNotify 处理支付通知
func (s *Service) HandleNotify(ctx context.Context, provider string, req *payment.NotifyRequest) ([]byte, error) {
	// 获取支付提供商
	prov, err := payment.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	// 处理通知
	notifyResp, err := prov.HandleNotify(ctx, req)
	if err != nil {
		logger.Error("handle notify failed", zap.Error(err))
		return nil, err
	}

	// 查询订单
	order, err := s.orderRepo.GetByOutTradeNo(ctx, notifyResp.OutTradeNo)
	if err != nil {
		logger.Error("order not found", zap.String("out_trade_no", notifyResp.OutTradeNo))
		return notifyResp.ReturnData, nil
	}

	// 记录日志
	s.logPayment(ctx, order.ID, order.OrderNo, "notify", provider, req, notifyResp, "success", "")

	// 更新订单状态
	if notifyResp.Status != order.Status {
		oldStatus := order.Status
		order.Status = notifyResp.Status
		if notifyResp.TradeNo != "" {
			order.TradeNo = notifyResp.TradeNo
		}
		if notifyResp.Status == entity.OrderStatusSuccess {
			now := time.Now()
			order.PaymentTime = &now
		}
		if err := s.orderRepo.Update(ctx, order); err != nil {
			logger.Error("failed to update order", zap.Error(err))
			return notifyResp.ReturnData, err
		}

		// 如果订单状态变为成功，且有通知URL，添加通知任务
		if order.Status == entity.OrderStatusSuccess && oldStatus != entity.OrderStatusSuccess && order.NotifyURL != "" {
			notifyData := map[string]interface{}{
				"order_no":     order.OrderNo,
				"out_trade_no": order.OutTradeNo,
				"trade_no":     order.TradeNo,
				"amount":       order.Amount,
				"currency":     order.Currency,
				"status":       order.Status,
				"payment_time": order.PaymentTime,
				"subject":      order.Subject,
			}

			if err := s.notifyService.AddNotify(ctx, order.ID, order.OrderNo, order.NotifyURL, notifyData); err != nil {
				logger.Error("failed to add notify task",
					zap.Uint64("order_id", order.ID),
					zap.String("order_no", order.OrderNo),
					zap.Error(err))
				// 不影响主流程，继续返回
			} else {
				logger.Info("notify task added",
					zap.Uint64("order_id", order.ID),
					zap.String("order_no", order.OrderNo),
					zap.String("notify_url", order.NotifyURL))
			}
		}
	}

	return notifyResp.ReturnData, nil
}

// GetConfigByID 根据配置ID获取支付配置
func (s *Service) GetConfigByID(ctx context.Context, configID uint64) (map[string]interface{}, error) {
	config, err := s.configRepo.GetByID(ctx, configID)
	if err != nil {
		return nil, err
	}

	return config.ConfigData, nil
}

// generateOrderNo 生成订单号
func (s *Service) generateOrderNo() string {
	return fmt.Sprintf("UNI%s%s", time.Now().Format("20060102150405"), uuid.New().String()[:8])
}

// logPayment 记录支付日志
func (s *Service) logPayment(ctx context.Context, orderID uint64, orderNo, action, provider string, request, response interface{}, status, errorMsg string) {
	log := &entity.PaymentLog{
		OrderID:      orderID,
		OrderNo:      orderNo,
		Action:       action,
		Provider:     provider,
		RequestData:  toConfigData(request),
		ResponseData: toConfigData(response),
		Status:       status,
		ErrorMsg:     errorMsg,
	}

	s.logRepo.Create(ctx, log)
}

// toConfigData 转换为ConfigData
func toConfigData(data interface{}) entity.ConfigData {
	if data == nil {
		return nil
	}

	if m, ok := data.(map[string]interface{}); ok {
		return entity.ConfigData(m)
	}

	return entity.ConfigData{}
}
