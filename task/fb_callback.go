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
	"web-api/service/common"
	"web-api/service/fb"
	"web-api/util"
)

func ProcessFbSyncTransaction() {
	for {
		ctx := context.TODO()
		iter := cache.RedisSyncTransactionClient.Scan(ctx, 0, "fb:*", 0).Iterator()
		keys := make(map[string][]string)
		for iter.Next(ctx) {
			str := strings.Split(iter.Val(), ":")
			if len(str) == 3 {
				arr, ok := keys[str[1]]
				if !ok {
					keys[str[1]] = make([]string, 0)
				}
				keys[str[1]] = append(arr, str[2])
			}
		}

		var wg sync.WaitGroup
		for key, arr := range keys {
			sort.Strings(arr)
			wg.Add(1)
			go func(key string, arr []string) {
				defer wg.Done()
				for _, a := range arr {
					redisKey := fmt.Sprintf(`fb:%s:%s`, key, a)
					v := cache.RedisSyncTransactionClient.Get(context.TODO(), redisKey)
					var orderPayRequest callback.OrderPayRequest
					err := json.Unmarshal([]byte(v.Val()), &orderPayRequest)
					if err != nil {
						util.Log().Error("Task:ProcessFbSyncTransaction error", err, orderPayRequest)
						continue
					}

					err = common.ProcessTransaction(&fb.Callback{Request: orderPayRequest})
					if err != nil {
						util.Log().Error("Task:ProcessFbSyncTransaction error", err, orderPayRequest)
						return
					}

					fmt.Println("Task:ProcessFbSyncTransaction deleting redis key", redisKey)
					_, err = cache.RedisSyncTransactionClient.Del(context.TODO(), redisKey).Result()
					if err != nil {
						util.Log().Error("Task:ProcessFbSyncTransaction redis delete key error", err, orderPayRequest)
						return
					}
					fmt.Println("Task:ProcessFbSyncTransaction deleted redis key", redisKey)

					//go sendSettlementNotification(orderPayRequest.MerchantUserId)
				}
			}(key, arr)
		}
		wg.Wait()
		time.Sleep(1 * time.Second)
	}
}
