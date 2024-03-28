package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
	"web-api/conf"
	"web-api/model"
	"web-api/server"
	"web-api/util"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-contrib/pprof"
)

func TestMain(t *testing.T) {
	conf.Init()

	// fb
	type betReport struct {
		UserId     int64 `json:"user_id" form:"user_id" gorm:"column:user_id;uniqueIndex:idx_game_user_type;type:bigint;"`
		GameType   int64 `json:"game_type" form:"game_type" gorm:"column:game_type;uniqueIndex:idx_game_user_type;type:int;"`
		Count      int64
		Bet        int64      `json:"bet" form:"bet" gorm:"column:bet;"`
		Wager      int64      `json:"wager" form:"wager" gorm:"column:wager;"`
		Win        int64      `json:"win" form:"win" gorm:"column:win;"`
		ProfitLoss int64      `json:"profit_loss" form:"profit_loss" gorm:"column:profit_loss;"`
		Status     int64      `json:"status" form:"status" gorm:"column:status;"`
		BetTime    *time.Time `json:"bet_time" form:"bet_time" gorm:"column:bet_time;type:timestamp;"`
		RewardTime *time.Time `json:"reward_time" form:"reward_time" gorm:"column:reward_time;type:timestamp;"`
	}
	start := time.Time{}
	end := time.Now()
	tableName := "taya_bet_report"

	r := []betReport{}
	err := model.DB.Select("user_id,  count(*) as count,sum(bet) as bet, sum(wager) as wager, sum(win) as win, sum(profit_loss) as profit_loss, max(reward_time) as reward_time").
		Table(tableName).Where("status = 5").Where("reward_time >= ?", start).Where("reward_time < ?", end).Group("user_id").Find(&r).Error
	// r = util.Filter(r, func(fbr models.FbBetReport) bool {
	// 	return fbr.Status == 5
	// })
	fmt.Println(err)
	p := util.MapSlice(r, func(input betReport) models.DailyProgressReport {
		return models.DailyProgressReport{
			Date:       start,
			UserID:     input.UserId,
			Name:       tableName,
			Bet:        input.Bet,
			Wager:      input.Wager,
			Win:        input.Win,
			ProfitLoss: input.ProfitLoss,
		}
	})
	fmt.Println(p)
}

func TestApi(t *testing.T) {
	ch := make(chan struct{})
	go Init(ch)
	<-ch
	OrderNoList := []string{}
	for i, orderNo := range OrderNoList {
		body := finpay.PaymentOrderCallBackReq{
			MerchantOrderNo:    orderNo,
			PaymentOrderNo:     fmt.Sprintf("Test.%s.%d", orderNo, i),
			PaymentOrderStatus: "COMPLETED",
			MerchantId:         "",
			MerchantAppId:      os.Getenv("FINPAY_MERCHANT_ID"),
			Amount:             111,
			Currency:           "PHP",
			// Sign                    :,
		}
		body.MockSign()
		b, _ := json.Marshal(body)
		buf := bytes.NewBuffer(b)

		req, err := http.NewRequest("POST", "http://localhost:3000/callback/finpay/payment-order", buf)
		if err != nil {
			// handle err
			panic(err)
		}
		req.Header.Set("Content-Type", "application/json")

		req.Header.Set("Signature", "super_signature_staging")
		req.Header.Set("Timestamp", "1683081388")
		req.Header.Set("Deviceinfo", `{"model":"Google","osVersion":"","platform":"android","uuid":"abcdef123456","version":"1.0.3"}`)
		req.Header.Set("Timezone", "Asia/Singapore")
		req.Header.Set("Authorization", "bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb3VudHJ5X2NvZGUiOiIiLCJleHAiOjE3MjQyMzQ5MTUsImlhdCI6MTY5MjY5ODkxNSwibW9iaWxlIjoiIiwidXNlcl9pZCI6NTQ2fQ.kUkgHfUE2QmfKxYkiBfslIrmnTWMhdi_d6DhwZaKTRA")

		cl := http.Client{}
		_, err = cl.Do(req)
		if err != nil {
			panic(err)
		}
	}
}

func Init(ch chan struct{}) {
	conf.Init()

	r := server.NewRouter()
	pprof.Register(r)
	ch <- struct{}{}
	r.Run(":" + os.Getenv("PORT"))

	defer model.IPDB.Close()
}
