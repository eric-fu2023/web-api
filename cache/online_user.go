package cache

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func SetUserOnline(ctx *gin.Context, userId string, page string, status bool) (error) {
	// detail := OnlineStatusInfo{
	// 	Id: userId,
	// 	Page: page,
	// 	Status: status,

	// }

	val := map[string]interface{}{
		"id":    userId,
		"page": page,
		"status": status,
	}
	key := fmt.Sprintf("online_user:%s", userId)

	res := RedisSessionClient.HSet(ctx, key, val)
	if res.Err() != nil && res.Err() != redis.Nil {
		return res.Err()
	}
	RedisSessionClient.Expire(ctx, key, 12*time.Second)
	return nil 
}