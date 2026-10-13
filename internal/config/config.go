package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddress   string
	DatabaseURL     string
	RedisAddr       string
	RedisPassword   string
	JWTSecret       string
	LogLevel        string
	AllowedOrigins  string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		ServerAddress:   getEnv("SERVER_ADDRESS", ":8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/simpeg?sslmode=disable"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		JWTSecret:       getEnv("JWT_SECRET", "change-me-super-secret"),
		LogLevel:        getEnv("LOG_LEVEL", "debug"),
		AllowedOrigins:  getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		AccessTokenTTL:  getDurationEnv("ACCESS_TOKEN_TTL", 6*time.Hour),
		RefreshTokenTTL: getDurationEnv("REFRESH_TOKEN_TTL", 7*24*time.Hour),
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
