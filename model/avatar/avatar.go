package avatar

import (
	"context"
	"math"
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

func GetAvatarUrlListTeamup() []string {
	lName := os.Getenv("AVATAR_URL_LIST_NAME") // todo lift up to Init
	switch lName {
	case "2":
		return bataceTeamupAvatarList
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
	if rand.Float64() < 0.5 {
		return defaultBaAvatar // TODO : move defaultBaAvatar up to params
	}
	pool := GetRotatingAvatarPool(GetAvatarUrlListTeamup(), 8)
	rndIdx := rand.Intn(8)
	return pool[rndIdx]
}

func GetAvatarPoolWithPercentDefault(percentDefault float64, poolSize int) []string {
	percentNonDefault := float64(1.0) - percentDefault
	numRealImages := math.Floor(float64(poolSize) * percentNonDefault)

	pool := GetRotatingAvatarPool(GetAvatarUrlListTeamup(), int(numRealImages))
	poolCopy := make([]string, len(pool))
	copy(poolCopy, pool)
	
	for i := 0; i < poolSize - int(numRealImages); i++ {
		poolCopy = append(poolCopy, defaultBaAvatar) // TODO : move defaultBaAvatar up to params
	}
	
	rand.Shuffle(len(poolCopy), func(i, j int) {
		poolCopy[i], poolCopy[j] = poolCopy[j], poolCopy[i]
	})
	return poolCopy
}

func GetAvatarPoolWithMaxReal(maxReal, poolSize int, allReal bool) []string {
	numReal := poolSize
	if !allReal {
		numReal = rand.Intn(maxReal + 1)
	}

	if numReal > poolSize {
		numReal = poolSize
	}

	pool := GetRotatingAvatarPool(GetAvatarUrlListTeamup(), numReal)
	poolCopy := make([]string, len(pool))
	copy(poolCopy, pool)

	for i := 0; i < poolSize - numReal; i++ {
		poolCopy = append(poolCopy, defaultBaAvatar) // TODO : move defaultBaAvatar up to params
	}

	rand.Shuffle(len(poolCopy), func(i, j int) {
		poolCopy[i], poolCopy[j] = poolCopy[j], poolCopy[i]
	})
	return poolCopy
}