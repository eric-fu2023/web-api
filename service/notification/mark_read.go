package notification

import (
	"context"
	"errors"
	"log"

	"web-api/model"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

func MarkReadByUserAndSelectedNotifications(userId int64, userNotificationIds []int64) error {
	err := model.DB.Model(ploutos.UserNotification{}).Scopes(model.ByUserId(userId), model.ByIds(userNotificationIds)).Update(`is_read`, true).Error
	return err
}

type UserNotificationMarkReadForm struct {
	UserNotificationId int64                            `form:"user_notification_id" json:"user_notification_id"`
	NotificationId     int64                            `form:"notification_id" json:"notification_id"`
	CategoryType       ploutos.NotificationCategoryType `form:"category_type" json:"category_type"`
}

// MarkNotificationsAsRead
func MarkNotificationsAsRead(ctx context.Context, user model.User, notifications []UserNotificationMarkReadForm) error {
	var err error
	//userId := user.ID
	//for _, notification := range notifications {
	//	category := notification.CategoryType
	//
	//	var marker ReadMarker
	//	switch category {
	//	case ploutos.NotificationCategoryTypeSystem:
	//		marker := &NotificationCategoryTypeSystemMarker{UserId: userId}
	//	//case ploutos.NotificationCategoryTypeNotification:
	//
	//	default:
	//		return errors.New("MarkNotificationsAsRead: unknown or invalid notification category")
	//	}
	//
	//	marker.Mark()
	//	err = errors.Join()
	//}
	return err
}

// MarkNotificationAsRead
func MarkNotificationAsRead(ctx context.Context, user model.User, notification UserNotificationMarkReadForm) (any, error) {
	userId := user.ID

	var marker ReadMarker
	switch notification.CategoryType {
	case ploutos.NotificationCategoryTypeSystem:
		marker = &CategoryTypeSystemMarker{
			UserId: userId,
		}
	case
		ploutos.NotificationCategoryTypePromotion,
		ploutos.NotificationCategoryTypeNotification,
		ploutos.NotificationCategoryTypeSportsBet,
		ploutos.NotificationCategoryTypeGame,
		ploutos.NotificationCategoryTypeStream:

	default:
		return nil, errors.New("MarkNotificationsAsRead: unknown or invalid notification category")
	}

	return marker.Mark(ctx)
}

// ReadMarker uses 2-step approach to mark user's notification as read [i.e Mark].
// if notification's broadcast option is set to all (not selective), the user's notification will not be created upfront.
// That is, the record insertion will defer til [ReadMarker.Mark].

type ReadMarker interface {
	getOrCreateUserNotification(ctx context.Context, tx *gorm.DB) (ploutos.UserNotification, error)
	markUserNotification(ctx context.Context, tx *gorm.DB, userNotificationId int64) (ploutos.UserNotification, error)

	// Mark
	// 1. getOrCreateUserNotification to get the record which stores the read flag.
	// 2. markUserNotification to set the read flag.
	Mark(context.Context) (userNotificationId int64, _ error)
}

var _ ReadMarker = &CategoryTypeSystemMarker{}

type CategoryTypeSystemMarker struct {
	UserId int64

	db *gorm.DB
}

func (n *CategoryTypeSystemMarker) getOrCreateUserNotification(ctx context.Context, tx *gorm.DB) (ploutos.UserNotification, error) {
	return ploutos.UserNotification{}, errors.New("implement me")
}

func (n *CategoryTypeSystemMarker) markUserNotification(ctx context.Context, tx *gorm.DB, userNotificationId int64) (ploutos.UserNotification, error) {
	return ploutos.UserNotification{}, errors.New("implement me")
}

func (n *CategoryTypeSystemMarker) Mark(ctx context.Context) (int64, error) {
	ctx = rfcontext.AppendCallDesc(ctx, " (n *CategoryTypeSystemMarker) Mark")

	var userNotificationId int64
	err := n.db.Transaction(func(tx *gorm.DB) error {
		userNotif, err := n.getOrCreateUserNotification(ctx, tx)
		if err != nil {
			return err
		}
		userNotificationId = userNotif.ID
		_, err = n.markUserNotification(ctx, tx, userNotificationId)
		return err
	})

	ctx = rfcontext.AppendParams(ctx, "", map[string]interface{}{
		"user_id": userNotificationId,
	})

	if err != nil {
		log.Println(rfcontext.FmtJSON(rfcontext.AppendError(ctx, err, "db.Transaction")))
		return 0, err
	}
	return 0, errors.New("implement me")
}
