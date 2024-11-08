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
type GeneralNotification struct {
	ID             int64         `gorm:"primarykey" json:"id"` // 主键ID
	UserId         int64         `json:"user_id" form:"user_id" gorm:"column:user_id"`
	Text           string        `json:"text" form:"text" gorm:"column:text"`
	NotificationId int64         `json:"notification_id" form:"notification_id" gorm:"column:notification_id"`
	IsRead         bool          `json:"is_read" form:"is_read" gorm:"column:is_read"`
	Notification   *Notification `json:"notification,omitempty" form:"-" gorm:"references:NotificationId;foreignKey:ID"`
}

func (GeneralNotification) TableName() string {
	return "user_notifications"
}

// FindGeneralOne
// categoryType references [ploutos.Notification.Category]
// notificationId references [ploutos.Notification]
// userNotificationId references [ploutos.UserNotification]
func FindGeneralOne(ctx context.Context, user model.User, categoryType ploutos.NotificationCategoryType, notificationId int64, userNotificationId int64) (GeneralNotification, error) {
	var notif GeneralNotification
	switch categoryType {
	case ploutos.NotificationCategoryTypeNotification:
		err := model.DB.Debug().Model(GeneralNotification{}).Where("user_id = ?", user.ID).Scopes(model.ByIds([]int64{userNotificationId})).Find(&notif).Error
		if err != nil {
			return GeneralNotification{}, err
		}
	default:
		return GeneralNotification{}, errors.New("unknown or invalid notification category")
	}

	return notif, nil
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
	//	var marker MarkReadNotificationInterface
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

	switch notification.CategoryType {
	case ploutos.NotificationCategoryTypeSystem:
		_ = CategoryTypeSystemMarker{
			UserId: userId,
		}.UserId

	//case ploutos.NotificationCategoryTypeNotification:

	default:
		return nil, errors.New("MarkNotificationsAsRead: unknown or invalid notification category")
	}

	return nil, nil
}

type MarkReadNotificationInterface interface {
	GetOrCreateUserNotification(ctx context.Context) (ploutos.UserNotification, error)
	MarkUserNotification(ctx context.Context) (ploutos.UserNotification, error)
	Mark() (ploutos.UserNotification, error)
}

var _ MarkReadNotificationInterface = &CategoryTypeSystemMarker{}

type CategoryTypeSystemMarker struct {
	UserId int64
}

func (n *CategoryTypeSystemMarker) GetOrCreateUserNotification(context.Context) (ploutos.UserNotification, error) {
	return ploutos.UserNotification{}, errors.New("implement me")
}

func (n *CategoryTypeSystemMarker) MarkUserNotification(ctx context.Context) (ploutos.UserNotification, error) {
	return ploutos.UserNotification{}, errors.New("implement me")
}

func (n *CategoryTypeSystemMarker) Mark() (ploutos.UserNotification, error) {
	return ploutos.UserNotification{}, errors.New("implement me")
}
