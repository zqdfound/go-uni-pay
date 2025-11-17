package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zqdfound/go-uni-pay/internal/domain/entity"
	"github.com/zqdfound/go-uni-pay/internal/domain/repository"
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
			if err := apiLogRepo.Create(c.Request.Context(), apiLog); err != nil {
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
