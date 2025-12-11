package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zqdfound/go-uni-pay/internal/api/handler"
	"github.com/zqdfound/go-uni-pay/internal/api/router"
	"github.com/zqdfound/go-uni-pay/internal/domain/repository"
	"github.com/zqdfound/go-uni-pay/internal/infrastructure/cache"
	"github.com/zqdfound/go-uni-pay/internal/infrastructure/config"
	"github.com/zqdfound/go-uni-pay/internal/infrastructure/database"
	"github.com/zqdfound/go-uni-pay/internal/service/admin"
	"github.com/zqdfound/go-uni-pay/internal/service/auth"
	"github.com/zqdfound/go-uni-pay/internal/service/notify"
	"github.com/zqdfound/go-uni-pay/internal/service/payment"
	"github.com/zqdfound/go-uni-pay/pkg/logger"
	"go.uber.org/zap"

	// 导入支付提供商，触发init注册
	_ "github.com/zqdfound/go-uni-pay/internal/payment/alipay"
	_ "github.com/zqdfound/go-uni-pay/internal/payment/paypal"
	_ "github.com/zqdfound/go-uni-pay/internal/payment/stripe"
	_ "github.com/zqdfound/go-uni-pay/internal/payment/wechat"
)

func main() {
	// 加载配置
	configPath := "configs/config-dev.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	if err := config.Load(configPath); err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(logger.Config{
		Level:      config.Cfg.Logger.Level,
		Filename:   config.Cfg.Logger.Filename,
		MaxSize:    config.Cfg.Logger.MaxSize,
		MaxAge:     config.Cfg.Logger.MaxAge,
		MaxBackups: config.Cfg.Logger.MaxBackups,
		Compress:   config.Cfg.Logger.Compress,
	}); err != nil {
		fmt.Printf("failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("starting go-uni-pay server...")

	// 初始化数据库
	if err := database.Init(database.Config{
		DSN:             config.Cfg.Database.GetDSN(),
		MaxIdleConns:    config.Cfg.Database.MaxIdleConns,
		MaxOpenConns:    config.Cfg.Database.MaxOpenConns,
		ConnMaxLifetime: config.Cfg.Database.ConnMaxLifetime,
	}); err != nil {
		logger.Fatal("failed to init database", zap.Error(err))
	}
	defer database.Close()

	logger.Info("database connected")

	// 初始化Redis
	if err := cache.Init(cache.Config{
		Addr:     config.Cfg.Redis.GetRedisAddr(),
		Password: config.Cfg.Redis.Password,
		DB:       config.Cfg.Redis.DB,
		PoolSize: config.Cfg.Redis.PoolSize,
	}); err != nil {
		logger.Fatal("failed to init redis", zap.Error(err))
	}
	defer cache.Close()

	logger.Info("redis connected")

	// 创建仓储
	db := database.GetDB()
	userRepo := repository.NewMySQLUserRepository(db)
	paymentConfigRepo := repository.NewMySQLPaymentConfigRepository(db)
	paymentOrderRepo := repository.NewMySQLPaymentOrderRepository(db)
	paymentLogRepo := repository.NewMySQLPaymentLogRepository(db)
	apiLogRepo := repository.NewMySQLAPILogRepository(db)
	notifyQueueRepo := repository.NewMySQLNotifyQueueRepository(db)
	adminRepo := repository.NewMySQLAdminRepository(db)

	// 创建服务
	authService := auth.NewService(userRepo)
	adminService := admin.NewService(adminRepo, config.Cfg)

	// 先创建通知服务
	notifyService := notify.NewService(
		notifyQueueRepo,
		config.Cfg.Notify.WorkerCount,
		time.Duration(config.Cfg.Notify.RetryInterval)*time.Second,
		config.Cfg.Notify.MaxRetry,
	)

	// 创建支付服务，注入通知服务
	paymentService := payment.NewService(paymentOrderRepo, paymentConfigRepo, paymentLogRepo, notifyService)

	// 启动通知服务
	notifyService.Start()
	defer notifyService.Stop()

	// 创建处理器
	paymentHandler := handler.NewPaymentHandler(paymentService)
	adminHandler := handler.NewAdminHandler(adminService)
	managementHandler := handler.NewManagementHandler(
		userRepo,
		paymentConfigRepo,
		paymentOrderRepo,
		paymentLogRepo,
		apiLogRepo,
		notifyQueueRepo,
	)

	// 设置Gin模式
	gin.SetMode(config.Cfg.Server.Mode)

	// 创建路由
	r := router.SetupRouter(authService, paymentHandler, adminHandler, managementHandler, adminService, apiLogRepo)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         config.Cfg.Server.GetServerAddr(),
		Handler:      r,
		ReadTimeout:  config.Cfg.Server.GetReadTimeout(),
		WriteTimeout: config.Cfg.Server.GetWriteTimeout(),
	}

	// 启动服务器
	go func() {
		logger.Info("server starting",
			zap.String("addr", srv.Addr),
			zap.String("mode", config.Cfg.Server.Mode))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	// 等待中断信号优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited")
}
