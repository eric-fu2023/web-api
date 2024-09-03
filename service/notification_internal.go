package service

import (
	"fmt"
	"math/rand"
	"strconv"
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
	spinNote       = "spin"
	minIndex       = 0
	maxIndex       = 2
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
		popUpTitle := []string{conf.GetI18N(lang).T(common.NOTIFICATION_POPUP_WINLOSE_FIRST_TITLE), conf.GetI18N(lang).T(common.NOTIFICATION_POPUP_WINLOSE_SECOND_TITLE), conf.GetI18N(lang).T(common.NOTIFICATION_POPUP_WINLOSE_THIRD_TITLE)}
		popUpWinDesc := []string{conf.GetI18N(lang).T(common.NOTIFICATION_POPUP_WIN_FIRST_DESC), conf.GetI18N(lang).T(common.NOTIFICATION_POPUP_WIN_SECOND_DESC), conf.GetI18N(lang).T(common.NOTIFICATION_POPUP_WIN_THIRD_DESC)}
		popUpLoseDesc := []string{conf.GetI18N(lang).T(common.NOTIFICATION_POPUP_LOSE_FIRST_DESC), conf.GetI18N(lang).T(common.NOTIFICATION_POPUP_LOSE_SECOND_DESC), conf.GetI18N(lang).T(common.NOTIFICATION_POPUP_LOSE_THIRD_DESC)}

		win_lose, _ := strconv.ParseInt(p.Params["win_lose_amount"], 10, 64)
		rank, _ := strconv.ParseInt(p.Params["ranking"], 10, 64)

		// randomly pick titles between the 3
		// get the win/lose value from params and determine whether to display win/lose desc
		randIndex := rand.Intn(maxIndex - minIndex + 1)

		if win_lose < 0 {
			// meaning the user is currently losing money.
			lose := win_lose * -1 // negate
			ranking := rank * -1

			text = fmt.Sprintf(popUpLoseDesc[randIndex], lose, ranking)

		} else {
			text = fmt.Sprintf(popUpWinDesc[randIndex], win_lose, rank)
		}

		title = popUpTitle[randIndex]

	case spinNote:
		notificationType = consts.Notification_Type_Spin
		spinTitle := []string{conf.GetI18N(lang).T(common.NOTIFICATION_SPIN_FIRST_TITLE), conf.GetI18N(lang).T(common.NOTIFICATION_SPIN_SECOND_TITLE), conf.GetI18N(lang).T(common.NOTIFICATION_SPIN_THIRD_TITLE)}
		spinDesc := []string{conf.GetI18N(lang).T(common.NOTIFICATION_SPIN_FIRST_DESC), conf.GetI18N(lang).T(common.NOTIFICATION_SPIN_SECOND_DESC), conf.GetI18N(lang).T(common.NOTIFICATION_SPIN_THIRD_DESC)}

		randIndex := rand.Intn(maxIndex - minIndex + 1)

		title = spinTitle[randIndex]
		text = spinDesc[randIndex]
	}
	common.SendNotification(p.UserID, notificationType, title, text)
	r.Data = "Success"
	return
}
