package service

import (
	"context"
	"log"
	"time"

	"web-api/model"
	"web-api/serializer"
	"web-api/service/backend_for_frontend/game_vendor_pane"
	"web-api/util/i18n"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type CounterService struct {
}

func (service *CounterService) Get(c *gin.Context) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var counter ploutos.UserCounter
	err := model.DB.Model(ploutos.UserCounter{}).Scopes(model.ByUserId(user.ID)).Find(&counter).Error
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
	ctx := rfcontext.Spawn(context.Background())

	_, derr := game_vendor_pane.CountBetReports(ctx, user.ID, time.Time{}, []int64{})
	if derr != nil {
		ctx = rfcontext.AppendError(ctx, derr, "get db")
		log.Println(rfcontext.Fmt(ctx))
	}

	counters := model.UserCounters{
		Order:        counter.OrderCount,
		Transaction:  txCount,
		Notification: notificationCount,
	}

	return serializer.Response{
		Data: serializer.BuildUserCounters(c, counters),
	}
}

func (service *CounterService) countTransactions(userId int64, fromCreatedTime time.Time) (count int64, err error) {
	err = model.DB.Model(ploutos.CashOrder{}).Scopes(model.ByUserId(userId), model.ByCreatedAtGreaterThan(fromCreatedTime)).Count(&count).Error
	return
}

func (service *CounterService) countNotifications(userId int64, fromCreatedTime time.Time) (count int64, err error) {
	err = model.DB.Model(ploutos.UserNotification{}).Scopes(model.ByUserId(userId), model.ByCreatedAtGreaterThan(fromCreatedTime)).Count(&count).Error
	return
}
