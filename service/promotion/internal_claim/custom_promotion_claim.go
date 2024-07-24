package internalclaim

import (
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/promotion"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type CustomPromotionClaimRequest struct {
	PromotionRequestID int64 `json:"promotion_request_id"`
	UserID             int64 `json:"user_id"`      //needed
	PromotionID        int64 `json:"promotion_id"` //craft voucher
	Amount             int64 `json:"amount"`
	WagerMultiplier    int64 `json:"wager_multiplier"`
}

// get voucher template
// craft voucher
// claim promotion - check unique id
// mark vip reward record

func (p CustomPromotionClaimRequest) Handle(c *gin.Context) (r serializer.Response, err error) {
	now := time.Now().UTC()

	promo, err := model.GetPromotion(c, p.PromotionID)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}

	voucherTemplate := models.VoucherTemplate{}
	voucherTemplate.WagerMultiplier = p.WagerMultiplier
	voucherTemplate.PromotionType = models.PromotionTypeCustomTemplate
	// promo.Type = 12
	voucher, err := promotion.ClaimVoucherByType(c, promo, models.PromotionSession{}, voucherTemplate, p.UserID, p.PromotionRequestID, p.Amount, now, 0, models.VipIncrementDetail{})
	if err != nil {
		util.GetLoggerEntry(c).Error(err)
	}
	util.GetLoggerEntry(c).Infof("Generated voucher: %#v", voucher)
	r.Data = voucher.ID
	return
}
