package database

import (
	"fmt"
	"time"

	"github.com/zqdfound/go-uni-pay/internal/domain/entity"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Config 数据库配置
type Config struct {
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
}

// Init 初始化数据库连接
func Init(cfg Config) error {
	var err error

	// 连接数据库
	DB, err = gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 获取底层的sql.DB
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// 自动迁移（可选，生产环境建议使用SQL脚本）
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	return nil
}

// autoMigrate 自动迁移表结构
func autoMigrate() error {
	return DB.AutoMigrate(
		&entity.User{},
		&entity.PaymentConfig{},
		&entity.PaymentOrder{},
		&entity.PaymentLog{},
		&entity.APILog{},
		&entity.NotifyQueue{},
	)
}

// Close 关闭数据库连接
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
