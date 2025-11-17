package payment

import (
	"fmt"
	"sync"

	apperrors "github.com/zqdfound/go-uni-pay/pkg/errors"
)

var (
	registry = &Registry{
		providers: make(map[string]Provider),
	}
)

// Registry 支付提供商注册表
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// Register 注册支付提供商
func Register(provider Provider) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	name := provider.GetName()
	registry.providers[name] = provider
}

// GetProvider 获取支付提供商
func GetProvider(name string) (Provider, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	provider, ok := registry.providers[name]
	if !ok {
		return nil, apperrors.New(apperrors.ErrProviderNotFound, fmt.Sprintf("provider %s not found", name))
	}

	return provider, nil
}

// GetAllProviders 获取所有支付提供商
func GetAllProviders() map[string]Provider {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	providers := make(map[string]Provider, len(registry.providers))
	for name, provider := range registry.providers {
		providers[name] = provider
	}

	return providers
}

// HasProvider 检查提供商是否存在
func HasProvider(name string) bool {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	_, ok := registry.providers[name]
	return ok
}
