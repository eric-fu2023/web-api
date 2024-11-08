package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"blgit.rfdev.tech/taya/common-function/rfcontext"

	"web-api/model"
	"web-api/serializer"
	"web-api/service/backend_for_frontend/game_history_pane"
	"web-api/service/common"
	notificationservice "web-api/service/notification"
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

	err = _MarkReadByUserAndSelectedNotifications(user.ID, ids)
	if err != nil {
		r = serializer.DBErr(c, service, i18n.T("general_error"), err)
		return
	}
	r.Msg = "Success"
	return
}

func _MarkReadByUserAndSelectedNotifications(userId int64, userNotificationIds []int64) error {
	err := model.DB.Model(ploutos.UserNotification{}).Scopes(model.ByUserId(userId), model.ByIds(userNotificationIds)).Update(`is_read`, true).Error
	return err
}

type GetUserNotificationRequest struct {
	UserNotificationId int64 `form:"user_notification_id" json:"user_notification_id"`
	NotificationId     int64 `form:"notification_id" json:"notification_id"`
	CategoryType       int64 `form:"category_type" json:"category_type"`
}

func GetUserNotification(c *gin.Context, req GetUserNotificationRequest) (serializer.Response, error) {
	u, _ := c.Get("user")
	user := u.(model.User)

	ctx := rfcontext.AppendCallDesc(rfcontext.Spawn(context.Background()), "GetUserNotification")
	notif, err := notificationservice.FindOne(ctx, user, ploutos.NotificationCategoryType(req.CategoryType), req.NotificationId, req.UserNotificationId)
	if err != nil {
		return serializer.Response{}, err
	}

	return serializer.Response{
		Code:  0,
		Data:  notif,
		Msg:   "",
		Error: "",
	}, err
}
