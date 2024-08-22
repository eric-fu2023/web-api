package util

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func FindFromRedis[T any](c *gin.Context, client *redis.Client, key string, obj *T) {
	cacheData := client.Get(context.TODO(), key)
	if cacheData.Err() != nil {
		if cacheData.Err() != redis.Nil {
			GetLoggerEntry(c).Warn("Failed to retrieve from Redis", cacheData.Err())
		}
		return
	}
	// deserialize into map
	err := json.Unmarshal([]byte(cacheData.Val()), &obj)
	if err != nil {
		GetLoggerEntry(c).Warn("Failed to deserializ json", err.Error())
	}
}

func CacheIntoRedis[T any](c *gin.Context, client *redis.Client, key string, expiration time.Duration, obj *T) {
	if jsonStr, err := json.Marshal(&obj); err == nil {
		if res := client.Set(context.TODO(), key, jsonStr, expiration); res.Err() != nil {
			GetLoggerEntry(c).Warn("Failed to cache into Redis", res.Err())
		}
	} else {
		GetLoggerEntry(c).Warn("Failed to serialize json", err.Error())
	}
}
