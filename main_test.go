package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"web-api/conf"
	"web-api/model"
	"web-api/server"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	"github.com/gin-contrib/pprof"
)

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
			PaymentType:        os.Getenv("FINPAY_MERCHANT_APP_ID"),
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
