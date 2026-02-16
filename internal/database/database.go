package database

import (
	"fmt"

	"github.com/joefazee/learning-go-shop/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(cfg config.DatabaseConfig) (*gorm.DB, error) {
	// ✅ ถ้ามี URL/DSN ให้ใช้ก่อน (เหมาะกับ deploy)
	dsn := cfg.DSN
	if dsn == "" {
		// ✅ fallback: ประกอบจาก DB_* แบบเดิม
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
			cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode, cfg.TimeZone,
		)
	} 

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}
