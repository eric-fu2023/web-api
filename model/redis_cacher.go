package model

import (
	"context"
	"github.com/go-gorm/caches/v3"
	"github.com/go-redis/redis/v8"
	"time"
)

const (
	KeyCacheInfo = "_cache_info"
)

type RedisCacher struct {
	Redis *redis.Client
}

type CacheInfo struct {
	Prefix string
	Ttl    int64
}

func (c *RedisCacher) Get(ctx context.Context, key string, q *caches.Query[any]) (*caches.Query[any], error) {
	cacheInfo, exists := getCacheInfo(ctx)
	if !exists {
		return nil, nil
	}
	key = cacheInfo.Prefix + key
	res, err := c.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err = q.Unmarshal([]byte(res)); err != nil {
		return nil, err
	}
	return q, nil
}

func (c *RedisCacher) Store(ctx context.Context, key string, val *caches.Query[any]) error {
	cacheInfo, exists := getCacheInfo(ctx)
	if !exists {
		return nil
	}
	key = cacheInfo.Prefix + key
	res, err := val.Marshal()
	if err != nil {
		return err
	}
	var ttl int64 = 60
	if cacheInfo.Ttl != 0 {
		ttl = cacheInfo.Ttl
	}
	c.Redis.Set(ctx, key, res, time.Duration(ttl)*time.Second)
	return nil
}

func getCacheInfo(ctx context.Context) (cacheInfo CacheInfo, exists bool) {
	ci := ctx.Value(KeyCacheInfo)
	if ci == nil {
		return
	}
	cacheInfo, exists = ci.(CacheInfo)
	return
}
