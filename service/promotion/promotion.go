package promotion

import (
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type PromotionList struct {
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
	r.Data = util.MapSlice(list, func(input models.Promotion) serializer.PromotionCover {
		return serializer.BuildPromotionCover(input, deviceInfo.Platform)
	})
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
	)
	if loggedIn {
		progress = ProgressByType(c, promotion, session, user.ID, now)
		claimStatus = ClaimStatusByType(c, promotion, session, user.ID, now)
		reward = RewardByType(c, promotion, session, user.ID, progress, now)
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

	r.Data = serializer.BuildPromotionDetail(progress, reward, deviceInfo.Platform, promotion, session, voucherView, claimStatus)
	return
}
