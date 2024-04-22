package promotion

import (
	"context"
	"errors"
	"fmt"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"

	models "blgit.rfdev.tech/taya/ploutos-object"
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
	deviceInfo, _ := util.GetDeviceInfo(c)
	i18n := c.MustGet("i18n").(i18n.I18n)

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
	voucher, err := Claim(c, now, promotion, session, user)
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

func Claim(c context.Context, now time.Time, promotion models.Promotion, session models.PromotionSession, user model.User) (voucher models.Voucher, err error) {
	mutex := cache.RedisLockClient.NewMutex(fmt.Sprintf(userPromotionSessionClaimKey, user.ID, session.ID), redsync.WithExpiry(5*time.Second))
	mutex.Lock()
	defer mutex.Unlock()
	var (
		progress    int64
		reward      int64
		claimStatus serializer.ClaimStatus
		template    models.VoucherTemplate
	)
	claimStatus = ClaimStatusByType(c, promotion, session, user.ID, now)
	if claimStatus.HasClaimed {
		err = errors.New("double_claim")
		// r = serializer.Err(c, p, serializer.CodeGeneralError, "Already Claimed", err)
		return
	}
	if time.Unix(claimStatus.ClaimEnd, 0).Before(now) || time.Unix(claimStatus.ClaimStart, 0).After(now) {
		err = errors.New("unavailable_for_now")
		// r = serializer.Err(c, p, serializer.CodeGeneralError, "Unavailable for now", err)
		return
	}
	progress = ProgressByType(c, promotion, session, user.ID, now)
	reward = RewardByType(c, promotion, session, user.ID, progress, now)
	template, err = model.VoucherTemplateGetByPromotion(c, promotion.ID)
	if err != nil {
		// r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	if reward == 0 {
		err = errors.New("nothing_to_claim")
		// r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("nothing_to_claim"), err)
		return
	}
	voucher, err = ClaimVoucherByType(c, promotion, session, template, user, reward, now)
	if err != nil {
		// r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	// r.Data = serializer.BuildVoucher(voucher, deviceInfo.Platform)
	return
}
