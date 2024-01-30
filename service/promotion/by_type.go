package promotion

import (
	"context"
	"fmt"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

func ProgressByType(c context.Context, p models.Promotion, s models.PromotionSession, userID int64, now time.Time) (progress int64) {
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

func ClaimStatusByType(c context.Context, p models.Promotion, s models.PromotionSession, userID int64, now time.Time) (claim serializer.ClaimStatus) {
	claim.ClaimStart = s.ClaimStart.Unix()
	claim.ClaimEnd = s.ClaimEnd.Unix()
	switch p.Type {
	case models.PromotionTypeFirstDepB, models.PromotionTypeFirstDepIns:
		v, err := model.VoucherGetByUserSession(c, userID, s.ID)
		if err == nil && v.ID != 0 {
			claim.HasClaimed = true
		} else {
			order, err := model.FirstTopup(c, userID)
			if err == nil {
				claim.ClaimEnd = order.CreatedAt.Add(7 * 24 * time.Hour).Unix()
			}
		}
	default:
		v, err := model.VoucherGetByUserSession(c, userID, s.ID)
		if err == nil && v.ID != 0 {
			claim.HasClaimed = true
		}
	}
	return
}

func ClaimVoucherByType(c context.Context, p models.Promotion, s models.PromotionSession, v models.VoucherTemplate, rewardAmount, userID int64, now time.Time) (voucher models.Voucher, err error) {
	voucher = CraftVoucherByType(c, p, s, v, rewardAmount, userID, now)
	switch p.Type {
	case models.PromotionTypeFirstDepB, models.PromotionTypeReDepB, models.PromotionTypeBeginnerB:
		//add money and insert voucher
		// add cash order
		err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) error {
			err = CreateCashOrder(tx, voucher, p.Type, userID, rewardAmount)
			if err != nil {
				return err
			}
			err = tx.Create(&voucher).Error
			if err != nil {
				return err
			}
			return nil
		})
	case models.PromotionTypeFirstDepIns, models.PromotionTypeReDepIns:
		//insert voucher only
		err = model.DB.Create(&voucher).Error
	}
	return
}

func CreateCashOrder(tx *gorm.DB, voucher models.Voucher, promoType, userId, rewardAmount int64) error {
	txType := promotionTxTypeMapping[promoType]
	sum, err := model.UserSum{}.UpdateUserSumWithDB(tx,
		userId,
		rewardAmount,
		voucher.WagerMultiplier*rewardAmount,
		0,
		txType,
		"")
	if err != nil {
		return err
	}
	orderType := promotionOrderTypeMapping[promoType]
	dummyOrder := models.CashOrder{
		ID:                    uuid.NewString(),
		UserId:                userId,
		OrderType:             orderType,
		Status:                models.CashOrderStatusSuccess,
		Notes:                 "dummy",
		AppliedCashInAmount:   rewardAmount,
		ActualCashInAmount:    rewardAmount,
		EffectiveCashInAmount: rewardAmount,
		BalanceBefore:         sum.Balance - rewardAmount,
		WagerChange:           voucher.WagerMultiplier * rewardAmount,
	}
	err = tx.Create(&dummyOrder).Error
	if err != nil {
		return err
	}
	return nil
}

func CraftVoucherByType(c context.Context, p models.Promotion, s models.PromotionSession, v models.VoucherTemplate, rewardAmount, userID int64, now time.Time) (voucher models.Voucher) {
	endAt := earlier(v.EndAt, now.Add(time.Duration(v.ExpiryValue)*time.Second))
	status := models.VoucherStatusReady
	isUsable := false
	switch p.Type {
	case models.PromotionTypeFirstDepB, models.PromotionTypeReDepB, models.PromotionTypeBeginnerB:
		status = models.VoucherStatusRedeemed
	case models.PromotionTypeFirstDepIns, models.PromotionTypeReDepIns:
		isUsable = true
	}

	voucher = models.Voucher{

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
		PromotionType:      v.PromotionType,
		PromotionID:        p.ID,
		UsageDetails:       v.UsageDetails,
		Image:              v.Image,
		WagerMultiplier:    v.WagerMultiplier,
		PromotionSessionID: s.ID,
		IsUsable:           isUsable,
		// ReferenceType
		// ReferenceID
		// TransactionID
	}
	return
}

func ValidateVoucherUsageByType(v models.Voucher, oddsFormat, matchType int, odds float64, betAmount int64) (ret bool) {
	ret = false
	switch v.PromotionType {
	case models.PromotionTypeFirstDepIns, models.PromotionTypeReDepIns:
		if matchType != MatchTypeNotStarted {
			return
		}
		if oddsFormat != OddsFormatEU {
			return
		}
		if !ValidateUsageDetailsByType(v, matchType, odds, betAmount) {
			return
		}
		ret = true
	}
	return
}

func ValidateUsageDetailsByType(v models.Voucher, matchType int, odds float64, betAmount int64) (ret bool) {
	ret = false
	switch v.PromotionType {
	case models.PromotionTypeFirstDepIns, models.PromotionTypeReDepIns:
		u := v.GetUsageDetails()
		ret = false
		// for _, c := range u.MatchType {
		// 	if !c.Condition(fmt.Sprintf("%d", matchType), "number") {
		// 		return
		// 	}
		// }
		for _, c := range u.Odds {
			if !c.Condition(fmt.Sprintf("%f", odds), "number") {
				return
			}
		}
		// for _, c := range u.BetAmount {
		// 	if !c.Condition(fmt.Sprintf("%d", betAmount), "number") {
		// 		return
		// 	}
		// }
		if betAmount < v.Amount {
			return
		}
		ret = true
	}
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
