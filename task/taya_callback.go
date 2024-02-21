package task

import (
	"blgit.rfdev.tech/taya/game-service/fb/callback"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/service/common"
	"web-api/service/taya"
	"web-api/util"
	"web-api/util/i18n"
)

func ProcessTayaSyncTransaction() {
	for {
		ctx := context.TODO()
		iter := cache.RedisSyncTransactionClient.Scan(ctx, 0, "taya:*", 0).Iterator()
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
					redisKey := fmt.Sprintf(`taya:%s:%s`, key, a)
					v := cache.RedisSyncTransactionClient.Get(context.TODO(), redisKey)
					var orderPayRequest callback.OrderPayRequest
					err := json.Unmarshal([]byte(v.Val()), &orderPayRequest)
					if err != nil {
						util.Log().Error("Task:ProcessTayaSyncTransaction error", err, orderPayRequest)
						continue
					}

					err = common.ProcessTransaction(&taya.Callback{Request: orderPayRequest})
					if err != nil {
						util.Log().Error("Task:ProcessTayaSyncTransaction error", err, orderPayRequest)
						return
					}

					fmt.Println("Task:ProcessTayaSyncTransaction deleting redis key", redisKey)
					_, err = cache.RedisSyncTransactionClient.Del(context.TODO(), redisKey).Result()
					if err != nil {
						util.Log().Error("Task:ProcessTayaSyncTransaction redis delete key error", err, orderPayRequest)
						return
					}
					fmt.Println("Task:ProcessTayaSyncTransaction deleted redis key", redisKey)

					//go sendSettlementNotification(orderPayRequest.MerchantUserId)
				}
			}(key, arr)
		}
		wg.Wait()
		time.Sleep(1 * time.Second)
	}
}

func sendSettlementNotification(username string) {
	var user ploutos.User
	e := model.DB.Where(`username`, username).First(&user).Error
	if e == nil {
		i18n := i18n.I18n{}
		if e = i18n.LoadLanguages("en"); e == nil {
			common.SendNotification(user.ID, consts.Notification_Type_Bet_Settlement, i18n.T("notification_bet_settlement_title"), i18n.T("notification_bet_settlement"))
		}
	}
}
