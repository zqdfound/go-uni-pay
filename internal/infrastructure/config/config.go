package config

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

var Cfg *Config

// Config 全局配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Notify   NotifyConfig   `mapstructure:"notify"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	ReadTimeout  int `mapstructure:"read_timeout"`
	WriteTimeout int `mapstructure:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"`
}

// NotifyConfig 通知配置
type NotifyConfig struct {
	RetryInterval int `mapstructure:"retry_interval"`
	MaxRetry      int `mapstructure:"max_retry"`
	WorkerCount   int `mapstructure:"worker_count"`
}

// Load 加载配置文件
func Load(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置
	if err := viper.Unmarshal(&Cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database)
}

// GetRedisAddr 获取Redis地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetServerAddr 获取服务器地址
func (c *ServerConfig) GetServerAddr() string {
	return fmt.Sprintf(":%d", c.Port)
}

// GetReadTimeout 获取读取超时时间
func (c *ServerConfig) GetReadTimeout() time.Duration {
	return time.Duration(c.ReadTimeout) * time.Second
}

// GetWriteTimeout 获取写入超时时间
func (c *ServerConfig) GetWriteTimeout() time.Duration {
	return time.Duration(c.WriteTimeout) * time.Second
}

// GetJWTExpire 获取JWT过期时间
func (c *JWTConfig) GetJWTExpire() time.Duration {
	return time.Duration(c.Expire) * time.Second
}
