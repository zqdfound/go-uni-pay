package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zqdfound/go-uni-pay/internal/api/handler"
	"github.com/zqdfound/go-uni-pay/internal/api/middleware"
	"github.com/zqdfound/go-uni-pay/internal/domain/repository"
	"github.com/zqdfound/go-uni-pay/internal/service/auth"
)

// SetupRouter 设置路由
func SetupRouter(
	authService *auth.Service,
	paymentHandler *handler.PaymentHandler,
	apiLogRepo repository.APILogRepository,
) *gin.Engine {
	r := gin.New()

	// 全局中间件
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LoggerMiddleware(apiLogRepo))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API版本1
	v1 := r.Group("/api/v1")
	{
		// 公开接口（不需要认证）
		public := v1.Group("/public")
		{
			// 支付通知回调（由第三方支付平台调用）
			// URL格式: /api/v1/public/notify/:provider/:config_id
			// 例如: /api/v1/public/notify/paypal/123
			public.POST("/notify/:provider/:config_id", paymentHandler.HandleNotify)
		}

		// 需要认证的接口
		auth := v1.Group("")
		auth.Use(middleware.AuthMiddleware(authService))
		// 添加限流：每分钟最多100次请求
		auth.Use(middleware.RateLimitMiddleware(100, time.Minute))
		{
			// 支付相关接口
			payment := auth.Group("/payment")
			{
				payment.POST("/create", paymentHandler.CreatePayment)
				payment.GET("/query/:order_no", paymentHandler.QueryPayment)
			}
		}
	}

	return r
}
