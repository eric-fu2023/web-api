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
	CategoryType ploutos.NotificationCategoryType   `form:"category_type" json:"category_type"`
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

	var marker ReadMarker
	marker = &UserNotificationMarker{
		UserId:             userId,
		UserNotificationId: uNotifId,
		NotificationId:     notifId,
		//CategoryType:       notification.CategoryType,
	}

	var userNotificationId int64
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		userNotif, err := marker.getOrCreateUserNotification(ctx, tx)
		if err != nil {
			return err
		}
		userNotificationId = userNotif.ID
		err = marker.markUserNotification(ctx, tx, userNotificationId)
		return err
	})

	ctx = rfcontext.AppendParams(ctx, "", map[string]interface{}{
		"user_id": userNotificationId,
	})

	if err != nil {
		log.Println(rfcontext.FmtJSON(rfcontext.AppendError(ctx, err, "db.Transaction")))
		return userNotificationId, err
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
