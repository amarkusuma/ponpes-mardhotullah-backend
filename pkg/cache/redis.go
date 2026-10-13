package cache

import (
	"github.com/redis/go-redis/v9"
	"github.com/simpeg/backend/internal/config"
)

func NewRedis(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
}
