package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"
	"web-api/cache"
	"web-api/conf"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"

	"github.com/gin-gonic/gin"
	"github.com/leekchan/accounting"
)

const (
	TEAMUP_GAME_POPUP_NOTIFICATION_TYPE = "popup_teamup_game"
	vipPromoteNote                      = "vip_promotion"
	popUpNote                           = "popup_winlose"
	spinNote                            = "spin"
	minIndex                            = 0
	maxIndex                            = 2
)

type InternalNotificationPushRequest struct {
	UserID int64             `form:"user_id" json:"user_id" binding:"required"`
	Type   string            `form:"type" json:"type" binding:"required"`
	Params map[string]string `form:"params" json:"params"`
}
type InternalNotificationPushAllRequest struct {
	Type   string            `form:"type" json:"type" binding:"required"`
	Lang   string            `form:"lang" json:"lang" binding:"required"`
	Params map[string]string `form:"params" json:"params"`
}

func (p InternalNotificationPushRequest) Handle(c *gin.Context) (r serializer.Response) {
	var notificationType, title, text string
	var resp serializer.Response
	brand := c.MustGet(`_brand`).(int)

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

		win_lose, _ := strconv.ParseFloat(p.Params["win_lose_amount"], 64)
		rank, _ := strconv.ParseInt(p.Params["ranking"], 10, 64)

		// randomly pick titles between the 3
		// get the win/lose value from params and determine whether to display win/lose desc
		randIndex := rand.Intn(maxIndex - minIndex + 1)

		if win_lose < 0 {
			// meaning the user is currently losing money.
			lose := win_lose * -1 // negate
			ranking := rank * -1

			value := lose / 100
			// check whether user's lang is en or zh
			if lang == "en" {

				text = fmt.Sprintf(popUpLoseDesc[randIndex], FormatINR(value), ranking)
			} else {
				ac := accounting.Accounting{Symbol: "$", Precision: 2}
				text = fmt.Sprintf(popUpLoseDesc[randIndex], ac.FormatMoney(value), ranking)
			}
		} else {
			value := win_lose / 100
			// check whether user's lang is en or zh
			if lang == "en" {
				text = fmt.Sprintf(popUpWinDesc[randIndex], FormatINR(value), rank)
			} else {
				ac := accounting.Accounting{Symbol: "$", Precision: 2}
				text = fmt.Sprintf(popUpWinDesc[randIndex], ac.FormatMoney(value), rank)

			}
		}

		title = popUpTitle[randIndex]

		winLoseResp, err := WinLoseMetadata(p.UserID)
		if err != nil {
			log.Println("Unable to obtain win_lose response from WinLoseMetadata function")
			return
		}

		// metadata needed for front end to navigate to a particular screen.
		resp.Data = PopupResponse{
			Type:  1,
			Float: []PopupFloat{},
			Data:  winLoseResp,
		}

		log.Printf("response data for win lose pop up: %+v", resp.Data)

	case TEAMUP_GAME_POPUP_NOTIFICATION_TYPE:
		notificationType = p.Type
		amount, _ := strconv.ParseInt(p.Params["amount"], 10, 64)
		endTime, _ := strconv.ParseInt(p.Params["end_time"], 10, 64)
		teamupId, _ := strconv.ParseInt(p.Params["teamup_id"], 10, 64)
		gameType, _ := strconv.ParseInt(p.Params["game_type"], 10, 64)

		teamup, _ := model.GetTeamUpByTeamUpId(teamupId)

		var resp serializer.Response
		providerName := consts.GameProviderNameMap[p.Params["provider"]]
		providerIcon := consts.GameProviderNameToImgMap[p.Params["provider"]]

		_, teamupType := model.GetGameTypeSlice(brand, int(gameType))

		resp.Data = TeamupGamePopUpNotification{
			TeamupId:     int64(teamupId),
			EndTime:      int64(endTime),
			Amount:       int64(amount) / 100,
			ProviderName: consts.GameProviderNameMap[p.Params["provider"]],
			Icon:         consts.GameProviderNameToImgMap[p.Params["provider"]],
			OrderId:      teamup.OrderId,
			TeamupType:   int64(teamupType),
		}

		title = conf.GetI18N(lang).T("notification_teamup_start_game_title")
		text = conf.GetI18N(lang).T("notification_teamup_start_game_content")

		// common.SendNotification(3697, notificationType, title, text, resp)                                                                                                       // DEBUG PURPOSE
		// common.SendTeamupGamePopupNotificationSocketMsg(3697, int64(teamupId), int64(endTime), int64(amount)/100, int64(teamupType), providerName, providerIcon, teamup.OrderId) // DEBUG PURPOSE

		// common.SendNotification(3713, notificationType, title, text, resp)                                                                                                       // DEBUG PURPOSE
		// common.SendTeamupGamePopupNotificationSocketMsg(3713, int64(teamupId), int64(endTime), int64(amount)/100, int64(teamupType), providerName, providerIcon, teamup.OrderId) // DEBUG PURPOSE

		// common.SendNotification(3722, notificationType, title, text, resp)                                                                                                       // DEBUG PURPOSE
		// common.SendTeamupGamePopupNotificationSocketMsg(3722, int64(teamupId), int64(endTime), int64(amount)/100, int64(teamupType), providerName, providerIcon, teamup.OrderId) // DEBUG PURPOSE

		// common.SendNotification(3761, notificationType, title, text, resp)                                                                                                       // DEBUG PURPOSE
		// common.SendTeamupGamePopupNotificationSocketMsg(3761, int64(teamupId), int64(endTime), int64(amount)/100, int64(teamupType), providerName, providerIcon, teamup.OrderId) // DEBUG PURPOSE

		// common.SendNotification(455, notificationType, title, text, resp)                                                                                                       // DEBUG PURPOSE
		// common.SendTeamupGamePopupNotificationSocketMsg(455, int64(teamupId), int64(endTime), int64(amount)/100, int64(teamupType), providerName, providerIcon, teamup.OrderId) // DEBUG PURPOSE

		// common.SendNotification(3621, notificationType, title, text, resp)                                                                                                       // DEBUG PURPOSE
		// common.SendTeamupGamePopupNotificationSocketMsg(3621, int64(teamupId), int64(endTime), int64(amount)/100, int64(teamupType), providerName, providerIcon, teamup.OrderId) // DEBUG PURPOSE

		common.SendTeamupGamePopupNotificationSocketMsg(p.UserID, int64(teamupId), int64(endTime), int64(amount)/100, int64(teamupType), providerName, providerIcon, teamup.OrderId)
	}
	common.SendNotification(p.UserID, notificationType, title, text, resp)
	r.Data = "Success"
	return
}

