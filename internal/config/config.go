package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	AWS      AWSConfig
	Upload   UploadConfig
}

type ServerConfig struct {
	Port    string
	GinMode string
}

type DatabaseConfig struct {
	// ✅ รองรับแบบ DB_* (local/dev)
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string

	// ✅ รองรับแบบ DATABASE_URL / DSN (prod/deploy)
	// ตัวอย่าง:
	// postgres://postgres:password@localhost:5432/ecommerce_shop?sslmode=disable
	DSN string
}

type JWTConfig struct {
	Secret              string
	ExpiresIn           time.Duration
	RefreshTokenExpires time.Duration
}

type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	S3Bucket        string
	S3Endpoint      string
}

type UploadConfig struct {
	Path        string
	MaxFileSize int64

	// ✅ เพิ่มจากไฟล์ล่าง: เลือก provider (local | s3)
	UploadProvider string
}

func Load() (*Config, error) {
	// ✅ โหลด .env ถ้ามี (ถ้าไม่มีไม่ error)
	_ = godotenv.Load()

	jwtExpiresIn := mustParseDuration(getEnv("JWT_EXPIRES_IN", "24h"))
	refreshTokenExpires := mustParseDuration(getEnv("REFRESH_TOKEN_EXPIRES_IN", "720h"))
	maxUploadSize := mustParseInt64(getEnv("MAX_UPLOAD_SIZE", "10485760"), 10, 64)

	cfg := &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),

			// ✅ เปลี่ยน default ตรงนี้ให้เป็น DB ที่คุณสร้างไว้
			Name: getEnv("DB_NAME", "ecommerce_shop"),

			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "UTC"),

			// ✅ ถ้ามี DATABASE_URL จะใช้ก่อน (ผ่าน database.New)
			DSN: firstNonEmpty(
				os.Getenv("DATABASE_URL"),
				os.Getenv("DB_DSN"),
			),
		},
		JWT: JWTConfig{
			Secret:              getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
			ExpiresIn:           jwtExpiresIn,
			RefreshTokenExpires: refreshTokenExpires,
		},
		AWS: AWSConfig{
			Region:          getEnv("AWS_REGION", "us-east-1"),
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", "test"),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", "test"),
			S3Bucket:        getEnv("AWS_S3_BUCKET", "ecommerce-uploads"),
			S3Endpoint:      getEnv("AWS_S3_ENDPOINT", "http://localhost:4566"),
		},
		Upload: UploadConfig{
			Path:        getEnv("UPLOAD_PATH", "./uploads"),
			MaxFileSize: maxUploadSize,

			// ✅ เพิ่มจากไฟล์ล่าง
			UploadProvider: getEnv("UPLOAD_PROVIDER", "local"),
		},
	}

	// ✅ validation กัน config หลุด ๆ
	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Server.Port == "" {
		return errors.New("config: PORT is required")
	}

	// ✅ validate upload provider
	switch cfg.Upload.UploadProvider {
	case "", "local", "s3":
		// ok (ถ้าว่างถือว่า ok เพราะ default มักเป็น local)
	default:
		return fmt.Errorf("config: UPLOAD_PROVIDER must be 'local' or 's3' (got %q)", cfg.Upload.UploadProvider)
	}

	// ถ้ามี DSN แล้วไม่ต้องบังคับ DB_* ทุกตัว
	if cfg.Database.DSN != "" {
		return nil
	}

	// ถ้าไม่มี DSN ต้องมีค่าหลัก ๆ ให้ครบ
	if cfg.Database.Host == "" {
		return errors.New("config: DB_HOST is required (or set DATABASE_URL)")
	}
	if cfg.Database.Port == "" {
		return errors.New("config: DB_PORT is required (or set DATABASE_URL)")
	}
	if cfg.Database.User == "" {
		return errors.New("config: DB_USER is required (or set DATABASE_URL)")
	}
	if cfg.Database.Name == "" {
		return errors.New("config: DB_NAME is required (or set DATABASE_URL)")
	}
	if cfg.Database.SSLMode == "" {
		return errors.New("config: DB_SSL_MODE is required (or set DATABASE_URL)")
	}
	if cfg.Database.TimeZone == "" {
		cfg.Database.TimeZone = "UTC"
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func mustParseDuration(v string) time.Duration {
	d, err := time.ParseDuration(v)
	if err != nil {
		// ถ้าพิมพ์ผิด ให้ fallback แบบปลอดภัย
		return 24 * time.Hour
	}
	return d
}

func mustParseInt64(s string, base int, bitSize int) int64 {
	n, err := strconv.ParseInt(s, base, bitSize)
	if err != nil {
		// fallback 10MB
		return 10 * 1024 * 1024
	}
	return n
}

// (optional) helper ถ้าอยากพิมพ์ debug ได้ง่าย
func (d DatabaseConfig) DebugString() string {
	if d.DSN != "" {
		return "DATABASE_URL/DB_DSN is set"
	}
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s tz=%s",
		d.Host, d.Port, d.User, d.Name, d.SSLMode, d.TimeZone,
	)
}
