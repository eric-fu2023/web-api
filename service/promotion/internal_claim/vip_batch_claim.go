package internalclaim

import (
	"time"

	"web-api/model"
	"web-api/serializer"
	"web-api/service/promotion"
	"web-api/util"

	"github.com/gin-gonic/gin"
)

type VipBatchClaimRequest struct {
	UserIDList  []int64 `json:"user_id_list"` //needed
	PromotionID int64   `json:"promotion_id"` //craft voucher
	After       int64   `json:"after"`
}

// get voucher template
// craft voucher
// claim promotion - check unique id
// mark vip reward record

func (p VipBatchClaimRequest) Handle(c *gin.Context) (r serializer.Response, err error) {
	now := time.Now().UTC()

	promo, err := model.PromotionGetActiveNoBrand(c, p.PromotionID, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	// tz := time.FixedZone("local", int(promotion.Timezone))
	// now = now.In(tz)
	session, err := model.GetActivePromotionSessionByPromotionId(c, p.PromotionID, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	for _, uid := range p.UserIDList {
		voucher, err := promotion.Claim(c, now, promo, session, uid, nil)
		if err != nil {
			util.GetLoggerEntry(c).Error(err)
		}
		util.GetLoggerEntry(c).Infof("Generated voucher: %#v", voucher)
	}
	r.Data = "success"
	return
}
