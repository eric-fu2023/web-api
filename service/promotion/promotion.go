package promotion

import (
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"

	"github.com/gin-gonic/gin"
)

type PromotionList struct {
}

func (p PromotionList) Handle(c *gin.Context) (r serializer.Response, err error) {
	now := time.Now()
	brand := c.MustGet(`_brand`).(int)
	// u, loggedIn := c.Get("user")
	// user := u.(model.User)
	list, err := model.PromotionList(c, brand, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	r.Data = util.MapSlice(list, serializer.BuildPromotionCover)
	return
}

type PromotionDetail struct {
	ID int64 `form:"id" json:"id"`
}

func (p PromotionDetail) Handle(c *gin.Context) (r serializer.Response, err error) {
	now := time.Now()
	brand := c.MustGet(`_brand`).(int)
	u, loggedIn := c.Get("user")
	user := u.(model.User)
	promotion, err := model.PromotionGetActive(c, brand, p.ID, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	tz := time.FixedZone("local", int(promotion.Timezone))
	now = now.In(tz)
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
		reward = promotion.GetRewardDetails().GetReward(progress)
	}
	if claimStatus.HasClaimed {
		v, err := model.VoucherGetByUserSession(c, user.ID, session.ID)
		if err != nil {
		} else {
			voucherView = serializer.BuildVoucher(v)
		}
	} else {
		v, err := model.VoucherTemplateGetByPromotion(c, p.ID)
		if err != nil {
		} else {
			voucherView = serializer.BuildVoucherFromTemplate(v, reward)
		}
	}

	r.Data = serializer.BuildPromotionDetail(progress, reward, promotion, session, voucherView)
	return
}
