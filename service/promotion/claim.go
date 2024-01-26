package promotion

import (
	"errors"
	"fmt"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"

	"github.com/gin-gonic/gin"
	"github.com/go-redsync/redsync/v4"
)

const (
	userPromotionSessionClaimKey = "user_promotion_session_claim_lock:%d:%d"
)

type PromotionClaim struct {
	ID int64 `form:"id" json:"id"`
}

func (p PromotionClaim) Handle(c *gin.Context) (r serializer.Response, err error) {

	now := time.Now()
	brand := c.MustGet(`_brand`).(int)
	user := c.MustGet("user").(model.User)
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
		template    model.VoucherTemplate
	)
	mutex := cache.RedisLockClient.NewMutex(fmt.Sprintf(userPromotionSessionClaimKey, user.ID, session.ID), redsync.WithExpiry(5*time.Second))
	mutex.Lock()
	defer mutex.Unlock()
	claimStatus = ClaimStatusByType(c, promotion, session, user.ID, now)
	if claimStatus.HasClaimed {
		err = errors.New("double_claim")
		r = serializer.Err(c, p, serializer.CodeGeneralError, "Already Claimed", err)
		return
	}
	if time.Unix(claimStatus.ClaimEnd, 0).Before(now) || time.Unix(claimStatus.ClaimStart, 0).Before(now) {
		err = errors.New("unavailable_for_now")
		r = serializer.Err(c, p, serializer.CodeGeneralError, "Unavailable for now", err)
		return
	}
	progress = ProgressByType(c, promotion, session, user.ID, now)
	reward = promotion.GetRewardDetails().GetReward(progress)
	template, err = model.VoucherTemplateGetByPromotion(c, p.ID)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	voucher, err := ClaimVoucherByType(c, promotion, session, template, reward, user.ID, now)
	if err != nil {
		r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	r.Data = serializer.BuildVoucher(voucher)
	return
}
