package repository

import (
	"context"
	"time"

	"github.com/zqdfound/go-uni-pay/internal/domain/entity"
	apperrors "github.com/zqdfound/go-uni-pay/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MySQLUserRepository MySQL用户仓储实现
type MySQLUserRepository struct {
	db *gorm.DB
}

// NewMySQLUserRepository 创建MySQL用户仓储
func NewMySQLUserRepository(db *gorm.DB) *MySQLUserRepository {
	return &MySQLUserRepository{db: db}
}

func (r *MySQLUserRepository) Create(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return apperrors.Wrap(apperrors.ErrDatabaseInsert, "failed to create user", err)
	}
	return nil
}

func (r *MySQLUserRepository) GetByID(ctx context.Context, id uint64) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.New(apperrors.ErrNotFound, "user not found")
		}
		return nil, apperrors.Wrap(apperrors.ErrDatabaseQuery, "failed to get user", err)
	}
	return &user, nil
}

func (r *MySQLUserRepository) GetByAPIKey(ctx context.Context, apiKey string) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).Where("api_key = ?", apiKey).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.New(apperrors.ErrNotFound, "user not found")
		}
		return nil, apperrors.Wrap(apperrors.ErrDatabaseQuery, "failed to get user", err)
	}
	return &user, nil
}

func (r *MySQLUserRepository) Update(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return apperrors.Wrap(apperrors.ErrDatabaseUpdate, "failed to update user", err)
	}
	return nil
}

// MySQLPaymentConfigRepository MySQL支付配置仓储实现
type MySQLPaymentConfigRepository struct {
	db *gorm.DB
}

// NewMySQLPaymentConfigRepository 创建MySQL支付配置仓储
func NewMySQLPaymentConfigRepository(db *gorm.DB) *MySQLPaymentConfigRepository {
	return &MySQLPaymentConfigRepository{db: db}
}

func (r *MySQLPaymentConfigRepository) Create(ctx context.Context, config *entity.PaymentConfig) error {
	if err := r.db.WithContext(ctx).Create(config).Error; err != nil {
		return apperrors.Wrap(apperrors.ErrDatabaseInsert, "failed to create config", err)
	}
	return nil
}

func (r *MySQLPaymentConfigRepository) GetByID(ctx context.Context, id uint64) (*entity.PaymentConfig, error) {
	var config entity.PaymentConfig
	if err := r.db.WithContext(ctx).First(&config, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.New(apperrors.ErrNotFound, "config not found")
		}
		return nil, apperrors.Wrap(apperrors.ErrDatabaseQuery, "failed to get config", err)
	}
	return &config, nil
}

func (r *MySQLPaymentConfigRepository) GetByUserAndProvider(ctx context.Context, userID uint64, provider string) ([]*entity.PaymentConfig, error) {
	var configs []*entity.PaymentConfig
	if err := r.db.WithContext(ctx).Where("user_id = ? AND provider = ?", userID, provider).Find(&configs).Error; err != nil {
		return nil, apperrors.Wrap(apperrors.ErrDatabaseQuery, "failed to get configs", err)
	}
	return configs, nil
}

func (r *MySQLPaymentConfigRepository) GetActiveByUserAndProvider(ctx context.Context, userID uint64, provider string) (*entity.PaymentConfig, error) {
	var config entity.PaymentConfig
	if err := r.db.WithContext(ctx).Where("user_id = ? AND provider = ? AND status = 1", userID, provider).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.New(apperrors.ErrConfigNotFound, "active config not found")
		}
		return nil, apperrors.Wrap(apperrors.ErrDatabaseQuery, "failed to get active config", err)
	}
	return &config, nil
}

func (r *MySQLPaymentConfigRepository) Update(ctx context.Context, config *entity.PaymentConfig) error {
	if err := r.db.WithContext(ctx).Save(config).Error; err != nil {
		return apperrors.Wrap(apperrors.ErrDatabaseUpdate, "failed to update config", err)
	}
	return nil
}

func (r *MySQLPaymentConfigRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&entity.PaymentConfig{}, id).Error; err != nil {
		return apperrors.Wrap(apperrors.ErrDatabaseDelete, "failed to delete config", err)
	}
	return nil
}

// MySQLPaymentOrderRepository MySQL支付订单仓储实现
type MySQLPaymentOrderRepository struct {
	db *gorm.DB
}

// NewMySQLPaymentOrderRepository 创建MySQL支付订单仓储
func NewMySQLPaymentOrderRepository(db *gorm.DB) *MySQLPaymentOrderRepository {
	return &MySQLPaymentOrderRepository{db: db}
}

func (r *MySQLPaymentOrderRepository) Create(ctx context.Context, order *entity.PaymentOrder) error {
	if err := r.db.WithContext(ctx).Create(order).Error; err != nil {
		return apperrors.Wrap(apperrors.ErrDatabaseInsert, "failed to create order", err)
	}
	return nil
}

func (r *MySQLPaymentOrderRepository) GetByID(ctx context.Context, id uint64) (*entity.PaymentOrder, error) {
	var order entity.PaymentOrder
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.New(apperrors.ErrOrderNotFound, "order not found")
		}
		return nil, apperrors.Wrap(apperrors.ErrDatabaseQuery, "failed to get order", err)
	}
	return &order, nil
}

func (r *MySQLPaymentOrderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*entity.PaymentOrder, error) {
	var order entity.PaymentOrder
	if err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.New(apperrors.ErrOrderNotFound, "order not found")
		}
		return nil, apperrors.Wrap(apperrors.ErrDatabaseQuery, "failed to get order", err)
	}
	return &order, nil
}

