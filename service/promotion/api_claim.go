package promotion

import (
	"time"

	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"

	"github.com/gin-gonic/gin"
)

const (
	userPromotionSessionClaimKey = "user_promotion_session_claim_lock:%d:%d"
)

type PromotionClaim struct {
	ID int64 `form:"id" json:"id"`
}

func (p PromotionClaim) Handle(c *gin.Context) (r serializer.Response, err error) {
	now := time.Now().UTC()
	brand := c.MustGet(`_brand`).(int)
	user := c.MustGet("user").(model.User)
	deviceInfo, _ := util.GetDeviceInfo(c)
	i18n := c.MustGet("i18n").(i18n.I18n)

	promotion, err := model.OngoingPromotionById(c, brand, p.ID, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	// tz := time.FixedZone("local", int(promotion.Timezone))
	// now = now.In(tz)
	session, err := model.GetActivePromotionSessionByPromotionId(c, p.ID, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	voucher, err := Claim(c, now, promotion, session, user.ID, &user)
	if err != nil {
		switch err.Error() {
		case "double_claim":
			r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("double_claim"), err)
		case "unavailable_for_now":
			r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("unavailable_for_now"), err)
		case "nothing_to_claim":
			r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("nothing_to_claim"), err)
		default:
			r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		}
		return
	}
	r.Data = serializer.BuildVoucher(voucher, deviceInfo.Platform)
	return
}