func (p InternalNotificationPushAllRequest) HandleAll(c *gin.Context) (r serializer.Response) {
	var resp serializer.Response

	lang := p.Lang
	notificationType := consts.Notification_Type_Spin
	spinTitle := []string{conf.GetI18N(lang).T(common.NOTIFICATION_SPIN_FIRST_TITLE), conf.GetI18N(lang).T(common.NOTIFICATION_SPIN_SECOND_TITLE), conf.GetI18N(lang).T(common.NOTIFICATION_SPIN_THIRD_TITLE)}
	spinDesc := []string{conf.GetI18N(lang).T(common.NOTIFICATION_SPIN_FIRST_DESC), conf.GetI18N(lang).T(common.NOTIFICATION_SPIN_SECOND_DESC), conf.GetI18N(lang).T(common.NOTIFICATION_SPIN_THIRD_DESC)}

	randIndex := rand.Intn(maxIndex - minIndex + 1)

	title := spinTitle[randIndex]
	text := spinDesc[randIndex]

	popUpSpinId, floats, err := SpinMetadata()

	if err != nil {
		log.Println("Unable to obtain popUpSpinId from SpinMetadata function")
		return
	}

	resp.Data = PopupResponse{
		Type:  5,
		Float: floats,
		Data:  popUpSpinId,
	}

	log.Printf("response data for spin pop up float: %v data: %+v", floats, popUpSpinId)

	common.PushNotificationAll(notificationType, title, text, resp)
	r.Data = "Success"
	return
}

func FormatINR(val float64) string {
	if val >= 10000000 {
		newValue := val / 10000000
		if newValue == float64(int(newValue)) {
			return fmt.Sprintf("%v crore", newValue)
		}
		return fmt.Sprintf("%.2f crore", newValue)
	} else if val >= 100000 {
		newValue := val / 100000
		if newValue == float64(int(newValue)) {
			return fmt.Sprintf("%v lakh", newValue)
		}
		return fmt.Sprintf("%.2f lakh", newValue)
	}

	if val == float64(int(val)) {
		return fmt.Sprintf("%v", val)
	}
	return fmt.Sprintf("%.2f", val)
}

