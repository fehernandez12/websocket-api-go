package cache

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	redis *redis.Client
}

func NewRedisCacheRepository() *RedisCache {
	return &RedisCache{
		redis: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}),
	}
}

func (r *RedisCache) Get(ctx context.Context, key string) (*string, error) {
	var dest string
	val, err := r.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, errors.New("key does not exist")
	}
	err = json.Unmarshal([]byte(val), &dest)
	if err != nil {
		return nil, err
	}
	return &dest, nil
}

func (r *RedisCache) Put(ctx context.Context, key string, value any) error {
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.redis.Set(ctx, key, val, 0).Err()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.redis.Del(ctx, key).Err()
}
