package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
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

	counters := model.UserCounters{
		Order:        counter.OrderCount,
		Transaction:  txCount,
		Notification: notificationCount,
	}
	return serializer.Response{
		Data: serializer.BuildUserCounters(c, counters),
	}
}

func (service *CounterService) countTransactions(userId int64, time time.Time) (count int64, err error) {
	err = model.DB.Model(ploutos.CashOrder{}).Scopes(model.ByUserId(userId), model.ByCreatedAtGreaterThan(time)).Count(&count).Error
	return
}

func (service *CounterService) countNotifications(userId int64, time time.Time) (count int64, err error) {
	err = model.DB.Model(ploutos.UserNotification{}).Scopes(model.ByUserId(userId), model.ByCreatedAtGreaterThan(time)).Count(&count).Error
	return
}