func SpinMetadata() (PopupSpinId, []PopupFloat, error) {
	// condition : 1 = app start , 2 = app resume.
	popup_types, err := model.GetPopupList(1)
	if err != nil {
		log.Println("SpinMetadata function: get PopupTypes error ", err)
		return PopupSpinId{}, []PopupFloat{}, err
	}
	var floats []PopupFloat
	var spin_promotion_id_int int
	for _, popup_type := range popup_types {
		if popup_type.PopupType == 5 && popup_type.CanFloat {
			// spin popup float
			spin_promotion_id_int, _ = strconv.Atoi(popup_type.Meta)
			// user still can spin, then we add the spin popup to float list.
			floats = append(floats, PopupFloat{
				Type: 5,
				Id:   spin_promotion_id_int,
			})
		}
	}

	spin_id_data := PopupSpinId{
		SpinId: spin_promotion_id_int,
	}

	return spin_id_data, floats, nil
}

func WinLoseMetadata(userId int64) (WinLosePopupResponse, error) {
	var user model.User

	// find user based on userID
	if err := model.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		return WinLosePopupResponse{}, err
	}

	now := time.Now()
	yesterdayStart := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	yesterdayEnd := yesterdayStart.Add(24 * time.Hour)

	// check if user has GGR yesterday
	type GGRRecords struct {
		GGR    float64 `json:"win_lose"`
		UserID int64   `json:"user_id"`
	}

	key := "popup/win_lose/" + now.Format("2006-01-02")
	total_ranking_key := "popup/win_lose/total_ranking/" + now.Format("2006-01-02")
	current_ranking_key := "popup/win_lose/ranking/" + now.Format("2006-01-02")

	res := cache.RedisClient.HGet(context.Background(), key, strconv.FormatInt(user.ID, 10))
	GGR, err := strconv.ParseFloat(res.Val(), 64)
	if err != nil {
		fmt.Println("convert GGR to float64 failed!!!!", err.Error())
	}
	myGGRRecord := GGRRecords{
		GGR:    GGR,
		UserID: user.ID,
	}

	var total_ranking int
	var current_ranking int
	current_ranking_string := cache.RedisClient.HGet(context.Background(), current_ranking_key, strconv.FormatInt(user.ID, 10)).Val()

	if GGR < 0 {
		total_ranking_string := cache.RedisClient.HGet(context.Background(), total_ranking_key, "lose").Val()
		total_ranking, _ = strconv.Atoi(total_ranking_string)
		current_ranking, _ = strconv.Atoi(current_ranking_string)
		current_ranking = -current_ranking
	} else {
		total_ranking_string := cache.RedisClient.HGet(context.Background(), total_ranking_key, "win").Val()
		total_ranking, _ = strconv.Atoi(total_ranking_string)
		current_ranking, _ = strconv.Atoi(current_ranking_string)
	}

	var members []WinLosePopupGGR
	if GGR > 0 {
		members = append(members,
			generateMemberGGR(user, myGGRRecord.GGR, rand.Intn(500), current_ranking, false, -1),
			generateMemberGGR(user, myGGRRecord.GGR, 0, current_ranking, true, 0),
			generateMemberGGR(user, myGGRRecord.GGR, -rand.Intn(500), current_ranking, false, 1))
	} else if GGR < 0 {
		members = append(members,
			generateMemberGGR(user, myGGRRecord.GGR, rand.Intn(500)+500, current_ranking, false, 2),
			generateMemberGGR(user, myGGRRecord.GGR, rand.Intn(500), current_ranking, false, 1),
			generateMemberGGR(user, myGGRRecord.GGR, 0, current_ranking, true, 0))
	}
	data := WinLosePopupResponse{
		CurrentRanking: current_ranking,
		TotalRanking:   total_ranking,
		Start:          yesterdayStart.Unix(),
		End:            yesterdayEnd.Unix(),
		GGR:            myGGRRecord.GGR / 100,
		IsWin:          myGGRRecord.GGR > 0,
		Member:         members,
	}

	return data, nil

}
