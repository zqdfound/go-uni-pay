package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zqdfound/go-uni-pay/internal/domain/entity"
	"github.com/zqdfound/go-uni-pay/internal/domain/repository"
	"github.com/zqdfound/go-uni-pay/internal/infrastructure/cache"
	apperrors "github.com/zqdfound/go-uni-pay/pkg/errors"
	"github.com/zqdfound/go-uni-pay/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Service 认证服务
type Service struct {
	userRepo repository.UserRepository
}

// NewService 创建认证服务
func NewService(userRepo repository.UserRepository) *Service {
	return &Service{
		userRepo: userRepo,
	}
}

// CreateUser 创建用户
func (s *Service) CreateUser(ctx context.Context, username, email string) (*entity.User, string, error) {
	// 生成API Key
	apiKey := s.generateAPIKey()

	// 生成API Secret
	apiSecret := s.generateAPISecret()

	// 加密API Secret
	hashedSecret, err := s.hashSecret(apiSecret)
	if err != nil {
		return nil, "", err
	}

	// 创建用户
	user := &entity.User{
		Username:  username,
		Email:     email,
		APIKey:    apiKey,
		APISecret: hashedSecret,
		Status:    1,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	// 返回用户和明文secret（仅此一次）
	return user, apiSecret, nil
}

// ValidateAPIKey 验证API Key（带缓存）
func (s *Service) ValidateAPIKey(ctx context.Context, apiKey string) (*entity.User, error) {
	// 构造缓存key
	cacheKey := fmt.Sprintf("auth:user:%s", apiKey)

	// 先尝试从缓存获取
	cached, err := cache.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var user entity.User
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			// 检查用户状态
			if user.Status != 1 {
				return nil, apperrors.New(apperrors.ErrForbidden, "user is disabled")
			}
			logger.Debug("api key cache hit", zap.String("api_key", apiKey))
			return &user, nil
		}
	}

	// 缓存未命中，从数据库查询
	user, err := s.userRepo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrUnauthorized, "invalid api key")
	}

	if user.Status != 1 {
		return nil, apperrors.New(apperrors.ErrForbidden, "user is disabled")
	}

	// 写入缓存，有效期10分钟
	if userJSON, err := json.Marshal(user); err == nil {
		cache.Set(ctx, cacheKey, string(userJSON), 10*time.Minute)
		logger.Debug("api key cached", zap.String("api_key", apiKey))
	}

	return user, nil
}

// InvalidateUserCache 使用户缓存失效
// 当用户信息更新时应该调用此方法
func (s *Service) InvalidateUserCache(ctx context.Context, apiKey string) error {
	cacheKey := fmt.Sprintf("auth:user:%s", apiKey)
	return cache.Del(ctx, cacheKey)
}

// ValidateAPIKeyAndSecret 验证API Key和Secret
func (s *Service) ValidateAPIKeyAndSecret(ctx context.Context, apiKey, apiSecret string) (*entity.User, error) {
	user, err := s.userRepo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrUnauthorized, "invalid api key or secret")
	}

	if user.Status != 1 {
		return nil, apperrors.New(apperrors.ErrForbidden, "user is disabled")
	}

	// 验证secret
	if err := s.verifySecret(apiSecret, user.APISecret); err != nil {
		return nil, apperrors.New(apperrors.ErrUnauthorized, "invalid api key or secret")
	}

	return user, nil
}

// generateAPIKey 生成API Key
func (s *Service) generateAPIKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("ak_%s", hex.EncodeToString(b))
}

// generateAPISecret 生成API Secret
func (s *Service) generateAPISecret() string {
	b := make([]byte, 64)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// hashSecret 加密Secret
func (s *Service) hashSecret(secret string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// verifySecret 验证Secret
func (s *Service) verifySecret(secret, hashedSecret string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedSecret), []byte(secret))
}
