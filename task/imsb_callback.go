package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"web-api/cache"
	"web-api/service/common"
	"web-api/service/imsb"
	"web-api/util"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	"blgit.rfdev.tech/taya/game-service/imsb/callback"
)

func ProcessImUpdateBalance(ctx context.Context) {
	skipWagerCalc := os.Getenv("GAME_IMSB_SKIP_WAGER_CALCULATION_AND_SETTLEMENT") == "TRUE"
	ctx = rfcontext.AppendCallDesc(ctx, "ProcessImUpdateBalance")
	ctx = rfcontext.AppendParams(ctx, "ProcessImUpdateBalance", map[string]interface{}{
		"skipWagerCalc": skipWagerCalc,
	})

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
			go func(_ctx context.Context, key string, arr []string) {
				defer wg.Done()
				for _, a := range arr {
					redisKey := fmt.Sprintf(`im:%s:%s`, key, a)
					v := cache.RedisSyncTransactionClient.Get(context.TODO(), redisKey)
					_ctx = rfcontext.AppendParams(_ctx, key, map[string]interface{}{
						"redisKey": redisKey,
					})
					var data callback.WagerDetail
					err := json.Unmarshal([]byte(v.Val()), &data)
					if err != nil {
						_ctx = rfcontext.AppendError(_ctx, err, "marshal")
						util.Log().Error(rfcontext.Fmt(_ctx))
						continue
					}

					_ctx = rfcontext.AppendParams(_ctx, redisKey, map[string]interface{}{
						"data": data,
					})

					if skipWagerCalc {
						err = common.ProcessImUpdateBalanceTransactionWithoutWagerCalculation(_ctx, &imsb.TransactionBuilder{Request: data})
					} else {
						err = common.ProcessImUpdateBalanceTransaction(_ctx, &imsb.TransactionBuilder{Request: data})
					}

					if err != nil {
						_ctx = rfcontext.AppendError(_ctx, err, "process update balance")
						util.Log().Error(rfcontext.Fmt(_ctx))
						return
					}

					_, err = cache.RedisSyncTransactionClient.Del(context.TODO(), redisKey).Result()
					if err != nil {
						_ctx = rfcontext.AppendError(_ctx, err, "redis delete key error")
						util.Log().Error(rfcontext.Fmt(_ctx))

					}

					log.Println(rfcontext.Fmt(_ctx)) // debug
				}
			}(ctx, key, arr)
		}
		wg.Wait()
		time.Sleep(1 * time.Second)

		log.Printf(rfcontext.Fmt(ctx))
	}
}
