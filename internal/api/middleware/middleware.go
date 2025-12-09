package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zqdfound/go-uni-pay/internal/domain/entity"
	"github.com/zqdfound/go-uni-pay/internal/domain/repository"
	"github.com/zqdfound/go-uni-pay/internal/infrastructure/cache"
	"github.com/zqdfound/go-uni-pay/internal/service/auth"
	apperrors "github.com/zqdfound/go-uni-pay/pkg/errors"
	"github.com/zqdfound/go-uni-pay/pkg/logger"
	"go.uber.org/zap"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			c.JSON(401, gin.H{
				"code":    apperrors.ErrUnauthorized,
				"message": "missing api key",
			})
			c.Abort()
			return
		}

		// 验证API Key
		user, err := authService.ValidateAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			c.JSON(401, gin.H{
				"code":    apperrors.ErrUnauthorized,
				"message": "invalid api key",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user", user)
		c.Set("user_id", user.ID)

		c.Next()
	}
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware(apiLogRepo repository.APILogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 读取请求体
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			requestBody = string(bodyBytes)
			// 重新设置请求体
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 使用自定义ResponseWriter捕获响应
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// 计算耗时
		duration := time.Since(start).Milliseconds()

		// 获取用户信息
		var userID uint64
		var apiKey string
		if user, exists := c.Get("user"); exists {
			u := user.(*entity.User)
			userID = u.ID
			apiKey = u.APIKey
		}

		// 创建API日志
		apiLog := &entity.APILog{
			UserID:         userID,
			APIKey:         apiKey,
			Method:         c.Request.Method,
			Path:           c.Request.URL.Path,
			Query:          c.Request.URL.RawQuery,
			RequestBody:    requestBody,
			ResponseStatus: c.Writer.Status(),
			ResponseBody:   blw.body.String(),
			IP:             c.ClientIP(),
			UserAgent:      c.Request.UserAgent(),
			Duration:       int(duration),
		}

		// 异步保存日志
		go func() {
			// 使用独立的 context，避免请求结束后 context 被取消
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := apiLogRepo.Create(ctx, apiLog); err != nil {
				logger.Error("failed to create api log", zap.Error(err))
			}
		}()

		// 打印日志
		logger.Info("api request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("duration_ms", duration),
			zap.String("ip", c.ClientIP()),
		)
	}
}

// bodyLogWriter 自定义ResponseWriter，用于捕获响应体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
				)

				c.JSON(500, gin.H{
					"code":    apperrors.ErrInternalServer,
					"message": "internal server error",
				})
			}
		}()

		c.Next()
	}
}

// RateLimitMiddleware API限流中间件
// 基于Redis实现滑动窗口限流
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID，如果未认证则使用IP地址
		var identifier string
		if userID, exists := c.Get("user_id"); exists {
			identifier = fmt.Sprintf("user:%v", userID)
		} else {
			identifier = fmt.Sprintf("ip:%s", c.ClientIP())
		}

		// 构造限流key
		key := fmt.Sprintf("ratelimit:%s", identifier)

		// 使用Redis进行计数
		ctx := c.Request.Context()
		count, err := cache.Client.Incr(ctx, key).Result()
		if err != nil {
			// Redis错误不应该阻止请求
			logger.Error("rate limit redis error", zap.Error(err))
			c.Next()
			return
		}

		// 第一次请求时设置过期时间
		if count == 1 {
			cache.Client.Expire(ctx, key, window)
		}

		// 设置响应头，告知客户端限流信息
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, limit-int(count))))

		// 检查是否超过限制
		if count > int64(limit) {
			logger.Warn("rate limit exceeded",
				zap.String("identifier", identifier),
				zap.Int64("count", count),
				zap.Int("limit", limit),
			)

			c.JSON(429, gin.H{
				"code":    1005,
				"message": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
