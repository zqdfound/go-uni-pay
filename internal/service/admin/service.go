package admin

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zqdfound/go-uni-pay/internal/domain/entity"
	"github.com/zqdfound/go-uni-pay/internal/domain/repository"
	"github.com/zqdfound/go-uni-pay/internal/infrastructure/config"
	apperrors "github.com/zqdfound/go-uni-pay/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// Service 管理员服务
type Service struct {
	adminRepo repository.AdminRepository
	jwtSecret string
	jwtExpire time.Duration
}

// NewService 创建管理员服务
func NewService(adminRepo repository.AdminRepository, cfg *config.Config) *Service {
	return &Service{
		adminRepo: adminRepo,
		jwtSecret: cfg.JWT.Secret,
		jwtExpire: cfg.JWT.GetJWTExpire(),
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string        `json:"token"`
	Admin     *entity.Admin `json:"admin"`
	ExpiresAt int64         `json:"expires_at"`
}

// Login 管理员登录
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// 查找管理员
	admin, err := s.adminRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrInvalidCredentials, "invalid username or password")
	}

	// 检查状态
	if admin.Status != 1 {
		return nil, apperrors.New(apperrors.ErrInvalidCredentials, "admin account is disabled")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return nil, apperrors.New(apperrors.ErrInvalidCredentials, "invalid username or password")
	}

	// 生成JWT Token
	token, expiresAt, err := s.generateToken(admin.ID, admin.Username)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternalServer, "failed to generate token", err)
	}

	// 更新最后登录时间
	now := time.Now()
	admin.LastLogin = &now
	if err := s.adminRepo.Update(ctx, admin); err != nil {
		// 更新失败不影响登录，只记录错误
		// TODO: 添加日志
	}

	return &LoginResponse{
		Token:     token,
		Admin:     admin,
		ExpiresAt: expiresAt,
	}, nil
}

// CreateAdminRequest 创建管理员请求
type CreateAdminRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
}

// CreateAdmin 创建管理员
func (s *Service) CreateAdmin(ctx context.Context, req *CreateAdminRequest) (*entity.Admin, error) {
	// 检查用户名是否已存在
	if _, err := s.adminRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, apperrors.New(apperrors.ErrInvalidParam, "username already exists")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternalServer, "failed to hash password", err)
	}

	admin := &entity.Admin{
		Username: req.Username,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Email:    req.Email,
		Status:   1,
	}

	if err := s.adminRepo.Create(ctx, admin); err != nil {
		return nil, err
	}

	return admin, nil
}

// UpdateAdminRequest 更新管理员请求
type UpdateAdminRequest struct {
	ID       uint64 `json:"id" binding:"required"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Status   *int8  `json:"status"`
	Password string `json:"password"`
}

// UpdateAdmin 更新管理员
func (s *Service) UpdateAdmin(ctx context.Context, req *UpdateAdminRequest) error {
	admin, err := s.adminRepo.GetByID(ctx, req.ID)
	if err != nil {
		return err
	}

	if req.Nickname != "" {
		admin.Nickname = req.Nickname
	}
	if req.Email != "" {
		admin.Email = req.Email
	}
	if req.Status != nil {
		admin.Status = *req.Status
	}
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return apperrors.Wrap(apperrors.ErrInternalServer, "failed to hash password", err)
		}
		admin.Password = string(hashedPassword)
	}

	return s.adminRepo.Update(ctx, admin)
}

// GetAdminByID 根据ID获取管理员
func (s *Service) GetAdminByID(ctx context.Context, id uint64) (*entity.Admin, error) {
	return s.adminRepo.GetByID(ctx, id)
}

// ListAdmins 获取管理员列表
func (s *Service) ListAdmins(ctx context.Context, page, pageSize int) ([]*entity.Admin, int64, error) {
	return s.adminRepo.List(ctx, page, pageSize)
}

// DeleteAdmin 删除管理员
func (s *Service) DeleteAdmin(ctx context.Context, id uint64) error {
	return s.adminRepo.Delete(ctx, id)
}

// VerifyToken 验证JWT Token
func (s *Service) VerifyToken(tokenString string) (uint64, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.New(apperrors.ErrInvalidToken, "invalid token signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return 0, "", apperrors.New(apperrors.ErrInvalidToken, "invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		adminID := uint64(claims["admin_id"].(float64))
		username := claims["username"].(string)
		return adminID, username, nil
	}

	return 0, "", apperrors.New(apperrors.ErrInvalidToken, "invalid token claims")
}

// generateToken 生成JWT Token
func (s *Service) generateToken(adminID uint64, username string) (string, int64, error) {
	expiresAt := time.Now().Add(s.jwtExpire).Unix()
	claims := jwt.MapClaims{
		"admin_id": adminID,
		"username": username,
		"exp":      expiresAt,
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt, nil
}
