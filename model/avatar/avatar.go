package avatar

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"time"
	"web-api/cache"

	"github.com/go-redis/redis/v8"
)

var defaultUrlList = urlList1

var defaultBaAvatar = "https://cdn.tayalive.com/batace-img/user/default_user_image/1.png"

func GetAvatarUrls() []string {
	lName := os.Getenv("AVATAR_URL_LIST_NAME") // todo lift up to Init
	switch lName {
	case "2":
		return urlList2
	default:
		return defaultUrlList
	}
}

func GetRandomAvatarUrl() string {
	urls := GetAvatarUrls()
	n := rand.Intn(len(urls))
	return urls[n]
}

func GetRotatingAvatarPool(urls []string, poolSize int) (finalPool []string) {

	if len(urls) == 0 {
		return
	}
	if poolSize > len(urls) {
		poolSize = len(urls)
	}

	redisKey := "avatar_index"

	idx := 0

	res := cache.RedisClient.Get(context.TODO(), redisKey)
	if res.Err() == nil || res.Err() == redis.Nil {
		idx, _ = strconv.Atoi(res.Val())
	}

	arrStart := idx 
	arrEnd := arrStart + poolSize

	if arrEnd <= len(urls) {
		finalPool = urls[arrStart:arrEnd]
		idx = arrEnd % len(urls)
	} else {
		rollover := arrEnd - len(urls)
		finalPool = urls[arrStart:]
		finalPool = append(finalPool, urls[:rollover]...)
		idx = rollover
	}

	setRes := cache.RedisClient.Set(context.TODO(), redisKey, strconv.Itoa(idx), 0)
	_ = setRes

	return finalPool
} 

func GetRandomAvatarUrlForTeamup() string {
	rand.Seed(time.Now().UnixNano())
	if rand.Float64() < 0.8 {
		return defaultBaAvatar
	}
	pool := GetRotatingAvatarPool(GetAvatarUrls(), 8)
	rndIdx := rand.Intn(8)
	return pool[rndIdx]
}