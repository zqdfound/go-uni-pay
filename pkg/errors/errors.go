package errors

import (
	"fmt"
)

// ErrorCode 错误码
type ErrorCode int

const (
	// 通用错误码 1000-1999
	ErrSuccess          ErrorCode = 0
	ErrInternalServer   ErrorCode = 1000
	ErrInvalidParam     ErrorCode = 1001
	ErrUnauthorized     ErrorCode = 1002
	ErrForbidden        ErrorCode = 1003
	ErrNotFound         ErrorCode = 1004
	ErrConflict         ErrorCode = 1005
	ErrTooManyRequests  ErrorCode = 1006

	// 支付相关错误码 2000-2999
	ErrPaymentCreate    ErrorCode = 2000
	ErrPaymentQuery     ErrorCode = 2001
	ErrPaymentNotify    ErrorCode = 2002
	ErrPaymentRefund    ErrorCode = 2003
	ErrPaymentCancel    ErrorCode = 2004
	ErrProviderNotFound ErrorCode = 2005
	ErrConfigNotFound   ErrorCode = 2006
	ErrOrderNotFound    ErrorCode = 2007
	ErrOrderStatus      ErrorCode = 2008
	ErrAmountInvalid    ErrorCode = 2009

	// 数据库错误码 3000-3999
	ErrDatabaseQuery    ErrorCode = 3000
	ErrDatabaseInsert   ErrorCode = 3001
	ErrDatabaseUpdate   ErrorCode = 3002
	ErrDatabaseDelete   ErrorCode = 3003

	// Redis错误码 4000-4999
	ErrRedisGet         ErrorCode = 4000
	ErrRedisSet         ErrorCode = 4001
	ErrRedisDel         ErrorCode = 4002
	ErrRedisLock        ErrorCode = 4003
)

// AppError 应用错误
type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *AppError) Unwrap() error {
	return e.Err
}

// New 创建新错误
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装错误
func Wrap(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// 错误消息映射
var errorMessages = map[ErrorCode]string{
	ErrSuccess:          "Success",
	ErrInternalServer:   "Internal server error",
	ErrInvalidParam:     "Invalid parameter",
	ErrUnauthorized:     "Unauthorized",
	ErrForbidden:        "Forbidden",
	ErrNotFound:         "Not found",
	ErrConflict:         "Conflict",
	ErrTooManyRequests:  "Too many requests",
	ErrPaymentCreate:    "Failed to create payment",
	ErrPaymentQuery:     "Failed to query payment",
	ErrPaymentNotify:    "Failed to process payment notification",
	ErrPaymentRefund:    "Failed to refund payment",
	ErrPaymentCancel:    "Failed to cancel payment",
	ErrProviderNotFound: "Payment provider not found",
	ErrConfigNotFound:   "Payment config not found",
	ErrOrderNotFound:    "Order not found",
	ErrOrderStatus:      "Invalid order status",
	ErrAmountInvalid:    "Invalid amount",
	ErrDatabaseQuery:    "Database query error",
	ErrDatabaseInsert:   "Database insert error",
	ErrDatabaseUpdate:   "Database update error",
	ErrDatabaseDelete:   "Database delete error",
	ErrRedisGet:         "Redis get error",
	ErrRedisSet:         "Redis set error",
	ErrRedisDel:         "Redis delete error",
	ErrRedisLock:        "Redis lock error",
}

// GetMessage 获取错误消息
func GetMessage(code ErrorCode) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "Unknown error"
}
