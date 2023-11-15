package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"web-api/cache"
	"web-api/model"

	"github.com/go-redis/redis/v8"
)

const (
	CachedConfigPrefix string = "cached_config"
)

func GetCachedConfig(ctx context.Context, key string) (value string, err error) {
	value, err = cache.RedisConfigClient.Get(ctx, fmt.Sprintf("%s:%s", CachedConfigPrefix, key)).Result()
	if errors.Is(err, redis.Nil) {
		value, err = getConfigFromDB(key)
		_ = cache.RedisConfigClient.Set(ctx, fmt.Sprintf("%s:%s", CachedConfigPrefix, key), value, 10*time.Minute)
	}
	return
}

func getConfigFromDB(key string) (value string, err error) {
	err = model.DB.Table("app_configs").Select("value").Where("key", key).Scan(&value).Error
	return
}
