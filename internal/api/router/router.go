package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zqdfound/go-uni-pay/internal/api/handler"
	"github.com/zqdfound/go-uni-pay/internal/api/middleware"
	"github.com/zqdfound/go-uni-pay/internal/domain/repository"
	"github.com/zqdfound/go-uni-pay/internal/service/admin"
	"github.com/zqdfound/go-uni-pay/internal/service/auth"
)

// SetupRouter 设置路由
func SetupRouter(
	authService *auth.Service,
	paymentHandler *handler.PaymentHandler,
	adminHandler *handler.AdminHandler,
	managementHandler *handler.ManagementHandler,
	adminService *admin.Service,
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
		authenticated := v1.Group("")
		authenticated.Use(middleware.AuthMiddleware(authService))
		// 添加限流：每分钟最多100次请求
		authenticated.Use(middleware.RateLimitMiddleware(100, time.Minute))
		{
			// 支付相关接口
			payment := authenticated.Group("/payment")
			{
				payment.POST("/create", paymentHandler.CreatePayment)
				payment.GET("/query/:order_no", paymentHandler.QueryPayment)
			}
		}

		// 管理后台接口
		admin := v1.Group("/admin")
		{
			// 管理员登录（无需认证）
			admin.POST("/login", adminHandler.Login)

			// 需要管理员认证的接口
			adminAuth := admin.Group("")
			adminAuth.Use(middleware.AdminAuthMiddleware(adminService))
			{
				// 管理员管理
				adminAuth.GET("/admins", adminHandler.ListAdmins)
				adminAuth.GET("/admins/:id", adminHandler.GetAdmin)
				adminAuth.POST("/admins", adminHandler.CreateAdmin)
				adminAuth.PUT("/admins", adminHandler.UpdateAdmin)
				adminAuth.DELETE("/admins/:id", adminHandler.DeleteAdmin)
				adminAuth.GET("/current", adminHandler.GetCurrentAdmin)

				// 用户管理
				adminAuth.GET("/users", managementHandler.ListUsers)
				adminAuth.GET("/users/:id", managementHandler.GetUser)

				// 支付配置管理
				adminAuth.GET("/configs", managementHandler.ListPaymentConfigs)
				adminAuth.GET("/configs/:id", managementHandler.GetPaymentConfig)

				// 订单管理
				adminAuth.GET("/orders", managementHandler.ListOrders)
				adminAuth.GET("/orders/:id", managementHandler.GetOrder)

				// 支付日志
				adminAuth.GET("/payment-logs", managementHandler.ListPaymentLogs)

				// API调用日志
				adminAuth.GET("/api-logs", managementHandler.ListAPILogs)

				// 通知队列
				adminAuth.GET("/notify-queue", managementHandler.ListNotifyQueue)
			}
		}
	}

	return r
}
