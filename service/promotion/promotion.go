package promotion

import (
	"encoding/json"
	"fmt"
	"time"

	"web-api/model"
	"web-api/serializer"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type PromotionList struct {
	IsLoggedIn bool `json:"is_logged_in" form:"is_logged_in"`
}

func (p PromotionList) ListCategories(c *gin.Context) (r serializer.Response, err error) {
	detail, err := models.GetDictionaryValues("promotionCategory", model.DB)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	r.Data = util.MapSlice(detail, serializer.BuildSysDictionaryDetail)
	return
}

var PromotionDevice = map[string]models.PromotionDeviceType{
	"pc":      models.PromotionDevicePC,
	"h5":      models.PromotionDeviceH5orM,
	"m":       models.PromotionDeviceH5orM,
	"ios":     models.PromotionDeviceIOS,
	"android": models.PromotionDeviceAndroid}

func (p PromotionList) Handle(c *gin.Context) (r serializer.Response, err error) {
	now := time.Now()
	brand := c.MustGet(`_brand`).(int)
	deviceInfo, _ := util.GetDeviceInfo(c)

	u, _ := c.Get("user")
	var user model.User
	if u != nil {
		user = u.(model.User)
	}

	promotions, err := model.OngoingPromotions(c, brand, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	childPromotionCoversMap := make(map[int64][]serializer.PromotionCover)
	promotionCovers := []serializer.PromotionCover{}
	for _, promotion := range promotions {
		isAllowDevice := false
		// Skip if not allow device.platform
		switch allowedDevices := promotion.DisplayDevices; len(allowedDevices) {
		case 0:
			isAllowDevice = true
		default:
			for _, allowedDevice := range allowedDevices {
				if PromotionDevice[deviceInfo.Platform] == models.PromotionDeviceType(allowedDevice) {
					isAllowDevice = true
					break
				}
			}
		}

		if !isAllowDevice {
			continue
		}

		promotionCover := serializer.BuildPromotionCover(promotion, deviceInfo.Platform)
		if promotionCover.ParentId != 0 {
			content := serializer.IncomingPromotionMatchList{}
			_ = json.Unmarshal(promotionCover.SubpageContent, &content)
			promotionCover.Name = content.Title
			promotionCover.Title = content.Title
			childPromotionCoversMap[promotion.ParentId] = append(childPromotionCoversMap[promotion.ParentId], promotionCover)
		} else {
			if promotion.LoginStatus == int32(models.CustomPromotionLoginStatusAny) || (promotion.LoginStatus == int32(models.CustomPromotionLoginStatusLogin) && user.ID != 0) {
				promotionCovers = append(promotionCovers, promotionCover)
			}
		}
	}

	for i, promotionCover := range promotionCovers {
		childrenPromotions, exists := childPromotionCoversMap[promotionCover.ID]
		if exists {
			promotionCovers[i].IsCustom = false
			promotionCovers[i].ChildrenPromotions = childrenPromotions
		}
	}

	r.Data = promotionCovers
	return
}

type PromotionDetail struct {
	ID int64 `form:"id" json:"id"`
}

func (p PromotionDetail) Handle(gCtx *gin.Context) (r serializer.Response, err error) {
	now := time.Now()

	brand := gCtx.MustGet(`_brand`).(int)

	u, hasUserInfo := gCtx.Get("user")

	user, _ := u.(model.User)

	deviceInfo, _ := util.GetDeviceInfo(gCtx)

	var promotion models.Promotion
	if p.ID == 99999 {
		// TODO : remove mock data
		promotion = models.Promotion{
			Type: models.PromotionTypeNewbie,
		}
	} else {
		promotion, err = model.OngoingPromotionById(gCtx, brand, p.ID, now)
	}

	if err != nil {
		r = serializer.Err(gCtx, p, serializer.CodeGeneralError, "", err)
		return
	}
	var (
		progress      int64
		reward        int64
		claimStatus   serializer.ClaimStatus
		voucherView   serializer.Voucher
		extra         any
		activeSession models.PromotionSession

		customData any
		newbieData any
		mDo        serializer.MissionDO
	)

	switch promotion.Type {

	case models.PromotionTypeDepositEarnMoreMission:
		topupRecords, getTopupsErr := model.TopupsByDateRange(gCtx, user.ID, promotion.StartAt, promotion.EndAt)
		if getTopupsErr != nil {
			err = getTopupsErr
			return
		}

		totalDepositedAmount := util.Sum(topupRecords, func(co model.CashOrder) int64 {
			return co.ActualCashInAmount
		})

		promotionMissions, getMissionErr := model.GetMissionByPromotionId(gCtx, brand, promotion.ID)
		if getMissionErr != nil {
			err = getMissionErr
			return
		}

		claimedVouchers, getVouchersErr := model.GetVouchersByUserAndPromotion(gCtx, user.ID, promotion.ID)
		if getMissionErr != nil {
			err = getVouchersErr
			return
		}

		mDo.Missions = promotionMissions
		mDo.CompletedMissions = claimedVouchers
		mDo.TotalDepositAmount = totalDepositedAmount

	case models.PromotionTypeCustomTemplate:
		customData = "anything"

	case models.PromotionTypeNewbie:
		newbieData = serializer.BuildDummyNewbiePromotion()

	default: // default promotion type..
		_session, activeSessionError := model.GetActivePromotionSessionByPromotionId(gCtx, p.ID, now)
		if activeSessionError != nil {
			r = serializer.Err(gCtx, p, serializer.CodeGeneralError, "", activeSessionError)
			return r, activeSessionError
		}
		activeSession = _session
		if hasUserInfo {
			progress, err = GetPromotionSessionProgress(gCtx, promotion, activeSession, user.ID)
			// FIXME
			// to remove error suppression
			// if err != nil && !errors.Is(err, ErrPromotionSessionUnknownPromotionType) {
			// 	log.Printf("!errors.Is(err, ErrPromotionSessionUnknownPromotionType) err, %v", err)
			// 	return r, err
			// }
			claimStatus = GetPromotionSessionClaimStatus(gCtx, promotion, activeSession, user.ID, now)
			reward, _, _, err = GetPromotionRewards(gCtx, promotion, user.ID, progress, now, &user)
			extra = GetPromotionExtraDetails(gCtx, promotion, user.ID, now)
		}
		if claimStatus.HasClaimed {
			v, err := model.GetVoucherByUserAndPromotionSession(gCtx, user.ID, activeSession.ID)
			if err != nil {
			} else {
				voucherView = serializer.BuildVoucher(v, deviceInfo.Platform)
			}
		} else {
			v, err := model.GetPromotionVoucherTemplateByPromotionId(gCtx, p.ID)
			if err != nil {
			} else {
				voucherView = serializer.BuildVoucherFromTemplate(v, reward, deviceInfo.Platform)
			}
		}
	}

	r.Data = serializer.BuildPromotionDetail(progress, reward, deviceInfo.Platform, promotion, activeSession, voucherView, claimStatus, extra, customData, newbieData, mDo)
	return
}

type PromotionCustomDetail struct {
	ID int64 `form:"id" json:"id"`
}

func (p PromotionCustomDetail) Handle(c *gin.Context) (r serializer.Response, err error) {
	if p.ID == 99999 { // TODO remove temporary
		return
	}

	now := time.Now()
	brand := c.MustGet(`_brand`).(int)
	u, _ := c.Get("user")
	user, _ := u.(model.User)
	// deviceInfo, _ := util.GetDeviceInfo(c)

	var parentPromotion models.Promotion
	var childPromotion models.Promotion

	promotion, err := model.PromotionGetActiveNoCheckStartEnd(c, brand, p.ID, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}

	outgoingRes := serializer.OutgoingCustomPromotionDetail{}

	if promotion.ParentId == 0 {
		parentPromotion = promotion
	} else {
		outgoingRes.IsCustomPromotion = false
		parentPromotion, err = model.OngoingPromotionById(c, brand, promotion.ParentId, now)
		if err != nil {
			r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
			return
		}
		childPromotion = promotion
	}

	outgoingRes.ParentInfo = serializer.IncomingPromotion{
		Id:   parentPromotion.ID,
		Name: parentPromotion.Name,
	}
	parentImages := serializer.IncomingPromotionImages{}
	err = json.Unmarshal([]byte(promotion.Image), &parentImages)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	outgoingRes.ParentInfo.Images = parentImages

	if childPromotion.ID == 0 {
		// IS PARENT
		subPromotions, err := model.PromotionGetSubActive(c, brand, parentPromotion.ID, now)
		if err != nil {
			fmt.Println(err)
		}

		for _, subPromo := range subPromotions {
			outgoingRes.IsCustomPromotion = false
			incomingMatchList := serializer.IncomingPromotionMatchList{}
			_ = json.Unmarshal(subPromo.SubpageContent, &incomingMatchList)

			if len(incomingMatchList.List) == 0 {
				continue
			}
			outgoingRes.ChildrenPromotions = append(outgoingRes.ChildrenPromotions, serializer.OutgoingCustomPromotionPreview{
				Id:    subPromo.ID,
				Title: incomingMatchList.Title,
			})
		}

		promotionPage := serializer.CustomPromotionPage{
			Title:       parentPromotion.Name,
			PromotionId: parentPromotion.ID,
		}

		if childPromotion.Action == nil {
			childPromotion.Action = parentPromotion.Action
			childPromotion.ID = parentPromotion.ID
		}

		incomingRequestAction := serializer.IncomingPromotionRequestAction{}
		err = json.Unmarshal(childPromotion.Action, &incomingRequestAction)
		if err != nil {
			fmt.Println(err)
		}

		outgoingRequestAction := serializer.BuildPromotionAction(c, incomingRequestAction, childPromotion.ID, user.ID)
		promotionPage.Action = outgoingRequestAction

		outgoingRes.PromotionInfo = promotionPage
	} else {
		// IS CHILD
		var content serializer.IncomingPromotionMatchList

		err = json.Unmarshal(childPromotion.SubpageContent, &content)
		if err != nil {
			fmt.Println(err)
		}

		promotionPage := serializer.CustomPromotionPage{
			Title:       parentPromotion.Name,
			PromotionId: parentPromotion.ID,
		}

		switch content.ListType {
		case "matches":
			promotionPage.PageItemListType = models.CustomPromotionTypeMatch
		case "game_vendor_brand":
			promotionPage.PageItemListType = models.CustomPromotionTypeGame
		default:
			promotionPage.PageItemListType = models.CustomPromotionTypeOthers
		}

		promotionPage.Desc = content.Desc

		customPromotionPageItem := serializer.BuildPromotionMatchList(content.List, childPromotion)
		promotionPage.PageItemList = customPromotionPageItem

		if childPromotion.Action == nil {
			childPromotion.Action = parentPromotion.Action
			childPromotion.ID = parentPromotion.ID
		}

		incomingRequestAction := serializer.IncomingPromotionRequestAction{}
		err = json.Unmarshal(childPromotion.Action, &incomingRequestAction)
		if err != nil {
			fmt.Println(err)
		}

		if len(incomingRequestAction.Fields) == 0 {
			err = json.Unmarshal(parentPromotion.Action, &incomingRequestAction)
			if err != nil {
				fmt.Println(err)
			}
		}

		outgoingRequestAction := serializer.BuildPromotionAction(c, incomingRequestAction, childPromotion.ID, user.ID)
		promotionPage.Action = outgoingRequestAction

		outgoingRes.PromotionInfo = promotionPage
		outgoingRes.ChildrenPromotions = make([]serializer.OutgoingCustomPromotionPreview, 0)
	}

	r.Data = outgoingRes
	return
}
