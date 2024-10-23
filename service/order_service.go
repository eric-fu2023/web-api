package service

import (
	"context"
	"log"
	"time"

	"web-api/model"
	"web-api/serializer"
	"web-api/service/backend_for_frontend/game_vendor_pane"
	"web-api/service/common"
	"web-api/util/i18n"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type CheckOrderService struct {
	BusinessId string `form:"business_id" json:"business_id" binding:"required"`
}

type OrderListService struct {
	PaneType  int64  `form:"type" json:"type" binding:"required"`
	IsParlay  bool   `form:"is_parlay" json:"is_parlay"`
	IsSettled *bool  `form:"is_settled" json:"is_settled"`
	Start     string `form:"start" json:"start" binding:"required"`
	End       string `form:"end" json:"end" binding:"required"`
	common.Page
}

func (service *CheckOrderService) CheckOrder(c *gin.Context) serializer.Response {
	// Check if Business Id already generated in Bet Report
	isFoundBetReport, _ := model.GetBetReportByBusinessId(service.BusinessId)
	return serializer.Response{
		Data: isFoundBetReport,
	}

}

func (service *OrderListService) List(c *gin.Context) serializer.Response {
	var err error
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

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

	rfCtx := rfcontext.Spawn(context.Background())
	rfCtx = rfcontext.AppendCallDesc(rfCtx, "OrderListService")

	gameVendorIds := game_vendor_pane.GameVendorIdsByPaneType[service.PaneType]
	if service.PaneType == game_vendor_pane.GamesPaneTypeCasino { // assumes all game vendors of game integrations are of the same type GamesPaneTypeCasino and include them in gameVendorIds
		var gi []ploutos.GameIntegration
		err = model.DB.Model(ploutos.GameIntegration{}).Preload(`GameVendors`).Find(&gi).Error
		if err != nil {
			return serializer.DBErr(c, service, i18n.T("general_error"), err)
		}
		for _, g := range gi {
			for _, v := range g.GameVendors {
				gameVendorIds = append(gameVendorIds, v.ID)
			}
		}
	}

	// status mapping
	// 0:  "Created",
	// 1:  "Confirming",
	// 2:  "Rejected",
	// 3:  "Canceled",
	// 4:  "Confirmed",
	// 5:  "Settled",
	// 6:  "EarlySettled",

	// IsSettled = nil
	// take all status

	// IsSettled = true
	// take status: [2, 3, 5, 6]
	// take sum_status: [5, 6]

	// IsSettled = false
	// take status: [0, 1, 4]
	// take sum_status: [0, 1, 4]

	statuses := []int64{2, 3, 5, 6}

	var sumStatuses []int64
	if service.IsSettled == nil { // "default" sumStatuses
		sumStatuses = statuses
	} else if /*service.IsSettled != nil &&*/ *service.IsSettled {
		sumStatuses = []int64{5, 6}
	} else if /*service.IsSettled != nil &&*/ !*service.IsSettled {
		sumStatuses = []int64{2, 3, 5, 6}
	}

	betReports, err := model.BetReports(rfCtx, user.ID, start, end, gameVendorIds, sumStatuses, service.IsParlay, service.IsSettled, service.Page.Page, service.Page.Limit)
	if err != nil {
		rfCtx = rfcontext.AppendError(rfCtx, err, ".BetReports")
		log.Println(rfcontext.Fmt(rfCtx))
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	for i, l := range betReports {
		l.ParseInfo()
		betReports[i] = l
	}

	orderSummary, err := model.BetReportsStats(rfCtx, user.ID, start, end, gameVendorIds, sumStatuses, service.IsParlay, service.IsSettled)
	if err != nil {
		rfCtx = rfcontext.AppendError(rfCtx, err, ".BetReportsStats")
		log.Println(rfcontext.Fmt(rfCtx))
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	go ResetUserCounter_OrderCount(user.ID)
	go game_vendor_pane.ResetUserCounter_Order_GamePane(user.ID, service.PaneType)

	return serializer.Response{
		Data: serializer.BuildPaginatedBetReport(c, betReports, orderSummary.Count, orderSummary.Amount, orderSummary.Win),
	}
}

func ResetUserCounter_OrderCount(userId int64) {
	model.DB.Model(ploutos.UserCounter{}).Scopes(model.ByUserId(userId)).Updates(map[string]interface{}{"order_count": 0, "order_last_seen": time.Now()})
}
