package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var Client *redis.Client

// Config Redis配置
type Config struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

// Init 初始化Redis连接
func Init(cfg Config) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	return nil
}

// Close 关闭Redis连接
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// Get 获取缓存
func Get(ctx context.Context, key string) (string, error) {
	return Client.Get(ctx, key).Result()
}

// Set 设置缓存
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return Client.Set(ctx, key, value, expiration).Err()
}

// Del 删除缓存
func Del(ctx context.Context, keys ...string) error {
	return Client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func Exists(ctx context.Context, keys ...string) (int64, error) {
	return Client.Exists(ctx, keys...).Result()
}

// Expire 设置过期时间
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return Client.Expire(ctx, key, expiration).Err()
}

// Incr 自增
func Incr(ctx context.Context, key string) (int64, error) {
	return Client.Incr(ctx, key).Result()
}

// Decr 自减
func Decr(ctx context.Context, key string) (int64, error) {
	return Client.Decr(ctx, key).Result()
}

// HGet 获取哈希字段
func HGet(ctx context.Context, key, field string) (string, error) {
	return Client.HGet(ctx, key, field).Result()
}

// HSet 设置哈希字段
func HSet(ctx context.Context, key string, values ...interface{}) error {
	return Client.HSet(ctx, key, values...).Err()
}

// HGetAll 获取哈希所有字段
func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return Client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func HDel(ctx context.Context, key string, fields ...string) error {
	return Client.HDel(ctx, key, fields...).Err()
}
