package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zqdfound/go-uni-pay/internal/service/admin"
	apperrors "github.com/zqdfound/go-uni-pay/pkg/errors"
)

// AdminHandler 管理员处理器
type AdminHandler struct {
	adminService *admin.Service
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler(adminService *admin.Service) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

// Login 管理员登录
func (h *AdminHandler) Login(c *gin.Context) {
	var req admin.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    apperrors.ErrInvalidParam,
			"message": err.Error(),
		})
		return
	}

	resp, err := h.adminService.Login(c.Request.Context(), &req)
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
		"data":    resp,
	})
}

// CreateAdmin 创建管理员
func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	var req admin.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    apperrors.ErrInvalidParam,
			"message": err.Error(),
		})
		return
	}

	adm, err := h.adminService.CreateAdmin(c.Request.Context(), &req)
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
		"data":    adm,
	})
}

// UpdateAdmin 更新管理员
func (h *AdminHandler) UpdateAdmin(c *gin.Context) {
	var req admin.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    apperrors.ErrInvalidParam,
			"message": err.Error(),
		})
		return
	}

	err := h.adminService.UpdateAdmin(c.Request.Context(), &req)
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
	})
}

// GetAdmin 获取管理员详情
func (h *AdminHandler) GetAdmin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{
			"code":    apperrors.ErrInvalidParam,
			"message": "invalid id",
		})
		return
	}

	adm, err := h.adminService.GetAdminByID(c.Request.Context(), id)
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
		"data":    adm,
	})
}

// ListAdmins 获取管理员列表
func (h *AdminHandler) ListAdmins(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	admins, total, err := h.adminService.ListAdmins(c.Request.Context(), page, pageSize)
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
		"data": gin.H{
			"list":      admins,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// DeleteAdmin 删除管理员
func (h *AdminHandler) DeleteAdmin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{
			"code":    apperrors.ErrInvalidParam,
			"message": "invalid id",
		})
		return
	}

	err = h.adminService.DeleteAdmin(c.Request.Context(), id)
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
	})
}

// GetCurrentAdmin 获取当前登录管理员信息
func (h *AdminHandler) GetCurrentAdmin(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(401, gin.H{
			"code":    apperrors.ErrUnauthorized,
			"message": "unauthorized",
		})
		return
	}

	adm, err := h.adminService.GetAdminByID(c.Request.Context(), adminID.(uint64))
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
		"data":    adm,
	})
}
