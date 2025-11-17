package entity

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// User 用户实体
type User struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	APIKey    string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"api_key"`
	APISecret string    `gorm:"type:varchar(128);not null" json:"-"`
	Status    int8      `gorm:"type:tinyint;not null;default:1" json:"status"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// PaymentConfig 支付配置实体
type PaymentConfig struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64     `gorm:"not null;index" json:"user_id"`
	Provider   string     `gorm:"type:varchar(20);not null;index" json:"provider"`
	ConfigName string     `gorm:"type:varchar(50);not null" json:"config_name"`
	ConfigData ConfigData `gorm:"type:json;not null" json:"config_data"`
	Status     int8       `gorm:"type:tinyint;not null;default:1" json:"status"`
	CreatedAt  time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 表名
func (PaymentConfig) TableName() string {
	return "payment_configs"
}

// ConfigData 配置数据（JSON类型）
type ConfigData map[string]interface{}

// Value 实现driver.Valuer接口
func (c ConfigData) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan 实现sql.Scanner接口
func (c *ConfigData) Scan(value interface{}) error {
	if value == nil {
		*c = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), c)
	}
	return json.Unmarshal(bytes, c)
}

// PaymentOrder 支付订单实体
type PaymentOrder struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo     string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"order_no"`
	UserID      uint64     `gorm:"not null;index" json:"user_id"`
	Provider    string     `gorm:"type:varchar(20);not null;index" json:"provider"`
	ConfigID    uint64     `gorm:"not null" json:"config_id"`
	OutTradeNo  string     `gorm:"type:varchar(64);not null;index" json:"out_trade_no"`
	TradeNo     string     `gorm:"type:varchar(64);index" json:"trade_no"`
	Subject     string     `gorm:"type:varchar(256);not null" json:"subject"`
	Body        string     `gorm:"type:text" json:"body"`
	Amount      float64    `gorm:"type:decimal(10,2);not null" json:"amount"`
	Currency    string     `gorm:"type:varchar(10);not null;default:'CNY'" json:"currency"`
	Status      string     `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	NotifyURL   string     `gorm:"type:varchar(512)" json:"notify_url"`
	ReturnURL   string     `gorm:"type:varchar(512)" json:"return_url"`
	ClientIP    string     `gorm:"type:varchar(45)" json:"client_ip"`
	ExtraData   ConfigData `gorm:"type:json" json:"extra_data"`
	PaymentTime *time.Time `gorm:"index" json:"payment_time"`
	CreatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 表名
func (PaymentOrder) TableName() string {
	return "payment_orders"
}

// OrderStatus 订单状态常量
const (
	OrderStatusPending    = "pending"
	OrderStatusProcessing = "processing"
	OrderStatusSuccess    = "success"
	OrderStatusFailed     = "failed"
	OrderStatusClosed     = "closed"
)

// PaymentLog 支付日志实体
type PaymentLog struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID      uint64     `gorm:"not null;index" json:"order_id"`
	OrderNo      string     `gorm:"type:varchar(64);not null;index" json:"order_no"`
	Action       string     `gorm:"type:varchar(50);not null;index" json:"action"`
	Provider     string     `gorm:"type:varchar(20);not null" json:"provider"`
	RequestData  ConfigData `gorm:"type:json" json:"request_data"`
	ResponseData ConfigData `gorm:"type:json" json:"response_data"`
	Status       string     `gorm:"type:varchar(20);not null" json:"status"`
	ErrorMsg     string     `gorm:"type:text" json:"error_msg"`
	CreatedAt    time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"created_at"`
}

// TableName 表名
func (PaymentLog) TableName() string {
	return "payment_logs"
}

// APILog API调用记录实体
type APILog struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint64    `gorm:"not null;index" json:"user_id"`
	APIKey         string    `gorm:"type:varchar(64);not null;index" json:"api_key"`
	Method         string    `gorm:"type:varchar(10);not null" json:"method"`
	Path           string    `gorm:"type:varchar(255);not null;index" json:"path"`
	Query          string    `gorm:"type:text" json:"query"`
	RequestBody    string    `gorm:"type:text" json:"request_body"`
	ResponseStatus int       `gorm:"not null" json:"response_status"`
	ResponseBody   string    `gorm:"type:text" json:"response_body"`
	IP             string    `gorm:"type:varchar(45)" json:"ip"`
	UserAgent      string    `gorm:"type:varchar(512)" json:"user_agent"`
	Duration       int       `gorm:"comment:请求耗时(毫秒)" json:"duration"`
	CreatedAt      time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"created_at"`
}

// TableName 表名
func (APILog) TableName() string {
	return "api_logs"
}

// NotifyQueue 通知队列实体
type NotifyQueue struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID       uint64     `gorm:"not null;index" json:"order_id"`
	OrderNo       string     `gorm:"type:varchar(64);not null;index" json:"order_no"`
	NotifyURL     string     `gorm:"type:varchar(512);not null" json:"notify_url"`
	NotifyData    ConfigData `gorm:"type:json;not null" json:"notify_data"`
	RetryCount    int        `gorm:"not null;default:0" json:"retry_count"`
	MaxRetry      int        `gorm:"not null;default:5" json:"max_retry"`
	Status        string     `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	LastError     string     `gorm:"type:text" json:"last_error"`
	NextRetryTime *time.Time `gorm:"index" json:"next_retry_time"`
	SuccessTime   *time.Time `json:"success_time"`
	CreatedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 表名
func (NotifyQueue) TableName() string {
	return "notify_queue"
}

// NotifyStatus 通知状态常量
const (
	NotifyStatusPending    = "pending"
	NotifyStatusProcessing = "processing"
	NotifyStatusSuccess    = "success"
	NotifyStatusFailed     = "failed"
)
