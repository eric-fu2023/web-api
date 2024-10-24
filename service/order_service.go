package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"web-api/model"
	"web-api/serializer"
	"web-api/service/backend_for_frontend/game_history_pane"
	"web-api/service/common"
	"web-api/util/i18n"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
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
	rfCtx = rfcontext.AppendDescription(rfCtx, fmt.Sprintf("service.IsSettled %#v", service.IsSettled))
	rfCtx = rfcontext.AppendDescription(rfCtx, fmt.Sprintf("service.PaneType %#v", service.PaneType))
	log.Println(rfcontext.Fmt(rfCtx))

	gameVendorIds, err := game_history_pane.GetGameVendorIdsByPaneType(service.PaneType)
	if err != nil {
		return serializer.Err(c, service, 500, i18n.T("general_error"), err)
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

	statuses := model.IsSettledFlagToPloutosIncludeStatuses(service.IsSettled, false)
	rfCtx = rfcontext.AppendDescription(rfCtx, fmt.Sprintf("statuses %#v", statuses))
	log.Println(rfcontext.Fmt(rfCtx))
	betReports, err := model.BetReports(rfCtx, user.ID, start, end, gameVendorIds, statuses, service.IsParlay, service.Page.Page, service.Page.Limit)
	if err != nil {
		rfCtx = rfcontext.AppendError(rfCtx, err, ".BetReports")
		log.Println(rfcontext.Fmt(rfCtx))
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	for i, l := range betReports {
		l.ParseInfo()
		betReports[i] = l
	}

	sumStatuses := model.IsSettledFlagToPloutosIncludeStatuses(service.IsSettled, true)
	rfCtx = rfcontext.AppendDescription(rfCtx, fmt.Sprintf("sumStatuses %#v", sumStatuses))
	log.Println(rfcontext.Fmt(rfCtx))

	orderSummary, err := model.BetReportsStats(rfCtx, user.ID, start, end, gameVendorIds, sumStatuses, service.IsParlay)
	if err != nil {
		rfCtx = rfcontext.AppendError(rfCtx, err, ".BetReportsStats")
		log.Println(rfcontext.Fmt(rfCtx))
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	go func() {
		ocerr := game_history_pane.ResetUserCounter_OrderCount(user.ID)
		if ocerr == nil { // debug
			rfCtx = rfcontext.AppendDescription(rfCtx, "OrderHistory_GamePane_LastSeen ok")
			log.Println(rfcontext.Fmt(rfCtx))
		}
	}()
	go func() {
		oherr := game_history_pane.AdvanceUserCounter_OrderHistory_GamePane_LastSeen(user.ID, service.PaneType, time.Now())
		if oherr == nil { // debug
			rfCtx = rfcontext.AppendDescription(rfCtx, "OrderHistory_GamePane_LastSeen ok")
			log.Println(rfcontext.Fmt(rfCtx))
		}
	}()

	return serializer.Response{
		Data: serializer.BuildPaginatedBetReport(c, betReports, orderSummary.Count, orderSummary.Amount, orderSummary.Win),
	}
}
