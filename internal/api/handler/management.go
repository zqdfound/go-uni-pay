package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zqdfound/go-uni-pay/internal/domain/repository"
	apperrors "github.com/zqdfound/go-uni-pay/pkg/errors"
)

// ManagementHandler 管理后台处理器
type ManagementHandler struct {
	userRepo        repository.UserRepository
	configRepo      repository.PaymentConfigRepository
	orderRepo       repository.PaymentOrderRepository
	paymentLogRepo  repository.PaymentLogRepository
	apiLogRepo      repository.APILogRepository
	notifyQueueRepo repository.NotifyQueueRepository
}

// NewManagementHandler 创建管理后台处理器
func NewManagementHandler(
	userRepo repository.UserRepository,
	configRepo repository.PaymentConfigRepository,
	orderRepo repository.PaymentOrderRepository,
	paymentLogRepo repository.PaymentLogRepository,
	apiLogRepo repository.APILogRepository,
	notifyQueueRepo repository.NotifyQueueRepository,
) *ManagementHandler {
	return &ManagementHandler{
		userRepo:        userRepo,
		configRepo:      configRepo,
		orderRepo:       orderRepo,
		paymentLogRepo:  paymentLogRepo,
		apiLogRepo:      apiLogRepo,
		notifyQueueRepo: notifyQueueRepo,
	}
}

// ListUsers 获取用户列表
func (h *ManagementHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := h.userRepo.List(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    apperrors.ErrInternalServer,
			"message": "failed to list users",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    apperrors.ErrSuccess,
		"message": "success",
		"data": gin.H{
			"list":      users,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetUser 获取用户详情
func (h *ManagementHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{
			"code":    apperrors.ErrInvalidParam,
			"message": "invalid id",
		})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(400, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}

		c.JSON(500, gin.H{
			"code":    apperrors.ErrInternalServer,
			"message": "internal server error",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    apperrors.ErrSuccess,
		"message": "success",
		"data":    user,
	})
}

// ListPaymentConfigs 获取支付配置列表
func (h *ManagementHandler) ListPaymentConfigs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	configs, total, err := h.configRepo.List(c.Request.Context(), page, pageSize, userID)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    apperrors.ErrInternalServer,
			"message": "failed to list payment configs",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    apperrors.ErrSuccess,
		"message": "success",
		"data": gin.H{
			"list":      configs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetPaymentConfig 获取支付配置详情
func (h *ManagementHandler) GetPaymentConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{
			"code":    apperrors.ErrInvalidParam,
			"message": "invalid id",
		})
		return
	}

	config, err := h.configRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(400, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}

		c.JSON(500, gin.H{
			"code":    apperrors.ErrInternalServer,
			"message": "internal server error",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    apperrors.ErrSuccess,
		"message": "success",
		"data":    config,
	})
}

// ListOrders 获取订单列表
func (h *ManagementHandler) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	orders, total, err := h.orderRepo.List(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    apperrors.ErrInternalServer,
			"message": "failed to list orders",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    apperrors.ErrSuccess,
		"message": "success",
		"data": gin.H{
			"list":      orders,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetOrder 获取订单详情
func (h *ManagementHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{
			"code":    apperrors.ErrInvalidParam,
			"message": "invalid id",
		})
		return
	}

	order, err := h.orderRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(400, gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
			return
		}

		c.JSON(500, gin.H{
			"code":    apperrors.ErrInternalServer,
			"message": "internal server error",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    apperrors.ErrSuccess,
		"message": "success",
		"data":    order,
	})
}

// ListPaymentLogs 获取支付日志列表
func (h *ManagementHandler) ListPaymentLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	logs, total, err := h.paymentLogRepo.ListAll(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    apperrors.ErrInternalServer,
			"message": "failed to list payment logs",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    apperrors.ErrSuccess,
		"message": "success",
		"data": gin.H{
			"list":      logs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// ListAPILogs 获取API调用日志列表
func (h *ManagementHandler) ListAPILogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	logs, total, err := h.apiLogRepo.ListAll(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    apperrors.ErrInternalServer,
			"message": "failed to list api logs",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    apperrors.ErrSuccess,
		"message": "success",
		"data": gin.H{
			"list":      logs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// ListNotifyQueue 获取通知队列列表
func (h *ManagementHandler) ListNotifyQueue(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	queues, total, err := h.notifyQueueRepo.List(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    apperrors.ErrInternalServer,
			"message": "failed to list notify queue",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    apperrors.ErrSuccess,
		"message": "success",
		"data": gin.H{
			"list":      queues,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
