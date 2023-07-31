package cache

import (
	"context"
	"github.com/chenyahui/gin-cache/persist"
	"os"
	"strconv"
	"web-api/util"

	"github.com/go-redis/redis/v8"
)

// RedisClient Redis缓存客户端单例
var RedisClient *redis.Client
var RedisLogClient *redis.Client
var RedisSessionClient *redis.Client
var RedisShareClient *redis.Client
var RedisStore *persist.RedisStore

// Redis 在中间件中初始化redis链接
func Redis() {
	db, _ := strconv.ParseUint(os.Getenv("REDIS_DB"), 10, 64)
	client := redis.NewClient(&redis.Options{
		Addr:       os.Getenv("REDIS_ADDR"),
		Password:   os.Getenv("REDIS_PW"),
		DB:         int(db),
		MaxRetries: 1,
	})

	if _, err := client.Ping(context.TODO()).Result(); err != nil {
		util.Log().Panic("连接Redis不成功", err)
	}

	RedisClient = client
}

func RedisSession() {
	db, _ := strconv.ParseUint(os.Getenv("REDIS_SESSION_DB"), 10, 64)
	client := redis.NewClient(&redis.Options{
		Addr:       os.Getenv("REDIS_ADDR"),
		Password:   os.Getenv("REDIS_PW"),
		DB:         int(db),
		MaxRetries: 1,
	})

	if _, err := client.Ping(context.TODO()).Result(); err != nil {
		util.Log().Panic("Failed to connect redis session db", err)
	}

	RedisSessionClient = client
}

func RedisLog() {
	db, _ := strconv.ParseUint(os.Getenv("REDIS_LOG_DB"), 10, 64)
	client := redis.NewClient(&redis.Options{
		Addr:       os.Getenv("REDIS_ADDR"),
		Password:   os.Getenv("REDIS_PW"),
		DB:         int(db),
		MaxRetries: 1,
	})

	if _, err := client.Ping(context.TODO()).Result(); err != nil {
		util.Log().Panic("连接Redis 2不成功", err)
	}

	RedisLogClient = client
}

func RedisShare() {
	db, _ := strconv.ParseUint(os.Getenv("REDIS_SHARE_DB"), 10, 64)
	client := redis.NewClient(&redis.Options{
		Addr:       os.Getenv("REDIS_ADDR"),
		Password:   os.Getenv("REDIS_PW"),
		DB:         int(db),
		MaxRetries: 1,
	})

	if _, err := client.Ping(context.TODO()).Result(); err != nil {
		util.Log().Panic("连接Redis 2不成功", err)
	}

	RedisShareClient = client
}

func SetupRedisStore() {
	RedisStore = persist.NewRedisStore(RedisClient)
}
