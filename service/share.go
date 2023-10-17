package service

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"time"
	"web-api/cache"
	"web-api/serializer"
)

const (
	RedisKeyShare = "share:"
)

type CreateShareService struct {
	Path   string `form:"path" json:"path" binding:"required"`
	Params string `form:"params" json:"params" binding:"required"`
}

func (service *CreateShareService) Create() (r serializer.Response, err error) {
	v := make(map[string]interface{})
	var j map[string]interface{}
	if err = json.Unmarshal([]byte(service.Params), &j); err == nil {
		v["path"] = service.Path
		v["params"] = j

		rstr := randStr(6)
		var js []byte
		js, err = json.Marshal(v)
		if err != nil {
			return
		}
		re := cache.RedisShareClient.Set(context.TODO(), RedisKeyShare+rstr, js, 168*time.Hour)
		if re.Err() != nil {
			err = re.Err()
			return
		}

		r = serializer.Response{
			Data: rstr,
		}
	}
	return
}

type GetShareService struct {
	Key string `form:"key" json:"key" binding:"required"`
}

func (service *GetShareService) Get() (r serializer.Response, err error) {
	re := cache.RedisShareClient.Get(context.TODO(), RedisKeyShare+service.Key)
	if re.Err() == redis.Nil {
		re = cache.RedisClient.Get(context.TODO(), RedisKeyShare+service.Key)
	}
	if re.Err() != nil {
		var a []int
		r = serializer.Response{
			Data: a,
		}
		return
	}
	var j map[string]interface{}
	err = json.Unmarshal([]byte(re.Val()), &j)
	if err != nil {
		return
	}
	r = serializer.Response{
		Data: j,
	}
	return
}

func randStr(n int) string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
