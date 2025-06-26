package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type redisStore struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisStore(addr, password string, db int) (CacheStore, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &redisStore{client: rdb, ctx: ctx}, nil
}

func (r *redisStore) Set(key, value string, ttl time.Duration) error {
	return r.client.Set(r.ctx, key, value, ttl).Err()
}

func (r *redisStore) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

func (r *redisStore) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}
