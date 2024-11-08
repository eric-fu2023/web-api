package notification

import (
	"context"
	"errors"
	"time"

	"web-api/model"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/lib/pq"
)

type _ = ploutos.Notification
type Notification struct {
	ID int64 `gorm:"primarykey" json:"id"` // 主键ID

	Title             string                           `gorm:"type:varchar(255)" json:"title"`
	Content           string                           `gorm:"type:text" json:"content"`
	Target            int32                            `gorm:"type:int" json:"target"`
	Vip               pq.Int32Array                    `gorm:"type:integer[]" json:"vip"`
	Category          ploutos.NotificationCategoryType `json:"category"`
	CategoryContentID int32                            `json:"category_content_id"`
	PushEnable        bool                             `json:"push_enable"`
	PushTitle         string                           `gorm:"type:varchar(255)" json:"push_title"`
	PushContent       string                           `gorm:"type:text" json:"push_content"`
	PushType          int32                            `json:"push_type"`
	PushTypeContentID int32                            `json:"push_type_content_id"`
	SendAt            time.Time                        `json:"send_at"`
	ExpiredAt         time.Time                        `json:"expired_at"`
	ImageUrl          string                           `gorm:"type:varchar(255)" json:"image_url"`
	ShortContent      string                           `gorm:"type:text" json:"short_content"`
}

func (Notification) TableName() string {
	return ploutos.TableNameNotifications
}

type _ = ploutos.UserNotification
type UserNotificationWithNotification struct {
	ID             int64         `gorm:"primarykey" json:"id"` // 主键ID
	UserId         int64         `json:"user_id" form:"user_id" gorm:"column:user_id"`
	Text           string        `json:"text" form:"text" gorm:"column:text"`
	NotificationId int64         `json:"notification_id" form:"notification_id" gorm:"column:notification_id"`
	IsRead         bool          `json:"is_read" form:"is_read" gorm:"column:is_read"`
	Notification   *Notification `json:"notification,omitempty" form:"-" gorm:"references:NotificationId;foreignKey:ID"`
}

func (UserNotificationWithNotification) TableName() string {
	return "user_notifications"
}

// FindOne
// categoryType references [ploutos.Notification.Category]
// notificationId references [ploutos.Notification]
// userNotificationId references [ploutos.UserNotification]
func FindOne(ctx context.Context, user model.User, categoryType ploutos.NotificationCategoryType, notificationId int64, userNotificationId int64) (UserNotificationWithNotification, error) {
	var notif UserNotificationWithNotification
	switch categoryType {
	case ploutos.NotificationCategoryTypeSystem, 0:
		err := model.DB.Debug().Model(UserNotificationWithNotification{}).Where("user_id = ?", user.ID).Scopes(model.ByIds([]int64{userNotificationId})).Find(&notif).Error
		if err != nil {
			return UserNotificationWithNotification{}, err
		}
	case ploutos.NotificationCategoryTypeNotification:
	case ploutos.NotificationCategoryTypeGame, ploutos.NotificationCategoryTypePromotion:
		fallthrough
	case ploutos.NotificationCategoryTypeSportsBet,
		ploutos.NotificationCategoryTypeStream:
		err := model.DB.Debug().Model(UserNotificationWithNotification{}).Where("user_id = ?", user.ID).Joins("Notification").Scopes(model.ByIds([]int64{userNotificationId})).Find(&notif).Error
		if err != nil {
			return UserNotificationWithNotification{}, err
		}
	default:
		return UserNotificationWithNotification{}, errors.New("unknown notification category")
	}

	return notif, nil
}
