package task

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"web-api/cache"
	"web-api/service"
	"web-api/service/saba"
	"web-api/util"
)

func ProcessSabaSettle() {
	for {
		ctx := context.TODO()
		iter := cache.RedisSyncTransactionClient.Scan(ctx, 0, "saba:*", 0).Iterator()
		keys := make(map[string][]string)
		for iter.Next(ctx) {
			str := strings.Split(iter.Val(), ":")
			arr, ok := keys[str[1]]
			if !ok {
				keys[str[1]] = make([]string, 0)
			}
			keys[str[1]] = append(arr, str[2])
		}

		var wg sync.WaitGroup
		for key, arr := range keys {
			sort.Strings(arr)
			wg.Add(1)
			go func(key string, arr []string) {
				defer wg.Done()
				for _, a := range arr {
					redisKey := fmt.Sprintf(`saba:%s:%s`, key, a)
					v := cache.RedisSyncTransactionClient.Get(context.TODO(), redisKey)
					var req saba.SettleRedis
					err := json.Unmarshal([]byte(v.Val()), &req)
					if err != nil {
						util.Log().Error("Task:ProcessFbSyncTransaction error", err, req)
						continue
					}
					clb := saba.Settle{Request: req.Txn, OperationId: req.OpId}
					err = service.ProcessTransaction(&clb)
					if err != nil {
						util.Log().Error("Task:ProcessFbSyncTransaction error", err, req)
						return
					}

					//_, err = cache.RedisSyncTransactionClient.Del(context.TODO(), redisKey).Result()
					if err != nil {
						util.Log().Error("Task:ProcessFbSyncTransaction redis delete key error", err, req)
						return
					}
				}
			}(key, arr)
		}
		wg.Wait()
		time.Sleep(1 * time.Second)
	}
}
