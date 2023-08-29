package task

import (
	"blgit.rfdev.tech/taya/game-service/fb/callback"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/service/fb"
	"web-api/util"
)

func ProcessFbSyncTransaction() {
	for {
		ctx := context.TODO()
		iter := cache.RedisSyncTransactionClient.Scan(ctx, 0, "*", 0).Iterator()
		keys := make(map[string][]string)
		for iter.Next(ctx) {
			str := strings.Split(iter.Val(), ":")
			arr, ok := keys[str[0]]
			if !ok {
				keys[str[0]] = make([]string, 0)
			}
			keys[str[0]] = append(arr, str[1])
		}

		var wg sync.WaitGroup
		for key, arr := range keys {
			sort.Strings(arr)
			wg.Add(1)
			go func(key string, arr []string) {
				defer wg.Done()
				for _, a := range arr {
					redisKey := fmt.Sprintf(`%s:%s`, key, a)
					v := cache.RedisSyncTransactionClient.Get(context.TODO(), redisKey)
					var orderPayRequest callback.OrderPayRequest
					err := json.Unmarshal([]byte(v.Val()), &orderPayRequest)
					if err != nil {
						util.Log().Error("Task:ProcessFbSyncTransaction error", err, orderPayRequest)
						continue
					}

					gpu, err := fb.GetGameProviderUser(consts.GameProvider["fb"], orderPayRequest.MerchantUserId)
					if err != nil {
						util.Log().Error("Task:ProcessFbSyncTransaction error", err, orderPayRequest)
						return
					}

					balance, remainingWager, maxWithdrawable, err := fb.GetSums(gpu)
					if err != nil {
						util.Log().Error("Task:ProcessFbSyncTransaction error", err, orderPayRequest)
						return
					}

					err = fb.ProcessTransaction(orderPayRequest, gpu.UserId, balance, remainingWager, maxWithdrawable)
					if err != nil {
						util.Log().Error("Task:ProcessFbSyncTransaction error", err, orderPayRequest)
						return
					}

					_, err = cache.RedisSyncTransactionClient.Del(context.TODO(), redisKey).Result()
					if err != nil {
						util.Log().Error("Task:ProcessFbSyncTransaction redis delete key error", err, orderPayRequest)
						return
					}
				}
			}(key, arr)
		}
		wg.Wait()
		time.Sleep(1 * time.Second)
	}
}
