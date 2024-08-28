package service

import (
	"fmt"
	"web-api/conf"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"

	"github.com/gin-gonic/gin"
)

const (
	vipPromoteNote = "vip_promotion"
	popUpNote      = "popup_winlose"
)

type InternalNotificationPushRequest struct {
	UserID int64             `form:"user_id" json:"user_id" binding:"required"`
	Type   string            `form:"type" json:"type" binding:"required"`
	Params map[string]string `form:"params" json:"params"`
}

func (p InternalNotificationPushRequest) Handle(c *gin.Context) (r serializer.Response) {
	var notificationType, title, text string
	lang := model.GetUserLang(p.UserID)
	switch p.Type {
	case vipPromoteNote:
		notificationType = consts.Notification_Type_Vip_Promotion
		title = conf.GetI18N(lang).T(common.NOTIFICATION_VIP_PROMOTION_TITLE)
		vipName := p.Params["name"]
		if vipName == "" {
			vipName = p.Params["vip_level"]
		}
		text = fmt.Sprintf(conf.GetI18N(lang).T(common.NOTIFICATION_VIP_PROMOTION), vipName)
	case popUpNote:
		notificationType = consts.Notification_Type_Pop_Up
		title = conf.GetI18N(lang).T(common.NOTIFICATION_POPUP_WINLOSE_TITLE)
		text = conf.GetI18N(lang).T(common.NOTIFICATION_POPUP_WINLOSE)
	}
	common.SendNotification(p.UserID, notificationType, title, text)
	r.Data = "Success"
	return
}
