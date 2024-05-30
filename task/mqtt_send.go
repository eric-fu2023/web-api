package task

import (
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"net/http"
	"net/http/httptest"
	"web-api/api"
)

func SendDummyMqtt() {
	c := cron.New(cron.WithSeconds())
	c.AddFunc("0 * * * * *", func() {
		send()
	})
	c.Start()
}

func send() {
	result := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(result)
	ctx.Request, _ = http.NewRequest("GET", "/finpay_redirect?user_id=0", nil)
	api.FinpayRedirect(ctx)
}
