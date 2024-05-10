package service

import (
	"fmt"
	"web-api/conf"
	"web-api/conf/consts"
	"web-api/serializer"
	"web-api/service/common"

	"github.com/gin-gonic/gin"
)

const (
	vipPromoteNote = "vip_promotion"
)

type InternalNotificationPushRequest struct {
	UserID int64             `form:"user_id" json:"user_id" binding:"required"`
	Type   string            `form:"type" json:"type" binding:"required"`
	Params map[string]string `form:"params" json:"params"`
}

func (p InternalNotificationPushRequest) Handle(c *gin.Context) (r serializer.Response) {
	var notificationType, title, text string
	switch p.Type {
	case vipPromoteNote:
		notificationType = consts.Notification_Type_Vip_Promotion
		title = conf.GetI18N(conf.GetDefaultLocale()).T(common.NOTIFICATION_VIP_PROMOTION_TITLE)
		text = fmt.Sprintf(conf.GetI18N(conf.GetDefaultLocale()).T(common.NOTIFICATION_VIP_PROMOTION), p.Params["vip_level"])
	}
	common.SendNotification(p.UserID, notificationType, title, text)
	r.Data = "Success"
	return
}
