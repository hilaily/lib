package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type IRedis interface {
	GetClient() *redis.Client
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type IConfig interface {
	Unmarshal(ptr any) error
}

func NewFromConfig(ctx context.Context, conf IConfig) (*_redis, error) {
	var config RedisConfig
	err := conf.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("get redis config failed: %w", err)
	}
	return NewRedis(ctx, &config)
}

func NewRedis(ctx context.Context, config *RedisConfig) (*_redis, error) {
	if config.Addr == "" {
		return nil, fmt.Errorf("addr is empty")
	}
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})
	return &_redis{client: client}, nil
}

type _redis struct {
	client *redis.Client
}

func (r *_redis) GetClient() *redis.Client {
	return r.client
}
