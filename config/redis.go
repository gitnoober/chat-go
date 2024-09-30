package config

import (
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

func loadRedisConfig() *RedisConfig {
	return &RedisConfig{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	}
}

func ConnectRedis(cfg *Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.RedisConfig.Addr,
		Password: cfg.RedisConfig.Password,
		DB:       cfg.RedisConfig.DB,
	})
}
