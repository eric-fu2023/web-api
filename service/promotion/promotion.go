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

	list, err := model.PromotionList(c, brand, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	parentIdToPromotionMap := make(map[int64][]serializer.PromotionCover)
	promotionCoverList := []serializer.PromotionCover{}
	for _, promotion := range list {
		isAllowDevice := false
		// Skip if not allow device.platform
		if len(promotion.DisplayDevices) != 0 {
			for _, allowedDevices := range promotion.DisplayDevices {
				if PromotionDevice[deviceInfo.Platform] == models.PromotionDeviceType(allowedDevices) {
					isAllowDevice = true
				}
			}
		} else {
			isAllowDevice = true
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
			parentIdToPromotionMap[promotion.ParentId] = append(parentIdToPromotionMap[promotion.ParentId], promotionCover)
		} else {
			if promotion.LoginStatus == int32(models.CustomPromotionLoginStatusAny) || (promotion.LoginStatus == int32(models.CustomPromotionLoginStatusLogin) && user.ID != 0) {
				promotionCoverList = append(promotionCoverList, promotionCover)
			}
		}
	}

	for i, promotionCover := range promotionCoverList {
		childrenPromotions, exists := parentIdToPromotionMap[promotionCover.ID]
		if exists {
			promotionCoverList[i].IsCustom = false
			promotionCoverList[i].ChildrenPromotions = childrenPromotions
		}
	}

	r.Data = promotionCoverList
	return
}

type PromotionDetail struct {
	ID int64 `form:"id" json:"id"`
}

func (p PromotionDetail) Handle(gCtx *gin.Context) (r serializer.Response, err error) {
	//ctx := contextify.AppendCtx(context.Background(), contextify.DefaultContextKey, fmt.Sprintf("%d (p PromotionDetail) Handle ", time.Now().UnixNano()))
	now := time.Now()
	//ctx = contextify.AppendCtx(ctx, contextify.DefaultContextKey, fmt.Sprintf("now %#v", now.String()))

	brand := gCtx.MustGet(`_brand`).(int)
	//ctx = contextify.AppendCtx(ctx, contextify.DefaultContextKey, fmt.Sprintf("brand %#v", brand))

	u, loggedIn := gCtx.Get("user")
	//ctx = contextify.AppendCtx(ctx, contextify.DefaultContextKey, fmt.Sprintf("loggedIn %#v", loggedIn))

	user, _ := u.(model.User)
	//ctx = contextify.AppendCtx(ctx, contextify.DefaultContextKey, fmt.Sprintf("user %#v", user))

	deviceInfo, _ := util.GetDeviceInfo(gCtx)
	//ctx = contextify.AppendCtx(ctx, contextify.DefaultContextKey, fmt.Sprintf("deviceInfo %#v", deviceInfo))

	var promotion models.Promotion
	if p.ID == 99999 {
		// TODO : remove mock data
		promotion = models.Promotion{
			Type: models.PromotionTypeNewbie,
		}
	} else {
		promotion, err = model.PromotionGetActive(gCtx, brand, p.ID, now)

		//ctx = contextify.AppendCtx(ctx, contextify.DefaultContextKey, fmt.Sprintf("[model.PromotionGetActive = promotion %v, promotion type %v,err %#v]", promotion, promotion.Type, err))
	}

	if err != nil {
		r = serializer.Err(gCtx, p, serializer.CodeGeneralError, "", err)
		//ctx = contextify.AppendCtx(ctx, contextify.DefaultContextKey, fmt.Sprintf("err != nil .returning err %#v", r))
		return
	}
	// tz := time.FixedZone("local", int(promotion.Timezone))
	// now = now.In(tz)

	var (
		progress    int64
		reward      int64
		claimStatus serializer.ClaimStatus
		voucherView serializer.Voucher
		extra       any
		session     models.PromotionSession

		customData any
		newbieData any
	)

	switch promotion.Type {

	case models.PromotionTypeCustomTemplate:
		customData = "anything"

	case models.PromotionTypeNewbie:
		newbieData = serializer.BuildDummyNewbiePromotion()

	default: // default promotion type..
		_session, err := model.PromotionSessionGetActive(gCtx, p.ID, now)
		if err != nil {
			r = serializer.Err(gCtx, p, serializer.CodeGeneralError, "", err)
			return r, err
		}
		session = _session
		if loggedIn {
			progress = ProgressByType(gCtx, promotion, session, user.ID, now)
			claimStatus = ClaimStatusByType(gCtx, promotion, session, user.ID, now)
			reward, _, _, err = RewardByType(gCtx, promotion, session, user.ID, progress, now, &user)
			extra = ExtraByType(gCtx, promotion, session, user.ID, progress, now)
			//ctx = contextify.AppendCtx(gCtx, contextify.DefaultContextKey, fmt.Sprintf("default promo type, user logged in. progress %#v, claimStatus %#v, reward %#v, extra %#v",
			//	progress,
			//	claimStatus,
			//	reward,
			//	extra,
			//))

			//log.Printf("%s\n", ctx.Value(contextify.DefaultContextKey))
		}
		if claimStatus.HasClaimed {
			v, err := model.VoucherGetByUserSession(gCtx, user.ID, session.ID)
			if err != nil {
			} else {
				voucherView = serializer.BuildVoucher(v, deviceInfo.Platform)
			}
		} else {
			v, err := model.VoucherTemplateGetByPromotion(gCtx, p.ID)
			if err != nil {
			} else {
				voucherView = serializer.BuildVoucherFromTemplate(v, reward, deviceInfo.Platform)
			}
		}
	}

	r.Data = serializer.BuildPromotionDetail(progress, reward, deviceInfo.Platform, promotion, session, voucherView, claimStatus, extra, customData, newbieData)

	//ctx = contextify.AppendCtx(ctx, contextify.DefaultContextKey, fmt.Sprintf("r.Data %#v", r.Data))
	//log.Printf("%s\n", ctx.Value(contextify.DefaultContextKey))

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
		parentPromotion, err = model.PromotionGetActive(c, brand, promotion.ParentId, now)
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
