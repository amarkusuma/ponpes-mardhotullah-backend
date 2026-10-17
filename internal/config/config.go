package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// AppEnv: "development" atau "production". Default "production" agar
	// lupa men-set env tidak pernah memicu seeding data dummy di server.
	AppEnv string

	ServerAddress   string
	DatabaseURL     string
	RedisAddr       string
	RedisPassword   string
	JWTSecret       string
	LogLevel        string
	AllowedOrigins  string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool

	// Akun admin awal untuk production. Hanya dipakai bila belum ada satu pun
	// user ber-role admin; dikosongkan berarti tidak membuat admin.
	AdminName     string
	AdminEmail    string
	AdminPassword string
}

// IsProduction bernilai true kecuali APP_ENV secara eksplisit "development".
func (c *Config) IsProduction() bool {
	return c.AppEnv != "development"
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		AppEnv: strings.ToLower(strings.TrimSpace(getEnv("APP_ENV", "production"))),

		ServerAddress:   getEnv("SERVER_ADDRESS", ":8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/simpeg?sslmode=disable"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		JWTSecret:       getEnv("JWT_SECRET", "change-me-super-secret"),
		LogLevel:        getEnv("LOG_LEVEL", "debug"),
		AllowedOrigins:  getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		AccessTokenTTL:  getDurationEnv("ACCESS_TOKEN_TTL", 6*time.Hour),
		RefreshTokenTTL: getDurationEnv("REFRESH_TOKEN_TTL", 7*24*time.Hour),

		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:    getEnv("MINIO_BUCKET", "ponpes"),
		MinioUseSSL:    getBoolEnv("MINIO_USE_SSL", false),

		AdminName:     getEnv("ADMIN_NAME", "Administrator"),
		AdminEmail:    getEnv("ADMIN_EMAIL", ""),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return fallback
}
