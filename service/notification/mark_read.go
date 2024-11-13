package notification

import (
	"context"
	"errors"
	"fmt"
	"log"

	"web-api/model"
	"web-api/serializer"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

func MarkReadByUserAndSelectedNotifications(userId int64, userNotificationIds []int64) error {
	err := model.DB.Model(ploutos.UserNotification{}).Scopes(model.ByUserId(userId), model.ByIds(userNotificationIds)).Update(`is_read`, true).Error
	return err
}

type ReadNotificationForm struct {
	Id           serializer.NotificationReferenceId `form:"reference_id" json:"reference_id"`
	CategoryType ploutos.NotificationCategoryType   `form:"category" json:"category"`
}

type UserNotificationMarkReadRequestV2_Options struct {
	All bool `form:"all" json:"all"`
}

type UserNotificationMarkReadRequestV2 struct {
	Notifications []ReadNotificationForm                    `form:"notifications"`
	Options       UserNotificationMarkReadRequestV2_Options `form:"options"`
}

// MarkNotificationsAsRead
func MarkNotificationsAsRead(ctx context.Context, user model.User, notifications []ReadNotificationForm) error {
	ctx = rfcontext.AppendCallDesc(ctx, "MarkNotificationsAsRead")
	var err error
	for _, notification := range notifications {
		_, _err := MarkNotificationAsRead(ctx, user, notification)
		log.Println(rfcontext.FmtJSON(rfcontext.AppendError(ctx, err, fmt.Sprintf("MarkNotificationAsRead: id %d", notification.Id))))
		err = errors.Join(err, _err)
	}
	return err
}

// MarkNotificationAsRead
func MarkNotificationAsRead(ctx context.Context, user model.User, notification ReadNotificationForm) (int64, error) {
	userId := user.ID

	_, notifId, err := notification.Id.IsNotificationId()
	if err != nil {
		log.Println(rfcontext.FmtJSON(rfcontext.AppendError(ctx, err, fmt.Sprintf("notification.Id.IsNotificationId"))))
		return 0, err
	}
	_, uNotifId, err := notification.Id.IsUserNotificationId()
	if err != nil {
		log.Println(rfcontext.FmtJSON(rfcontext.AppendError(ctx, err, fmt.Sprintf("notification.Id.IsUserNotificationId"))))
		return 0, err
	}

	marker := &UserNotificationMarker{
		UserId:             userId,
		UserNotificationId: uNotifId,
		NotificationId:     notifId,
	}

	userNotificationId, err := Mark(ctx, marker)
	ctx = rfcontext.AppendParams(ctx, "", map[string]interface{}{
		"user_id": userNotificationId,
	})

	if err != nil {
		log.Println(rfcontext.FmtJSON(rfcontext.AppendError(ctx, err, "Mark")))
		return 0, err
	}
	return userNotificationId, nil
}

// ReadMarker uses 2-step approach to mark user's notification as read [i.e Mark].
// if notification's broadcast option is set to all (not selective), the user's notification will not be created upfront.
// That is, the record insertion will defer til [ReadMarker.Mark].
type ReadMarker interface {
	getOrCreateUserNotification(ctx context.Context, tx *gorm.DB) (ploutos.UserNotification, error)
	markUserNotification(ctx context.Context, tx *gorm.DB, userNotificationId int64) error
}

var _ ReadMarker = &UserNotificationMarker{}

type UserNotificationMarker struct {
	UserId int64
	//CategoryType       ploutos.NotificationCategoryType
	NotificationId     int64
	UserNotificationId int64
}

// TypeHasNotification
// if the notification type appears in `notification` table
func TypeHasNotification(categoryType ploutos.NotificationCategoryType) (bool, error) {
	switch categoryType {
	case ploutos.NotificationCategoryTypeSystem:
		return false, nil
	case
		ploutos.NotificationCategoryTypePromotion,
		ploutos.NotificationCategoryTypeNotification,
		ploutos.NotificationCategoryTypeSportsBet,
		ploutos.NotificationCategoryTypeGame,
		ploutos.NotificationCategoryTypeStream:
		return true, nil
	default:
		return false, errors.New("MarkNotificationsAsRead: unknown or invalid notification category")
	}
}

func (n *UserNotificationMarker) getOrCreateUserNotification(ctx context.Context, tx *gorm.DB) (ploutos.UserNotification, error) {
	ctx = rfcontext.AppendCallDesc(ctx, "getOrCreateUserNotification")

	var userNotif ploutos.UserNotification
	err := tx.Debug().Model(ploutos.UserNotification{}).Where("id = ?", n.UserNotificationId).First(&userNotif).Error
	switch {
	case err == nil:
		return userNotif, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		ctx = rfcontext.AppendCallDesc(ctx, "handling errors.Is(err, gorm.ErrRecordNotFound)")
		//hasNotification, err := TypeHasNotification(n.CategoryType)
		//if err != nil {
		//	return ploutos.UserNotification{}, err
		//}
		//
		//if !hasNotification {
		//	return ploutos.UserNotification{}, errors.New("MarkNotificationsAsRead: notification not found in database")
		//}

		if n.NotificationId == 0 {
			return ploutos.UserNotification{}, errors.New("MarkNotificationsAsRead: notification not found in database. cannot create for user")
		}

		var notif ploutos.Notification
		err = tx.Debug().Model(ploutos.Notification{}).Where("id = ?", n.NotificationId).First(&notif).Error
		if err != nil {
			return ploutos.UserNotification{}, err
		}
		unotif := ploutos.UserNotification{
			UserId:         n.UserId,
			NotificationId: n.NotificationId,

			Text: notif.Content, // fixme
		}
		err = tx.Create(&unotif).Error
		if err != nil {
			return ploutos.UserNotification{}, err
		}
		return unotif, nil
	default:
		return ploutos.UserNotification{}, err
	}
}

func (n *UserNotificationMarker) markUserNotification(ctx context.Context, tx *gorm.DB, userNotificationId int64) error {
	// getOrCreateUserNotification
	err := tx.Model(&ploutos.UserNotification{}).
		Where("id = ?", userNotificationId).
		Updates(map[string]interface{}{
			"is_read": true,
		}).Error
	return err
}

// MarkNotificationsAsRead

// 2 sets to query and mark notifications of a user:
// 1. update existing `user_notifications`.read to true
// 2. do a reverse lookup for notifications not in user_notification, use a [ReadMarker]
func MarkAllNotificationsAsRead(ctx context.Context, user model.User) error {
	ctx = rfcontext.AppendCallDesc(ctx, "MarkAllNotificationsAsRead")
	userId := user.ID
	err := model.DB.Debug().Model(ploutos.UserNotification{}).Where("user_id = ?", userId).Updates(map[string]any{
		"is_read": true,
	}).Error

	if err != nil {
		log.Println(rfcontext.FmtJSON(rfcontext.AppendError(ctx, err, "\"is_read\": true,")))
		return err
	}

	var notificationIdsToAddToUser []int64
	err = model.DB.Debug().Raw("SELECT id FROM notifications WHERE NOT EXISTS (SELECT id FROM user_notifications WHERE user_notifications.user_id = ? AND user_notifications.notification_id = notifications.id);", user.ID).Scan(&notificationIdsToAddToUser).Error
	if err != nil {
		return err
	}

	ctx = rfcontext.AppendParams(ctx, "", map[string]interface{}{
		"notificationIdsToAddToUser": notificationIdsToAddToUser,
	})
	log.Println(rfcontext.FmtJSON(ctx))

	for _, notifId := range notificationIdsToAddToUser {
		marker := &UserNotificationMarker{
			UserId:             userId,
			NotificationId:     notifId,
			UserNotificationId: 0,
		}

		userNotificationId, err := Mark(ctx, marker)
		if err != nil {
			mCtx := rfcontext.AppendParams(ctx, "", map[string]interface{}{
				"user_id":  userNotificationId,
				"notif_id": notifId,
			})
			log.Println(rfcontext.FmtJSON(rfcontext.AppendError(mCtx, err, "Mark")))
		}
	}
	return nil
}

func Mark(ctx context.Context, marker ReadMarker) (int64, error) {
	var userNotificationId int64

	err := model.DB.Transaction(func(tx *gorm.DB) error {
		userNotif, err := marker.getOrCreateUserNotification(ctx, tx)
		if err != nil {
			return err
		}
		userNotificationId = userNotif.ID
		err = marker.markUserNotification(ctx, tx, userNotificationId)
		return err
	})
	if err != nil {
		return 0, err
	}
	return userNotificationId, nil
}

func AddReadNotificationsV2(ctx context.Context, user model.User, req UserNotificationMarkReadRequestV2) error {
	ctx = rfcontext.AppendCallDesc(rfcontext.Spawn(context.Background()), "domain.AddReadNotificationsV2")
	if len(req.Notifications) > 0 && req.Options.All {
		err := fmt.Errorf("either select some or all")
		log.Println(rfcontext.FmtJSON(rfcontext.AppendError(ctx, err, "domain.MarkNotificationsAsRead")))
		return err
	}

	switch {
	case req.Options.All:
		err := MarkAllNotificationsAsRead(ctx, user)
		if err != nil {
			log.Println(rfcontext.FmtJSON(rfcontext.AppendError(ctx, err, "domain.MarkAllNotificationsAsRead")))
			return err
		}
	default:
		err := MarkNotificationsAsRead(ctx, user, req.Notifications)
		if err != nil {
			log.Println(rfcontext.FmtJSON(rfcontext.AppendError(ctx, err, "domain.MarkNotificationsAsRead")))
			return err
		}
	}
	return nil
}
