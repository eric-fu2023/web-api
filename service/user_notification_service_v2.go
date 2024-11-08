package service

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"blgit.rfdev.tech/taya/common-function/rfcontext"

	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/backend_for_frontend/game_history_pane"
	"web-api/service/common"
	notificationservice "web-api/service/notification"
	"web-api/util/i18n"

	models "blgit.rfdev.tech/taya/ploutos-object"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type UserNotificationListServiceV2 struct {
	common.Page
	Category int `form:"category" json:"category"`
}

func (service *UserNotificationListServiceV2) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)
	var list []serializer.UserNotificationV2

	// system notification
	var notifications []ploutos.UserNotification
	err = model.DB.Scopes(model.ByUserId(user.ID), model.Paginate(service.Page.Page, service.Page.Limit), model.SortByCreated).Find(&notifications).Error
	if err != nil {
		r = serializer.DBErr(c, service, i18n.T("general_error"), err)
		return
	}

	if service.Category == consts.NotificationCategorySystem || service.Category == 0 {
		for _, notification := range notifications {
			list = append(list, serializer.BuildUserNotificationV2(c, notification))
		}
	}
	// category 999 is system msg, so we dun need to query cms notification
	if service.Category != consts.NotificationCategorySystem {
		// cms notifications,
		var notification_ids_with_read_status []serializer.NotificationIdsWithReadStatus
		var notification_ids []int64
		for _, notification := range notifications {
			notification_ids_with_read_status = append(notification_ids_with_read_status, serializer.NotificationIdsWithReadStatus{
				ID:     notification.NotificationId,
				IsRead: notification.IsRead,
			})
			notification_ids = append(notification_ids, notification.NotificationId)
		}
		var cms_notifications []ploutos.Notification
		// filter by category is category is not 0
		if service.Category != 0 {
			err = model.DB.Scopes(model.Paginate(service.Page.Page, service.Page.Limit), model.SortByCreated).Where("id in ? or target = 0", notification_ids).Where("category = ?", service.Category).Find(&cms_notifications).Error
		} else {
			err = model.DB.Scopes(model.Paginate(service.Page.Page, service.Page.Limit), model.SortByCreated).Where("id in ? or target = 0", notification_ids).Find(&cms_notifications).Error
		}
		if err != nil {
			r = serializer.DBErr(c, service, i18n.T("general_error"), err)
			return
		}
		for _, cms_notification := range cms_notifications {
			image_url := cms_notification.ImageUrl
			if cms_notification.Category == consts.NotificationCategoryGame {
				var game_image_url string
				err = model.DB.Select("app_icon").Table("sub_game_brand").Where("id = ?", cms_notification.CategoryContentID).Scan(&game_image_url).Error
				if err != nil {
					r = serializer.DBErr(c, service, i18n.T("general_error"), err)
					return
				}
				image_url = game_image_url
			}
			if cms_notification.Category == consts.NotificationCategoryPromotion {
				promotionImages := serializer.IncomingPromotionImages{}
				var promotion models.Promotion
				err = model.DB.Table("promotions").Where("id = ?", cms_notification.CategoryContentID).Scan(&promotion).Error
				if err != nil {
					r = serializer.DBErr(c, service, i18n.T("general_error"), err)
					return
				}
				// Unmarshal the JSONB data into the `PromotionImageUrl` struct
				err = json.Unmarshal([]byte(promotion.Image), &promotionImages)
				if err != nil {
					r = serializer.DBErr(c, service, i18n.T("general_error"), err)
					return
				}
				image_url = promotionImages.H5
			}
			if cms_notification.Category == consts.NotificationCategoryStream {
				var stream_image_url string
				err = model.DB.Select("img_url").Table("live_streams").Where("id = ?", cms_notification.CategoryContentID).Scan(&stream_image_url).Error
				if err != nil {
					r = serializer.DBErr(c, service, i18n.T("general_error"), err)
					return
				}
				image_url = stream_image_url
			}
			list = append(list, serializer.BuildCMSNotificationV2(c, cms_notification, notification_ids_with_read_status, image_url))
		}
	}

	// sort list according to ts
	sort.Slice(list, func(i, j int) bool {
		return list[i].Ts > (list[j].Ts)
	})
	go game_history_pane.AdvanceNotificationLastSeen(user.ID, time.Now())

	r = serializer.Response{
		Data: list,
	}
	return
}

type UserNotificationMarkReadServiceV2 struct {
	Ids string `form:"ids" json:"ids"`
}

func (service *UserNotificationMarkReadServiceV2) MarkRead(c *gin.Context) (r serializer.Response, err error) {
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

func _MarkReadByUserAndSelectedNotificationsV2(userId int64, userNotificationIds []int64) error {
	err := model.DB.Model(ploutos.UserNotification{}).Scopes(model.ByUserId(userId), model.ByIds(userNotificationIds)).Update(`is_read`, true).Error
	return err
}

type GetUserNotificationRequestV2 struct {
	UserNotificationId int64 `form:"user_notification_id" json:"user_notification_id"`
	NotificationId     int64 `form:"notification_id" json:"notification_id"`
	CategoryType       int64 `form:"category_type" json:"category_type"`
}

func GetUserNotificationV2(c *gin.Context, req GetUserNotificationRequest) (serializer.Response, error) {
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
