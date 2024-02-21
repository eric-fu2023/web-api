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
	CachedConfigPrefix        string = "cached_config"
	CachedConfigPrefixBranded string = "cached_config_branded"
)

func GetCachedConfigBranded(ctx context.Context, key string, brand int64) (value string, err error) {
	value, err = cache.RedisConfigClient.Get(ctx, fmt.Sprintf("%s:%d:%s", CachedConfigPrefix, brand, key)).Result()
	if errors.Is(err, redis.Nil) {
		value, err = getBrandedConfigFromDB(key, brand)
		_ = cache.RedisConfigClient.Set(ctx, fmt.Sprintf("%s:%d:%s", CachedConfigPrefix, brand, key), value, 10*time.Minute)
	}
	return
}

func GetCachedConfig(ctx context.Context, key string) (value string, err error) {
	value, err = cache.RedisConfigClient.Get(ctx, fmt.Sprintf("%s:%s", CachedConfigPrefix, key)).Result()
	if errors.Is(err, redis.Nil) {
		value, err = getConfigFromDB(key)
		_ = cache.RedisConfigClient.Set(ctx, fmt.Sprintf("%s:%s", CachedConfigPrefix, key), value, 10*time.Minute)
	}
	return
}

func getBrandedConfigFromDB(key string, brand int64) (value string, err error) {
	err = model.DB.Table("app_configs").Select("value").Where("brand_id", brand).Where("key", key).Scan(&value).Error
	return
}

func getConfigFromDB(key string) (value string, err error) {
	err = model.DB.Table("app_configs").Select("value").Where("key", key).Scan(&value).Error
	return
}