func (r *MySQLPaymentOrderRepository) GetByOutTradeNo(ctx context.Context, outTradeNo string) (*entity.PaymentOrder, error) {
	var order entity.PaymentOrder
	if err := r.db.WithContext(ctx).Where("out_trade_no = ?", outTradeNo).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.New(apperrors.ErrOrderNotFound, "order not found")
		}
		return nil, apperrors.Wrap(apperrors.ErrDatabaseQuery, "failed to get order", err)
	}
	return &order, nil
}

func (r *MySQLPaymentOrderRepository) Update(ctx context.Context, order *entity.PaymentOrder) error {
	if err := r.db.WithContext(ctx).Save(order).Error; err != nil {
		return apperrors.Wrap(apperrors.ErrDatabaseUpdate, "failed to update order", err)
	}
	return nil
}

func (r *MySQLPaymentOrderRepository) List(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.PaymentOrder, int64, error) {
	var orders []*entity.PaymentOrder
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PaymentOrder{})
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(apperrors.ErrDatabaseQuery, "failed to count orders", err)
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&orders).Error; err != nil {
		return nil, 0, apperrors.Wrap(apperrors.ErrDatabaseQuery, "failed to list orders", err)
	}

	return orders, total, nil
}

// MySQLPaymentLogRepository MySQL支付日志仓储实现
type MySQLPaymentLogRepository struct {
	db *gorm.DB
}

// NewMySQLPaymentLogRepository 创建MySQL支付日志仓储
func NewMySQLPaymentLogRepository(db *gorm.DB) *MySQLPaymentLogRepository {
	return &MySQLPaymentLogRepository{db: db}
}

func (r *MySQLPaymentLogRepository) Create(ctx context.Context, log *entity.PaymentLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return apperrors.Wrap(apperrors.ErrDatabaseInsert, "failed to create log", err)
	}
	return nil
}

func (r *MySQLPaymentLogRepository) List(ctx context.Context, orderID uint64, page, pageSize int) ([]*entity.PaymentLog, int64, error) {
	var logs []*entity.PaymentLog
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PaymentLog{}).Where("order_id = ?", orderID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(apperrors.ErrDatabaseQuery, "failed to count logs", err)
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, apperrors.Wrap(apperrors.ErrDatabaseQuery, "failed to list logs", err)
	}

	return logs, total, nil
}

// MySQLAPILogRepository MySQL API日志仓储实现
type MySQLAPILogRepository struct {
	db *gorm.DB
}

// NewMySQLAPILogRepository 创建MySQL API日志仓储
func NewMySQLAPILogRepository(db *gorm.DB) *MySQLAPILogRepository {
	return &MySQLAPILogRepository{db: db}
}

func (r *MySQLAPILogRepository) Create(ctx context.Context, log *entity.APILog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return apperrors.Wrap(apperrors.ErrDatabaseInsert, "failed to create api log", err)
	}
	return nil
}

func (r *MySQLAPILogRepository) List(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.APILog, int64, error) {
	var logs []*entity.APILog
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.APILog{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(apperrors.ErrDatabaseQuery, "failed to count api logs", err)
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, apperrors.Wrap(apperrors.ErrDatabaseQuery, "failed to list api logs", err)
	}

	return logs, total, nil
}

// MySQLNotifyQueueRepository MySQL通知队列仓储实现
type MySQLNotifyQueueRepository struct {
	db *gorm.DB
}

// NewMySQLNotifyQueueRepository 创建MySQL通知队列仓储
func NewMySQLNotifyQueueRepository(db *gorm.DB) *MySQLNotifyQueueRepository {
	return &MySQLNotifyQueueRepository{db: db}
}

func (r *MySQLNotifyQueueRepository) Create(ctx context.Context, queue *entity.NotifyQueue) error {
	if err := r.db.WithContext(ctx).Create(queue).Error; err != nil {
		return apperrors.Wrap(apperrors.ErrDatabaseInsert, "failed to create notify queue", err)
	}
	return nil
}

func (r *MySQLNotifyQueueRepository) GetByID(ctx context.Context, id uint64) (*entity.NotifyQueue, error) {
	var queue entity.NotifyQueue
	if err := r.db.WithContext(ctx).First(&queue, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.New(apperrors.ErrNotFound, "notify queue not found")
		}
		return nil, apperrors.Wrap(apperrors.ErrDatabaseQuery, "failed to get notify queue", err)
	}
	return &queue, nil
}

func (r *MySQLNotifyQueueRepository) GetPendingTasks(ctx context.Context, limit int) ([]*entity.NotifyQueue, error) {
	var queues []*entity.NotifyQueue
	now := time.Now()

	// 使用 FOR UPDATE SKIP LOCKED 避免多个 worker 获取相同任务
	// SKIP LOCKED 会跳过已被其他事务锁定的行，每个 worker 获取不同的任务
	if err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("status = ? AND retry_count < max_retry AND (next_retry_time IS NULL OR next_retry_time <= ?)",
			entity.NotifyStatusPending, now).
		Order("created_at ASC").
		Limit(limit).
		Find(&queues).Error; err != nil {
		return nil, apperrors.Wrap(apperrors.ErrDatabaseQuery, "failed to get pending tasks", err)
	}

	return queues, nil
}

func (r *MySQLNotifyQueueRepository) Update(ctx context.Context, queue *entity.NotifyQueue) error {
	if err := r.db.WithContext(ctx).Model(queue).Updates(map[string]interface{}{
		"status":          queue.Status,
		"retry_count":     queue.RetryCount,
		"last_error":      queue.LastError,
		"next_retry_time": queue.NextRetryTime,
		"success_time":    queue.SuccessTime,
	}).Error; err != nil {
		return apperrors.Wrap(apperrors.ErrDatabaseUpdate, "failed to update notify queue", err)
	}
	return nil
}
