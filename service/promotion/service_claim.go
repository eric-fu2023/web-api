package promotion

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"web-api/cache"
	"web-api/model"
	"web-api/serializer"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/go-redsync/redsync/v4"
)

func Claim(ctx context.Context, now time.Time, promotion ploutos.Promotion, session ploutos.PromotionSession, userID int64, user *model.User) (voucher ploutos.Voucher, err error) {
	mutex := cache.RedisLockClient.NewMutex(fmt.Sprintf(userPromotionSessionClaimKey, userID, session.ID), redsync.WithExpiry(5*time.Second))
	mutex.Lock()
	defer mutex.Unlock()
	var (
		progress        int64
		reward          int64
		claimStatus     serializer.ClaimStatus
		voucherTemplate ploutos.VoucherTemplate
	)

	ctx = rfcontext.AppendParams(rfcontext.AppendCallDesc(ctx, "Claim"), "Claim", map[string]interface{}{
		"promotion.Id":        promotion.ID,
		"promotion.Name":      promotion.Name,
		"promotion.Type.Name": ploutos.DefaultPromotionTypeNames[promotion.Type],
		"userID":              userID,
		"user":                user,
		"session.ID":          session.ID,
		"session.ClaimStart":  session.ClaimStart,
		"session.ClaimEnd":    session.ClaimEnd,
	})

	claimStatus = GetPromotionSessionClaimStatus(ctx, promotion, session, userID, now)
	if claimStatus.HasClaimed {
		err = errors.New("double_claim")
		ctx = rfcontext.AppendError(ctx, err, "GetPromotionSessionClaimStatus")
		log.Println(rfcontext.Fmt(ctx))
		// r = serializer.Err(c, p, serializer.CodeGeneralError, "Already Claimed", err)
		return
	}
	if time.Unix(claimStatus.ClaimEnd, 0).Before(now) || time.Unix(claimStatus.ClaimStart, 0).After(now) {
		err = errors.New("unavailable_for_now")
		// r = serializer.Err(c, p, serializer.CodeGeneralError, "Unavailable for now", err)
		return
	}
	fmt.Println("promotion.GetPromotionSessionProgress ")
	progress, err = GetPromotionSessionProgress(ctx, promotion, session, userID)
	// FIXME
	// to remove error suppression
	// if !errors.Is(err, ErrPromotionSessionUnknownPromotionType) {
	// 	return voucher, err
	// }
	fmt.Println("promotion.GetPromotionRewards ")
	reward, meetGapType, vipIncrementDetail, err := GetPromotionRewards(ctx, promotion, userID, progress, now, user)
	if err != nil {
		// r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	fmt.Println("promotion.GetPromotionVoucherTemplateByPromotionId ")
	voucherTemplate, err = model.GetPromotionVoucherTemplateByPromotionId(ctx, promotion.ID)
	if err != nil {
		// r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	if reward == 0 {
		err = errors.New("nothing_to_claim")
		// r = serializer.Err(c, p, serializer.CodeGeneralError, i18n.T("nothing_to_claim"), err)
		return
	}

	fmt.Println("promotion.ClaimVoucherByType ")
	voucher, err = ClaimVoucherByType(ctx, promotion, session, voucherTemplate, userID, 0, reward, now, meetGapType, vipIncrementDetail)
	if err != nil {
		// r = serializer.Err(c, p, serializer.CodeGeneralError, "", err)
		return
	}
	// r.Data = serializer.BuildVoucher(voucher, deviceInfo.Platform)
	return
}
