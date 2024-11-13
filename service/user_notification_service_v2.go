package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/backend_for_frontend/game_history_pane"
	"web-api/service/common"
	notificationservice "web-api/service/notification"
	"web-api/util/i18n"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type UserNotificationListServiceV2 struct {
	common.Page
	Category int `form:"category" json:"category"`
}
type UserNotificationListCategoryServiceV2 struct {
}

func (service *UserNotificationListCategoryServiceV2) ListCategory(c *gin.Context) (r serializer.Response, err error) {
	u, _ := c.Get("user")
	user := u.(model.User)
    category := CountsUnread(user.ID)
	r = serializer.Response{
		Data: category,
	}
	return
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
			err = model.DB.Scopes(model.Paginate(service.Page.Page, service.Page.Limit), model.SortByCreated).Where("id in ? or target = 0", notification_ids).Where("send_at < ?", time.Now().UTC().Add(8*time.Hour)).Where("expired_at > ? or expired_at is null", time.Now().UTC().Add(8*time.Hour)).Where("category = ?", service.Category).Find(&cms_notifications).Error
		} else {
			err = model.DB.Scopes(model.Paginate(service.Page.Page, service.Page.Limit), model.SortByCreated).Where("id in ? or target = 0", notification_ids).Where("send_at < ?", time.Now().UTC().Add(8*time.Hour)).Where("expired_at > ? or expired_at is null", time.Now().UTC().Add(8*time.Hour)).Find(&cms_notifications).Error
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
				var promotion ploutos.Promotion
				err = model.DB.Table("promotions").Where("id = ?", cms_notification.CategoryContentID).Scan(&promotion).Error
				if err != nil {
					r = serializer.DBErr(c, service, i18n.T("general_error"), err)
					return
				}
				// Unmarshal the JSONB data into the `PromotionImageUrl` struct
				err = json.Unmarshal(promotion.Image, &promotionImages)
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

type UserNotificationMarkReadRequestV2 struct {
	Notifications []notificationservice.ReadNotificationForm `form:"notifications"`
}

func AddReadNotificationsV2(c *gin.Context, req UserNotificationMarkReadRequestV2) (serializer.Response, error) {
	u, _ := c.Get("user")
	user := u.(model.User)

	ctx := rfcontext.AppendCallDesc(rfcontext.Spawn(context.Background()), "AddReadNotificationsV2")
	err := notificationservice.MarkNotificationsAsRead(ctx, user, req.Notifications)
	if err != nil {
		log.Println(rfcontext.FmtJSON(rfcontext.AppendError(ctx, err, "MarkNotificationsAsRead")))
		return serializer.Response{}, err
	}

	return serializer.Response{
		Code:  0,
		Data:  nil,
		Msg:   "",
		Error: "",
	}, err
}

type GetGeneralNotificationRequestV2 struct {
	Id serializer.NotificationReferenceId `form:"reference_id" json:"reference_id"`
	//UserNotificationId int64 `form:"user_notification_id" json:"user_notification_id"`
	//NotificationId     int64 `form:"notification_id" json:"notification_id"`
	//CategoryType int64 `form:"category_type" json:"category_type"`
}

func GetGeneralNotificationV2(c *gin.Context, req GetGeneralNotificationRequestV2) (serializer.Response, error) {
	u, _ := c.Get("user")
	user := u.(model.User)

	ctx := rfcontext.AppendCallDesc(rfcontext.Spawn(context.Background()), "GetGeneralNotificationV2")
	isUNotif, uNotifId, err := req.Id.IsUserNotificationId()
	if err != nil {
		return serializer.Response{}, err
	}
	if !isUNotif {
		return serializer.Response{}, fmt.Errorf("reference id type is not user's notification")
	}

	notif, err := notificationservice.FindGeneralOne(ctx, user, uNotifId)
	if err != nil {
		log.Println(rfcontext.FmtJSON(rfcontext.AppendError(ctx, err, "GetGeneralNotificationV2")))
		return serializer.Response{}, err
	}

	return serializer.Response{
		Code:  0,
		Data:  notif,
		Msg:   "",
		Error: "",
	}, err
}

func CountsUnread(userID int64) []serializer.UserNotificationUnreadCountsV2 {
	var results []serializer.UserNotificationUnreadCountsV2
	// Define the raw SQL query with UNION ALL
	query := `
        -- System notifications
        SELECT 
            999 AS id,
            COUNT(*) AS unread_counts
        FROM user_notifications
        WHERE notification_id IS NULL 
          AND is_read = FALSE 
          AND deleted_at IS NULL 
          AND user_id = ?

        UNION ALL

        -- All notifications
        SELECT 
            n.category AS id,
            COUNT(*) AS unread_counts
        FROM notifications n
        LEFT JOIN user_notifications un ON un.notification_id = n.id AND un.user_id = ?
        WHERE un.id IS NULL
          AND (n.expired_at > CURRENT_TIMESTAMP OR n.expired_at IS NULL)
          AND n.send_at < CURRENT_TIMESTAMP
          AND n.target = 0
          AND n.deleted_at IS NULL 
        GROUP BY n.category

        UNION ALL

        -- Special notifications
        SELECT 
            n.category AS id,
            COUNT(*) AS unread_counts
        FROM user_notifications un
        JOIN notifications n ON un.notification_id = n.id
        WHERE un.notification_id IS NOT NULL
          AND un.is_read = FALSE 
          AND un.deleted_at IS NULL 
          AND n.deleted_at IS NULL 
          AND n.target = 1 
          AND un.user_id = ?
        GROUP BY n.category;
    `

	// Execute the raw SQL query with userID as a parameter
	if err := model.DB.Raw(query, userID, userID, userID).Scan(&results).Error; err != nil {
		log.Fatal("error executing query:", err)
	}

	results = append([]serializer.UserNotificationUnreadCountsV2{{
		ID:           0,
		Label:        "All",
		UnreadCounts: 0},
	}, results...)
	// counts unread system
	for index, item := range results {
		var label = ""
		switch item.ID {
		case 0:
			label = "All"
		case 1:
			label = "Promotion"
		case 2:
			label = "General"
		case 3:
			label = "Bet"
		case 4:
			label = "Game"
		case 5:
			label = "Live Stream"
		case 999:
			label = "System"
		}
		results[index].Label = label
		results[0].UnreadCounts = results[0].UnreadCounts + results[index].UnreadCounts
	}
	return results
}
