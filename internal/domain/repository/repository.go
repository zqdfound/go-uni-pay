package repository

import (
	"context"

	"github.com/zqdfound/go-uni-pay/internal/domain/entity"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uint64) (*entity.User, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
}

// PaymentConfigRepository 支付配置仓储接口
type PaymentConfigRepository interface {
	Create(ctx context.Context, config *entity.PaymentConfig) error
	GetByID(ctx context.Context, id uint64) (*entity.PaymentConfig, error)
	GetByUserAndProvider(ctx context.Context, userID uint64, provider string) ([]*entity.PaymentConfig, error)
	GetActiveByUserAndProvider(ctx context.Context, userID uint64, provider string) (*entity.PaymentConfig, error)
	Update(ctx context.Context, config *entity.PaymentConfig) error
	Delete(ctx context.Context, id uint64) error
}

// PaymentOrderRepository 支付订单仓储接口
type PaymentOrderRepository interface {
	Create(ctx context.Context, order *entity.PaymentOrder) error
	GetByID(ctx context.Context, id uint64) (*entity.PaymentOrder, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*entity.PaymentOrder, error)
	GetByOutTradeNo(ctx context.Context, outTradeNo string) (*entity.PaymentOrder, error)
	Update(ctx context.Context, order *entity.PaymentOrder) error
	List(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.PaymentOrder, int64, error)
}

// PaymentLogRepository 支付日志仓储接口
type PaymentLogRepository interface {
	Create(ctx context.Context, log *entity.PaymentLog) error
	List(ctx context.Context, orderID uint64, page, pageSize int) ([]*entity.PaymentLog, int64, error)
}

// APILogRepository API日志仓储接口
type APILogRepository interface {
	Create(ctx context.Context, log *entity.APILog) error
	List(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.APILog, int64, error)
}

// NotifyQueueRepository 通知队列仓储接口
type NotifyQueueRepository interface {
	Create(ctx context.Context, queue *entity.NotifyQueue) error
	GetByID(ctx context.Context, id uint64) (*entity.NotifyQueue, error)
	GetPendingTasks(ctx context.Context, limit int) ([]*entity.NotifyQueue, error)
	Update(ctx context.Context, queue *entity.NotifyQueue) error
}
