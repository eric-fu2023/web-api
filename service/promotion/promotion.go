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

func (p PromotionList) Handle(c *gin.Context) (r serializer.Response, err error) {
	now := time.Now()
	brand := c.MustGet(`_brand`).(int)
	deviceInfo, _ := util.GetDeviceInfo(c)

	// u, loggedIn := c.Get("user")
	// user := u.(model.User)
	list, err := model.PromotionList(c, brand, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	parentIdToPromotionMap := make(map[int64][]serializer.PromotionCover)
	promotionCoverList := []serializer.PromotionCover{}
	for _, promotion := range list {
		promotionCover := serializer.BuildPromotionCover(promotion, deviceInfo.Platform)
		if promotionCover.ParentId != 0 {
			parentIdToPromotionMap[promotion.ParentId] = append(parentIdToPromotionMap[promotion.ParentId], promotionCover)
		} else {
			promotionCoverList = append(promotionCoverList, promotionCover)
		}
	}

	for i, promotionCover := range promotionCoverList {
		childrenPromotions, exists := parentIdToPromotionMap[promotionCover.ID]
		if exists {
			promotionCoverList[i].IsCustom = true
			promotionCoverList[i].ChildrenPromotions = childrenPromotions
		}
	}

	r.Data = promotionCoverList
	return
}

type PromotionDetail struct {
	ID int64 `form:"id" json:"id"`
}

func (p PromotionDetail) Handle(c *gin.Context) (r serializer.Response, err error) {
	now := time.Now()
	brand := c.MustGet(`_brand`).(int)
	u, loggedIn := c.Get("user")
	user, _ := u.(model.User)
	deviceInfo, _ := util.GetDeviceInfo(c)

	promotion, err := model.PromotionGetActive(c, brand, p.ID, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	// tz := time.FixedZone("local", int(promotion.Timezone))
	// now = now.In(tz)
	session, err := model.PromotionSessionGetActive(c, p.ID, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	var (
		progress    int64
		reward      int64
		claimStatus serializer.ClaimStatus
		voucherView serializer.Voucher
		extra       any
	)
	if loggedIn {
		progress = ProgressByType(c, promotion, session, user.ID, now)
		claimStatus = ClaimStatusByType(c, promotion, session, user.ID, now)
		reward, _, _, err = RewardByType(c, promotion, session, user.ID, progress, now)
		extra = ExtraByType(c, promotion, session, user.ID, progress, now)
	}
	if claimStatus.HasClaimed {
		v, err := model.VoucherGetByUserSession(c, user.ID, session.ID)
		if err != nil {
		} else {
			voucherView = serializer.BuildVoucher(v, deviceInfo.Platform)
		}
	} else {
		v, err := model.VoucherTemplateGetByPromotion(c, p.ID)
		if err != nil {
		} else {
			voucherView = serializer.BuildVoucherFromTemplate(v, reward, deviceInfo.Platform)
		}
	}

	r.Data = serializer.BuildPromotionDetail(progress, reward, deviceInfo.Platform, promotion, session, voucherView, claimStatus, extra)
	return
}

type PromotionCustomDetail struct {
	ID int64 `form:"id" json:"id"`
}

func (p PromotionCustomDetail) Handle(c *gin.Context) (r serializer.Response, err error) {
	now := time.Now()
	brand := c.MustGet(`_brand`).(int)
	u, _ := c.Get("user")
	user, _ := u.(model.User)
	// deviceInfo, _ := util.GetDeviceInfo(c)

	var parentPromotion models.Promotion
	var childPromotion models.Promotion

	promotion, err := model.PromotionGetActive(c, brand, p.ID, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}

	outgoingRes := serializer.OutgoingCustomPromotionDetail{}

	if promotion.ParentId == 0 {
		parentPromotion = promotion
	} else {
		parentPromotion, err = model.PromotionGetActive(c, brand, promotion.ParentId, now)
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
			outgoingRes.ChildrenPromotions = append(outgoingRes.ChildrenPromotions, serializer.OutgoingCustomPromotionPreview{
				Id:    subPromo.ID,
				Title: subPromo.Name,
			})
		}
	} else {
		// IS CHILD
		var content serializer.IncomingPromotionMatchList

		err = json.Unmarshal(childPromotion.SubpageContent, &content)
		if err != nil {
			fmt.Println(err)
		}

		promotionPage := serializer.CustomPromotionPage{
			Title:       childPromotion.Name,
			PromotionId: childPromotion.ID,
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

		incomingRequestAction := serializer.IncomingPromotionRequestAction{}
		err = json.Unmarshal(childPromotion.Action, &incomingRequestAction)
		if err != nil {
			fmt.Println(err)
		}

		outgoingRequestAction := serializer.BuildPromotionAction(c, incomingRequestAction, childPromotion.ID, user.ID)
		promotionPage.Action = outgoingRequestAction

		outgoingRes.PromotionInfo = promotionPage
		outgoingRes.ChildrenPromotions = make([]serializer.OutgoingCustomPromotionPreview, 0)
	}

	r.Data = outgoingRes
	return
}
