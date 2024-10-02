package task

import (
	"blgit.rfdev.tech/taya/game-service/imsb/callback"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"web-api/cache"
	"web-api/service/common"
	"web-api/service/imsb"
	"web-api/util"
)

func ProcessImUpdateBalance(ctx context.Context) {
	for {
		iter := cache.RedisSyncTransactionClient.Scan(ctx, 0, "im:*", 0).Iterator()
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
			time.Sleep(1 * time.Second)
			go func(key string, arr []string) {
				defer wg.Done()
				for _, a := range arr {
					redisKey := fmt.Sprintf(`im:%s:%s`, key, a)
					v := cache.RedisSyncTransactionClient.Get(context.TODO(), redisKey)
					var data callback.WagerDetail
					err := json.Unmarshal([]byte(v.Val()), &data)
					if err != nil {
						util.Log().Error("Task:ProcessImUpdateBalance error", err, data)
						continue
					}

					fmt.Println("DebugLog1234: Request.ActionId", data.ActionId)
					fmt.Println("DebugLog1234: Request.MemberCode", data.MemberCode)
					fmt.Println("DebugLog1234: Request.WagerNo", data.WagerNo)
					fmt.Println("DebugLog1234: Request.TransactionAmount", data.TransactionAmount)
					fmt.Println("DebugLog1234: Request.SourceWallet", data.SourceWallet)

					err = common.ProcessImUpdateBalanceTransaction(ctx, &imsb.TransactionBuilder{Request: data})
					if err != nil {
						util.Log().Error("Task:ProcessImUpdateBalance error", err, data)
						return
					}

					_, err = cache.RedisSyncTransactionClient.Del(context.TODO(), redisKey).Result()
					if err != nil {
						util.Log().Error("Task:ProcessImUpdateBalance redis delete key error", err, data)
						return
					}
				}
			}(key, arr)
		}
		wg.Wait()
		time.Sleep(1 * time.Second)
	}
}
