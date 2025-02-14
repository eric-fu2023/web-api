package service

import (
	"strconv"
	"strings"
	"time"
	"web-api/service/notification"

	"web-api/model"
	"web-api/serializer"
	"web-api/service/backend_for_frontend/game_history_pane"
	"web-api/service/common"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type UserNotificationListService struct {
	common.Page
}

func (service *UserNotificationListService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var notifications []ploutos.UserNotification
	err = model.DB.Scopes(model.ByUserId(user.ID), model.Paginate(service.Page.Page, service.Page.Limit), model.SortByCreated).Find(&notifications).Error
	if err != nil {
		r = serializer.DBErr(c, service, i18n.T("general_error"), err)
		return
	}
	var list []serializer.UserNotification
	for _, notification := range notifications {
		list = append(list, serializer.BuildUserNotification(c, notification))
	}

	go game_history_pane.AdvanceNotificationLastSeen(user.ID, time.Now())

	r = serializer.Response{
		Data: list,
	}
	return
}

type UserNotificationMarkReadService struct {
	Ids string `form:"ids" json:"ids"`
}

func (service *UserNotificationMarkReadService) MarkRead(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	var ids []int64
	for _, s := range strings.Split(service.Ids, ",") {
		if i, e := strconv.Atoi(strings.TrimSpace(s)); e == nil {
			ids = append(ids, int64(i))
		}
	}

	err = notification.MarkReadByUserAndSelectedNotifications(user.ID, ids)
	if err != nil {
		r = serializer.DBErr(c, service, i18n.T("general_error"), err)
		return
	}
	r.Msg = "Success"
	return
}
