package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RedisLock Redis分布式锁
type RedisLock struct {
	client *redis.Client
	key    string
	value  string
	ttl    time.Duration
}

// NewRedisLock 创建分布式锁
func NewRedisLock(client *redis.Client, key string, ttl time.Duration) *RedisLock {
	return &RedisLock{
		client: client,
		key:    fmt.Sprintf("lock:%s", key),
		value:  uuid.New().String(),
		ttl:    ttl,
	}
}

// Lock 获取锁
func (l *RedisLock) Lock(ctx context.Context) (bool, error) {
	result, err := l.client.SetNX(ctx, l.key, l.value, l.ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}
	return result, nil
}

// TryLock 尝试获取锁（带重试）
func (l *RedisLock) TryLock(ctx context.Context, maxRetries int, retryInterval time.Duration) error {
	for i := 0; i < maxRetries; i++ {
		acquired, err := l.Lock(ctx)
		if err != nil {
			return err
		}
		if acquired {
			return nil
		}

		// 等待后重试
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryInterval):
			// 继续重试
		}
	}
	return fmt.Errorf("failed to acquire lock after %d retries", maxRetries)
}

// Unlock 释放锁
func (l *RedisLock) Unlock(ctx context.Context) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if result.(int64) == 0 {
		return fmt.Errorf("lock not held or already expired")
	}

	return nil
}

// Refresh 刷新锁的过期时间
func (l *RedisLock) Refresh(ctx context.Context) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value, int(l.ttl.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("failed to refresh lock: %w", err)
	}

	if result.(int64) == 0 {
		return fmt.Errorf("lock not held or already expired")
	}

	return nil
}

// WithLock 使用锁执行函数
func WithLock(ctx context.Context, client *redis.Client, key string, ttl time.Duration, fn func() error) error {
	lock := NewRedisLock(client, key, ttl)

	// 尝试获取锁
	if err := lock.TryLock(ctx, 3, 100*time.Millisecond); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	// 确保释放锁
	defer func() {
		if err := lock.Unlock(context.Background()); err != nil {
			// 记录错误但不返回
			fmt.Printf("failed to unlock: %v\n", err)
		}
	}()

	// 执行函数
	return fn()
}
