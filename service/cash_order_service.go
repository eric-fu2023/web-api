package service

import (
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type CashOrderService struct {
	TopupOnly    bool   `form:"topup_only" json:"topup_only"`
	WithdrawOnly bool   `form:"withdraw_only" json:"withdraw_only"`
	StartTime    string `form:"start_time" json:"start_time"`
	EndTime      string `form:"end_time" json:"end_time"`
	common.Page
}

func (s CashOrderService) List(c *gin.Context) (r serializer.Response, err error) {
	u, _ := c.Get("user")
	user := u.(model.User)
	loc := c.MustGet("_tz").(*time.Location)
	i18n := c.MustGet("i18n").(i18n.I18n)

	var startTime *time.Time
	if val, err := time.ParseInLocation(consts.StdTimeFormat, s.StartTime, loc); err == nil {
		startTime = &val
	}
	var endTime *time.Time
	if val, err := time.ParseInLocation(consts.StdTimeFormat, s.EndTime, loc); err == nil {
		endTime = &val
	}

	list, err := model.CashOrder{}.List(user.ID, s.TopupOnly, s.WithdrawOnly, startTime, endTime, s.Page.Page, s.Limit)
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}
	r.Data = util.MapSlice(list, func(input model.CashOrder) serializer.GenericCashOrder {
		return serializer.BuildGenericCashOrder(input, i18n)
	})
	go updateTransactionLastSeen(user.ID)
	return
}

func updateTransactionLastSeen(userId int64) {
	model.DB.Model(ploutos.UserCounter{}).Scopes(model.ByUserId(userId)).Update(`transaction_last_seen`, time.Now())
}
