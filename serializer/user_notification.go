package serializer

import (
	"fmt"
	"strconv"
	"strings"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type UserNotification struct {
	ID     int64  `json:"id"`
	Text   string `json:"text"`
	Ts     int64  `json:"ts"`
	IsRead bool   `json:"is_read"`
}

func BuildUserNotification(c *gin.Context, a ploutos.UserNotification) (b UserNotification) {
	b = UserNotification{
		ID:     a.ID,
		Text:   a.Text,
		IsRead: a.IsRead,
	}
	if !a.CreatedAt.IsZero() {
		b.Ts = a.CreatedAt.Unix()
	}
	return
}

type NotificationIdsWithReadStatus struct {
	ID     int64
	IsRead bool
}

type UserNotificationResponseV2 struct {
	Notifications []UserNotificationV2             `json:"notifications"`
	Counts        []UserNotificationUnreadCountsV2 `json:"counts"`
}

type UserNotificationUnreadCountsV2 struct {
	ID           int    `json:"id"`
	Label        string `json:"label"`
	UnreadCounts int    `json:"unread_counts"`
}

type UserNotificationV2 struct {
	ID                int64  `json:"id"`
	ReferenceId       string `json:"reference_id"`
	Title             string `json:"text"`
	ImageUrl          string `json:"image_url"`
	ShortContent      string `json:"short_content"`
	Category          int    `json:"category"`
	CategoryContentId int    `json:"category_content_id"`
	Ts                int64  `json:"ts"`
	ExpiredAt         int64  `json:"expire_at"`
	IsRead            bool   `json:"is_read"`
}

func BuildUserNotificationV2(c *gin.Context, a ploutos.UserNotification) (b UserNotificationV2) {
	b = UserNotificationV2{
		ID:                a.ID,
		ReferenceId:       fmt.Sprintf("user_notification-%v", a.ID),
		Title:             "System Message",
		ImageUrl:          "",
		ShortContent:      a.Text,
		Category:          999,
		CategoryContentId: 0,
		ExpiredAt:         1920336779,
		IsRead:            a.IsRead,
	}
	if !a.CreatedAt.IsZero() {
		b.Ts = a.CreatedAt.Unix()
	}
	return
}

type NotificationReferenceId string

// BuildNotificationReferenceId
// inverse [DissectCompositeIds]
func BuildNotificationReferenceId(notificationdIdType string, userNotificationId int64) NotificationReferenceId {
	return NotificationReferenceId(fmt.Sprintf("%s-%d", notificationdIdType, userNotificationId))
}

func (refId NotificationReferenceId) IsUserNotificationId() (bool, int64, error) {
	vals := strings.Split(string(refId), "-")
	if len(vals) != 2 {
		return false, 0, fmt.Errorf("invalid reference id length (!= 2)")
	}

	if vals[0] != "user_notification" {
		return false, 0, nil
	}

	id, err := strconv.Atoi(vals[1])
	if err != nil {
		return false, 0, err
	}
	return true, int64(id), nil
}

func (refId NotificationReferenceId) IsNotificationId() (bool, int64, error) {
	vals := strings.Split(string(refId), "-")
	if len(vals) != 2 {
		return false, 0, fmt.Errorf("invalid reference id length (!= 2)")
	}

	if vals[0] != "notification" {
		return false, 0, nil
	}
	id, err := strconv.Atoi(vals[1])
	if err != nil {
		return false, 0, err
	}
	return true, int64(id), nil
}

func BuildCMSNotificationV2(_ *gin.Context, a ploutos.Notification, notifications_ids_with_read_status []NotificationIdsWithReadStatus, image_url string) (b UserNotificationV2) {
	isRead := false
	for _, notification_ids_with_read_status := range notifications_ids_with_read_status {
		if a.ID == notification_ids_with_read_status.ID {
			isRead = notification_ids_with_read_status.IsRead
		}
	}

	b = UserNotificationV2{
		ID:                a.ID,
		ReferenceId:       fmt.Sprintf("notification-%v", a.ID),
		Title:             a.Title,
		ImageUrl:          Url(image_url),
		ShortContent:      a.ShortContent,
		Category:          int(a.Category),
		CategoryContentId: int(a.CategoryContentID),
		IsRead:            isRead,
	}

	if !a.SendAt.IsZero() {
		b.Ts = a.SendAt.Unix()
	}

	if !a.ExpiredAt.IsZero() {
		b.ExpiredAt = a.ExpiredAt.Unix()
	}
	return
}
