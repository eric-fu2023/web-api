package promotion

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

func RewardByType(c context.Context, p models.Promotion, s models.PromotionSession, userID, progress int64, now time.Time) (reward int64) {
	switch p.Type {
	case models.PromotionTypeVipReferral:
		oneDayBefore, err := getOneDayBeforeDateString(now)
		if err != nil {
			util.GetLoggerEntry(c).Error("getOneDayBeforeDateString error", err)
			return
		}

		summaries, err := model.GetReferralAllianceSummaries(model.GetReferralAllianceSummaryCond{
			ReferrerIds:    []int64{userID},
			HasBeenClaimed: []bool{false},
			BetDateEnd:     oneDayBefore,
		})
		if err != nil {
			util.GetLoggerEntry(c).Error("GetReferralAllianceSummaries error", err)
			return
		}

		if len(summaries) == 0 {
			return 0
		}

		vipRecord, err := model.GetVipWithDefault(c, userID)
		if err != nil {
			return
		}
		rewardCap := vipRecord.VipRule.ReferralCap

		if summaries[0].TotalReward > rewardCap {
			return rewardCap
		}
		return summaries[0].TotalReward

	default:
		reward = p.GetRewardDetails().GetReward(progress)
	}
	return
}

func ProgressByType(c context.Context, p models.Promotion, s models.PromotionSession, userID int64, now time.Time) (progress int64) {
	switch p.Type {
	// not necessary
	// case models.PromotionTypeVipReferral, models.PromotionTypeVipRebate:
	// 	//separate handling based on separate table
	case models.PromotionTypeVipWeeklyB:
		//may need to check deposit requirement + vip
		vip, _ := model.GetVipWithDefault(c, userID)
		progress = vip.VipRule.VIPLevel
		// not necessary
	// case models.PromotionTypeVipBirthdayB, models.PromotionTypeVipPromotionB:
	// 	vip, _ := model.GetVipWithDefault(c, userID)
	// 	progress = vip.VipRule.VIPLevel
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
	case models.PromotionTypeFirstDepB, models.PromotionTypeReDepB, models.PromotionTypeBeginnerB, models.PromotionTypeOneTimeDepB:
		//add money and insert voucher
		// add cash order
		err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) error {
			err = CreateCashOrder(tx, voucher, p.Type, userID, rewardAmount, "")
			if err != nil {
				return err
			}
			err = tx.Create(&voucher).Error
			if err != nil {
				return err
			}
			if p.Type == models.PromotionTypeBeginnerB {
				err = model.CreateUserAchievement(userID, model.UserAchievementIdFirstAppLoginReward)
				if err != nil {
					return err
				}
			}
			if p.Type == models.PromotionTypeOneTimeDepB {
				err = model.CreateUserAchievement(userID, model.UserAchievementIdFirstDepositBonusReward)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err == nil {
			common.SendCashNotificationWithoutCurrencyId(userID, consts.Notification_Type_Deposit_Bonus, common.NOTIFICATION_DEPOSIT_BONUS_SUCCESS_TITLE, common.NOTIFICATION_DEPOSIT_BONUS_SUCCESS, rewardAmount)
		}
	case models.PromotionTypeVipReferral:
		err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) error {
			rewardRecords, err := ClaimReferralAllianceRewards(tx, userID, now)
			if err != nil {
				return fmt.Errorf("failed to claim rewards: %w", err)
			}

			vipRecord, err := model.GetVipWithDefault(c, userID)
			if err != nil {
				return fmt.Errorf("failed to get vip record: %w", err)
			}
			rewardCap := vipRecord.VipRule.ReferralCap

			var totalReward int64
			for _, r := range rewardRecords {
				totalReward += r.Amount
			}
			if totalReward > rewardCap {
				totalReward = rewardCap
			}

			var rewardRecordIds []int64
			for _, r := range rewardRecords {
				rewardRecordIds = append(rewardRecordIds, r.ID)
			}
			cashOrderNotes := util.JSON(map[string]any{
				"reward_record_ids": rewardRecordIds,
			})

			err = CreateCashOrder(tx, voucher, p.Type, userID, totalReward, cashOrderNotes)
			if err != nil {
				return fmt.Errorf("failed to create cash order: %w", err)
			}

			err = tx.Create(&voucher).Error
			if err != nil {
				return fmt.Errorf("failed to create voucher: %w", err)
			}

			return nil
		})
		if err == nil {
			common.SendCashNotificationWithoutCurrencyId(userID, consts.Notification_Type_Deposit_Bonus, common.NOTIFICATION_DEPOSIT_BONUS_SUCCESS_TITLE, common.NOTIFICATION_DEPOSIT_BONUS_SUCCESS, rewardAmount)
		}
	case models.PromotionTypeFirstDepIns, models.PromotionTypeReDepIns:
		//insert voucher only
		err = model.DB.Create(&voucher).Error
	}
	return
}

