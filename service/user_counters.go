package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"web-api/model"
	"web-api/serializer"
	"web-api/service/backend_for_frontend/game_history_pane"
	"web-api/util/i18n"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type _ = ploutos.UserCounter
type UserCounter struct {
	ploutos.BASE
	UserId                    int64
	OrderCount                int64
	OrderLastSeen             time.Time
	TransactionLastSeen       time.Time
	NotificationLastSeen      time.Time
	OnlineDuration            int64
	OnlineDurationLastLogId   sql.NullInt64
	GameHistorySportsLastSeen time.Time
	GameHistoryCasinoLastSeen time.Time
}

func (u *UserCounter) TableName() string {
	return ploutos.TableUserCounter
}

// LastSeenForGamePane to identify a particular db column and read its value. prioritise safeness by hardcoding instead of using reflection on field tags.
func (userCounter *UserCounter) LastSeenForGamePane(paneType game_history_pane.GamesHistoryPaneType) (time.Time, error) {
	switch paneType {
	case game_history_pane.GamesPaneAll:
		return userCounter.OrderLastSeen, nil
	case game_history_pane.GamesPaneTypeSports:
		return userCounter.GameHistorySportsLastSeen, nil
	case game_history_pane.GamesPaneTypeCasino:
		return userCounter.GameHistoryCasinoLastSeen, nil
	}
	return time.Time{}, fmt.Errorf("unknown last seen for game pane type")
}

type CounterService struct {
}

func (service *CounterService) Get(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	rfCtx := rfcontext.Spawn(context.Background())
	rfCtx = rfcontext.AppendCallDesc(rfCtx, "CounterService) Get")
	var counter UserCounter
	err := model.DB.Model(UserCounter{}).Scopes(model.ByUserId(user.ID)).Find(&counter).Error
	if err != nil {
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	txCount, err := service.countTransactions(user.ID, counter.TransactionLastSeen)
	if err != nil {
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	notificationCount, err := service.countNotifications(user.ID, counter.NotificationLastSeen)
	if err != nil {
		return serializer.DBErr(c, service, i18n.T("general_error"), err)
	}

	counters := model.UserCounters{
		Order:        counter.OrderCount,
		Transaction:  txCount,
		Notification: notificationCount,
	}

	now := time.Now()
	statuses := model.IsSettledFlagToPloutosIncludeStatuses(nil, false /* count for all reports yet to be seen*/)

	rfCtx = rfcontext.AppendCallDesc(rfCtx, fmt.Sprintf("IsSettledFlagToPloutosIncludeStatuses = %v", statuses))
	rfCtx = rfcontext.AppendCallDesc(rfCtx, fmt.Sprintf("game_history_pane.GamePaneHistoryTypes()  = %v", game_history_pane.GamePaneHistoryTypes()))
	log.Printf(rfcontext.Fmt(rfCtx))

	// gameHistoryPaneCounts
	// for type 0 (all), add all counts from others
	// for others types, count(reports) since last seen.
	gameHistoryPaneCounts := make(map[game_history_pane.GamesHistoryPaneType]int64)
	gameHistoryPaneCounts[game_history_pane.GamesPaneAll] = 0
	for _, gamePane := range game_history_pane.GamePaneHistoryTypes() {
		if gamePane == game_history_pane.GamesPaneAll {
			continue
		}
		gameHistoryPaneCounts[gamePane] = 0

		pCtx := rfcontext.AppendCallDesc(rfCtx, "counting for game history type: "+strconv.Itoa(int(gamePane)))
		lastSeen, err := counter.LastSeenForGamePane(gamePane)
		if err != nil {
			pCtx = rfcontext.AppendErrorAsWarn(pCtx, fmt.Errorf("%v %v", err, gamePane), "getting column name for game pane")
			log.Printf(rfcontext.Fmt(pCtx))
			continue
		}

		gameVendorIds, err := game_history_pane.GetGameVendorIdsByPaneType(gamePane)
		orderSummary, derr := model.BetReportsStats(rfCtx, user.ID, lastSeen, now, gameVendorIds, statuses, false)
		if derr != nil {
			pCtx = rfcontext.AppendErrorAsWarn(pCtx, fmt.Errorf("%v", gamePane), "getting game vendor id for game pane")
			log.Printf(rfcontext.Fmt(pCtx))
		}

		gameHistoryPaneCounts[gamePane] = orderSummary.Count
		gameHistoryPaneCounts[game_history_pane.GamesPaneAll] += orderSummary.Count
	}

	for gamePane, count := range gameHistoryPaneCounts {
		if gamePane != game_history_pane.GamesPaneAll {
			gameHistoryPaneCounts[game_history_pane.GamesPaneAll] += count
		}
	}

	data := serializer.BuildUserCounters(c, counters, gameHistoryPaneCounts)

	responseBody := serializer.Response{
		Data: data,
	}
	{ // debug
		rfCtx = rfcontext.AppendDescription(rfCtx, fmt.Sprintf("response body Struct: %#v", responseBody))
		jj, _ := json.Marshal(responseBody)
		rfCtx = rfcontext.AppendDescription(rfCtx, fmt.Sprintf("response body JSON: %s", string(jj)))
		log.Println(rfcontext.Fmt(rfCtx))
	}
	return responseBody
}

func (service *CounterService) countTransactions(userId int64, fromCreatedTime time.Time) (count int64, err error) {
	err = model.DB.Model(ploutos.CashOrder{}).Scopes(model.ByUserId(userId), model.ByCreatedAtGreaterThan(fromCreatedTime)).Count(&count).Error
	return
}

func (service *CounterService) countNotifications(userId int64, fromCreatedTime time.Time) (count int64, err error) {
	err = model.DB.Model(ploutos.UserNotification{}).Scopes(model.ByUserId(userId), model.ByCreatedAtGreaterThan(fromCreatedTime)).Count(&count).Error
	return
}
