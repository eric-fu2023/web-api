package notification

import (
	"web-api/model"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

func MarkReadByUserAndSelectedNotifications(userId int64, userNotificationIds []int64) error {
	err := model.DB.Model(ploutos.UserNotification{}).Scopes(model.ByUserId(userId), model.ByIds(userNotificationIds)).Update(`is_read`, true).Error
	return err
}