func ClaimReferralAllianceRewards(tx *gorm.DB, referrerId int64, now time.Time) ([]models.ReferralAllianceReward, error) {
	oneDayBefore, err := getOneDayBeforeDateString(now)
	if err != nil {
		return nil, fmt.Errorf("failed to get one day before date string: %w", err)
	}

	// Get reward records
	cond := model.GetReferralAllianceRewardsCond{
		ReferrerIds:    []int64{referrerId},
		HasBeenClaimed: []bool{false},
		BetDateEnd:     oneDayBefore,
	}
	rewardRecords, err := model.GetReferralAllianceRewards(cond)
	if err != nil {
		return nil, fmt.Errorf("failed to get reward records: %w", err)
	}

	var ids []int64
	for _, r := range rewardRecords {
		ids = append(ids, r.ID)
	}

	err = model.ClaimReferralAllianceRewards(tx, ids, now)
	if err != nil {
		return nil, fmt.Errorf("failed to claim rewards: %w", err)
	}

	return rewardRecords, nil
}

func CreateCashOrder(tx *gorm.DB, voucher models.Voucher, promoType, userId, rewardAmount int64, notes string) error {
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
	if notes == "" {
		notes = "dummy"
	}
	orderType := promotionOrderTypeMapping[promoType]
	dummyOrder := models.CashOrder{
		ID:                    uuid.NewString(),
		UserId:                userId,
		OrderType:             orderType,
		Status:                models.CashOrderStatusSuccess,
		Notes:                 models.EncryptedStr(notes),
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
	common.SendUserSumSocketMsg(userId, sum.UserSum)
	return nil
}

func CraftVoucherByType(c context.Context, p models.Promotion, s models.PromotionSession, v models.VoucherTemplate, rewardAmount, userID int64, now time.Time) (voucher models.Voucher) {
	endAt := earlier(v.EndAt, v.GetExpiryTimeStamp(now, p.Timezone))
	status := models.VoucherStatusReady
	isUsable := false
	switch p.Type {
	case models.PromotionTypeFirstDepB, models.PromotionTypeReDepB, models.PromotionTypeBeginnerB, models.PromotionTypeVipReferral:
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
		Name:               model.AmountReplace(v.Name, float64(rewardAmount)/100),
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

func ValidateVoucherUsageByType(v models.Voucher, oddsFormat, matchType int, odds float64, betAmount int64, isParley bool) (ret bool) {
	ret = false
	switch v.PromotionType {
	case models.PromotionTypeFirstDepIns, models.PromotionTypeReDepIns:
		if isParley {
			return
		}
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

func getOneDayBeforeDateString(now time.Time) (string, error) {
	tzOffsetStr, err := model.GetAppConfigWithCache("timezone", "offset_seconds")
	if err != nil {
		return "", fmt.Errorf("failed to get tz offset config: %w", err)
	}
	tzOffset, err := strconv.Atoi(tzOffsetStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse tz offset config: %w", err)
	}
	return now.In(time.FixedZone("", tzOffset)).AddDate(0, 0, -1).Format(time.DateOnly), nil
}
