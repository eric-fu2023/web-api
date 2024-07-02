package service

import (
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

var orderType = map[int64][]int64{
	1: {ploutos.GAME_FB, ploutos.GAME_SABA, ploutos.GAME_TAYA, ploutos.GAME_IMSB},
	2: {ploutos.GAME_HACKSAW, ploutos.GAME_DOLLAR_JACKPOT, ploutos.GAME_STREAM_GAME},
}

type OrderListService struct {
	Type      int64  `form:"type" json:"type" binding:"required"`
	IsParlay  bool   `form:"is_parlay" json:"is_parlay"`
	IsSettled *bool  `form:"is_settled" json:"is_settled"`
	Start     string `form:"start" json:"start" binding:"required"`
	End       string `form:"end" json:"end" binding:"required"`
	common.Page
}

type OrderSummary struct {
	Count  int64 `gorm:"column:count"`
	Amount int64 `gorm:"column:amount"`
	Win    int64 `gorm:"column:win"`
}

func (service *OrderListService) List(c *gin.Context) serializer.Response {
	var err error
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var list []ploutos.BetReport
	var orderSummary OrderSummary
	var start, end time.Time
	loc := c.MustGet("_tz").(*time.Location)
	if service.Start != "" {
		if v, e := time.ParseInLocation(time.DateOnly, service.Start, loc); e == nil {
			start = v.UTC()
		}
	}
	if service.End != "" {
		if v, e := time.ParseInLocation(time.DateOnly, service.End, loc); e == nil {
			end = v.UTC().Add(24*time.Hour - 1*time.Second)
		}
	}

	statuses := []int64{2, 5}
	sumStatuses := statuses
	if service.IsSettled != nil && *service.IsSettled {
		sumStatuses = []int64{5, 6}
	}
	types := orderType[service.Type]
	if service.Type == 2 { // 2: games (which include games from game integration)
		var gi []ploutos.GameIntegration
		err = model.DB.Model(ploutos.GameIntegration{}).Preload(`GameVendors`).Find(&gi).Error
		if err != nil {
			return serializer.DBErr(c, service, i18n.T("general_error"), err)
		}
		for _, g := range gi {
			for _, v := range g.GameVendors {
				types = append(types, v.ID)
			}
		}
	}
	err = model.DB.Model(ploutos.BetReport{}).Scopes(model.ByOrderListConditions(user.ID, types, sumStatuses, service.IsParlay, service.IsSettled, start, end)).
		Select(`COUNT(1) as count, SUM(bet) as amount, SUM(win-bet) as win`).Find(&orderSummary).Error
	if err != nil {
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}
	err = model.DB.Preload(`Voucher`).Preload(`ImVoucher`).Preload(`GameVendor`).
		Model(ploutos.BetReport{}).Scopes(model.ByOrderListConditions(user.ID, types, statuses, service.IsParlay, service.IsSettled, start, end), model.ByBetTimeSort, model.Paginate(service.Page.Page, service.Page.Limit)).
		Find(&list).Error
	if err != nil {
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	for i, l := range list {
		l.ParseInfo()
		list[i] = l
	}

	go updateOrderLastSeen(user.ID)

	return serializer.Response{
		Data: serializer.BuildPaginatedBetReport(c, list, orderSummary.Count, orderSummary.Amount, orderSummary.Win),
	}
}

func updateOrderLastSeen(userId int64) {
	model.DB.Model(ploutos.UserCounter{}).Scopes(model.ByUserId(userId)).Updates(map[string]interface{}{"order_count": 0, "order_last_seen": time.Now()})
}
