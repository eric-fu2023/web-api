package promotion

import (
	"context"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

func ProgressByType(c context.Context, p model.Promotion, s model.PromotionSession, userID int64, now time.Time) (progress int64) {
	switch p.Type {
	case models.PromotionTypeFirstDepB, models.PromotionTypeFirstDepIns:
		order, err := model.FirstTopup(c, userID)
		if util.IsGormNotFound(err) {
			return
		} else if err != nil {
			return
		}
		progress = order.AppliedCashInAmount
	case models.PromotionTypeReDepB:
		orders, err := model.ScopedTopupExceptAllTimeFirst(c, userID, s.TopupStart, s.TopupEnd)
		if err != nil {
			return
		}
		progress = util.Reduce(orders, func(amount int64, input model.CashOrder) int64 {
			return amount + input.AppliedCashInAmount
		}, 0)
	case models.PromotionTypeReDepIns:
		orders, err := model.ScopedTopupExceptAllTimeFirst(c, userID, s.TopupStart, s.TopupEnd)
		if err != nil {
			return
		}
		for _, o := range orders {
			if progress < o.AppliedCashInAmount {
				progress = o.AppliedCashInAmount
			}
		}
	}
	return
}

func ClaimStatusByType(c context.Context, p model.Promotion, s model.PromotionSession, userID int64, now time.Time) (claim serializer.ClaimStatus) {
	claim.ClaimStart = s.ClaimStart.Unix()
	claim.ClaimEnd = s.ClaimEnd.Unix()
	switch p.Type {
	case models.PromotionTypeFirstDepB, models.PromotionTypeFirstDepIns:
		_, err := model.VoucherGetByUserSession(c, userID, s.ID)
		if err == nil {
			claim.HasClaimed = true
		} else {
			order, err := model.FirstTopup(c, userID)
			if err == nil {
				claim.ClaimEnd = order.CreatedAt.Add(7 * 24 * time.Hour).Unix()
			}
		}
	default:
		_, err := model.VoucherGetByUserSession(c, userID, s.ID)
		if err == nil {
			claim.HasClaimed = true
		}
	}
	return
}

func ClaimVoucherByType(c context.Context, p model.Promotion, s model.PromotionSession, v model.VoucherTemplate, rewardAmount, userID int64, now time.Time) (voucher model.Voucher, err error) {
	voucher = CraftVoucherByType(c, p, s, v, rewardAmount, userID, now)
	switch p.Type {
	case models.PromotionTypeFirstDepB, models.PromotionTypeReDepB:
		//add money and insert voucher
	case models.PromotionTypeFirstDepIns, models.PromotionTypeReDepIns:
		//insert voucher only
		err = model.DB.Create(&voucher).Error
	}
	return
}

func CraftVoucherByType(c context.Context, p model.Promotion, s model.PromotionSession, v model.VoucherTemplate, rewardAmount, userID int64, now time.Time) (voucher model.Voucher) {
	endAt := earlier(v.EndAt, now.Add(time.Duration(v.ExpiryValue)*time.Second))
	status := models.VoucherStatusReady
	isUsable := false
	switch p.Type {
	case models.PromotionTypeFirstDepB, models.PromotionTypeReDepB:
		status = models.VoucherStatusRedeemed
	case models.PromotionTypeFirstDepIns, models.PromotionTypeReDepIns:
		isUsable = true
	}

	voucher = model.Voucher{
		Voucher: models.Voucher{
			UserID:            userID,
			Status:            status,
			StartAt:           now,
			EndAt:             endAt,
			VoucherTemplateID: v.ID,
			BrandID:           p.BrandID,
			Amount:            rewardAmount,
			// TransactionDetails
			Name:               model.AmountReplace(v.Name, rewardAmount),
			Description:        v.Description,
			Type:               v.Type,
			PromotionID:        p.ID,
			UsageDetails:       v.UsageDetails,
			Image:              v.Image,
			WagerMultiplier:    v.WagerMultiplier,
			PromotionSessionID: s.ID,
			IsUsable:           isUsable,
			// ReferenceType
			// ReferenceID
			// TransactionID
		}}
	return
}

func earlier(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func later(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return b
	}
	return a
}
