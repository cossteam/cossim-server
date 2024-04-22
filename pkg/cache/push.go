package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

const (
	PushExpireTime = 12 * time.Hour
	PushKeyPrefix  = "push:"
)

type PushCache interface {
	DeleteAllCache(ctx context.Context) error
}

func NewPushCacheRedis(addr, password string, db int) (*PushCacheRedis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &PushCacheRedis{
		client: client,
	}, nil
}

var _ PushCache = &PushCacheRedis{}

type PushCacheRedis struct {
	client *redis.Client
}

func (p *PushCacheRedis) DeleteAllCache(ctx context.Context) error {
	keys := make([]string, 0)
	iter := p.client.Scan(ctx, 0, PushKeyPrefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return p.client.Del(ctx, keys...).Err()
}
