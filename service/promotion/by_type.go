package promotion

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"web-api/cache"
	"web-api/conf"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

const (
	birthdayBonusRewardCacheKey = "birthday_bonus_reward_cache_key:%d"
)

func RewardByType(c context.Context, p models.Promotion, s models.PromotionSession, userID, progress int64, now time.Time) (reward, meetGapType int64, vipIncrementDetail models.VipIncrementDetail, err error) {
	switch int64(p.Type) {
	case models.PromotionTypeVipBirthdayB:
		err := cache.RedisStore.Get(fmt.Sprintf(birthdayBonusRewardCacheKey, userID), &reward)
		if errors.Is(err, persist.ErrCacheMiss) {
			user, ok := c.Value("user").(model.User)
			if !ok {
				return 0, 0, models.VipIncrementDetail{}, fmt.Errorf("RewardByType get reward of promotion type %d fail as user not obtained", p.Type)
			}
			date, _ := time.Parse(time.DateOnly, user.Birthday)
			reward = getBirtdayReward(c, date, userID)
		}
	case models.PromotionTypeVipRebate, models.PromotionTypeVipPromotionB, models.PromotionTypeVipWeeklyB:
		r := getSameDayVipRewardRecord(model.DB.Debug(), userID, p.ID)
		reward = r.Amount
	case models.PromotionTypeVipReferral:
		reward = rewardVipReferral(c, userID, now)
	default:
		var vip models.VipRecord
		vip, err = model.GetVipWithDefault(c, userID)
		if err != nil {
			return
		}
		reward, meetGapType, vipIncrementDetail = p.GetRewardDetails().GetReward(progress, vip.VipRule.VIPLevel)
	}
	return
}

func ProgressByType(ctx context.Context, p models.Promotion, s models.PromotionSession, userID int64, now time.Time) (progress int64) {
	//ctx = contextify.AppendCtx(ctx, contextify.DefaultContextKey, fmt.Sprintf("ProgressByType() p.Type %d", p.Type))
	switch int64(p.Type) {
	// not necessary
	// case models.PromotionTypeVipReferral, models.PromotionTypeVipRebate:
	// 	//separate handling based on separate table
	case models.PromotionTypeVipWeeklyB:
		//may need to check deposit requirement + vip
		vip, _ := model.GetVipWithDefault(ctx, userID)
		progress = vip.VipRule.VIPLevel
		// not necessary
	// case models.PromotionTypeVipBirthdayB, models.PromotionTypeVipPromotionB:
	// 	vip, _ := model.GetVipWithDefault(c, userID)
	// 	progress = vip.VipRule.VIPLevel
	case models.PromotionTypeFirstDepB, models.PromotionTypeFirstDepIns:
		order, err := model.FirstTopup(ctx, userID)
		//ctx = contextify.AppendCtx(ctx, contextify.DefaultContextKey, fmt.Sprintf("model.FirstTopup(ctx, userID) = order %#v, err %v", order, err))
		if util.IsGormNotFound(err) {
			return
		} else if err != nil {
			return
		}
		progress = order.AppliedCashInAmount
		log.Printf("progress is %d, userid %d order %d", progress, userID, order)
	case models.PromotionTypeReDepB:
		orders, err := model.ScopedTopupExceptAllTimeFirst(ctx, userID, s.TopupStart, s.TopupEnd)
		if err != nil {
			return
		}
		progress = util.Reduce(orders, func(amount int64, input model.CashOrder) int64 {
			return amount + input.AppliedCashInAmount
		}, 0)
	case models.PromotionTypeReDepIns:
		orders, err := model.ScopedTopupExceptAllTimeFirst(ctx, userID, s.TopupStart, s.TopupEnd)
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
	switch int64(p.Type) {
	case models.PromotionTypeVipRebate, models.PromotionTypeVipPromotionB, models.PromotionTypeVipWeeklyB, models.PromotionTypeVipBirthdayB:
		v, err := model.VoucherGetByUniqueID(c, models.GenerateVoucherUniqueId(int64(p.Type), p.ID, s.ID, userID, 0, buildSuffixByType(c, p, userID)))
		if err == nil && v.ID != 0 {
			claim.HasClaimed = true
		}
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

func ClaimVoucherByType(c context.Context, p models.Promotion, s models.PromotionSession, v models.VoucherTemplate, userID, promotionRequestID int64, rewardAmount int64, now time.Time, meetGapType int64, vipIncrementDetail models.VipIncrementDetail) (voucher models.Voucher, err error) {
	voucher = CraftVoucherByType(c, p, s, v, rewardAmount, userID, promotionRequestID, now, meetGapType, vipIncrementDetail)
	lang := model.GetUserLang(userID)

	switch int64(p.Type) {
	case models.PromotionTypeFirstDepB, models.PromotionTypeReDepB, models.PromotionTypeBeginnerB, models.PromotionTypeOneTimeDepB:
		//add money and insert voucher
		// add cash order
		err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) error {
			wagerChange := voucher.WagerMultiplier * rewardAmount
			err = CreateCashOrder(tx, int64(p.Type), userID, rewardAmount, wagerChange, "", "")
			if err != nil {
				return err
			}
			err = tx.Create(&voucher).Error
			if err != nil {
				return err
			}
			if int64(p.Type) == models.PromotionTypeBeginnerB {
				err = model.CreateUserAchievement(userID, models.UserAchievementIdFirstAppLoginReward)
				if err != nil {
					return err
				}
			}
			if int64(p.Type) == models.PromotionTypeOneTimeDepB {
				err = model.CreateUserAchievement(userID, models.UserAchievementIdFirstDepositBonusReward)
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
		err = claimVoucherReferralVip(c, p, voucher, userID, now)
		if err == nil {
			common.SendNotification(userID, consts.Notification_Type_Referral_Alliance, conf.GetI18N(lang).T(common.NOTIFICATION_REFERRAL_ALLIANCE_TITLE), conf.GetI18N(lang).T(common.NOTIFICATION_REFERRAL_ALLIANCE))
		}
	case models.PromotionTypeVipBirthdayB:
		err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) error {
			wagerChange := voucher.WagerMultiplier * rewardAmount
			err = CreateCashOrder(tx, int64(p.Type), userID, rewardAmount, wagerChange, "", "")
			if err != nil {
				return err
			}
			err = tx.Create(&voucher).Error
			if err != nil {
				return err
			}
			return nil
		})
		if err == nil {
			common.SendNotification(userID, consts.Notification_Type_Birthday_Bonus, conf.GetI18N(lang).T(common.NOTIFICATION_BIRTHDAY_BONUS_SUCCESS_TITLE), conf.GetI18N(lang).T(common.NOTIFICATION_BIRTHDAY_BONUS_SUCCESS))
		}
	case models.PromotionTypeVipRebate, models.PromotionTypeVipPromotionB, models.PromotionTypeVipWeeklyB:
		err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) error {
			wagerChange := voucher.WagerMultiplier * rewardAmount
			err = CreateCashOrder(tx, int64(p.Type), userID, rewardAmount, wagerChange, "", "")
			if err != nil {
				return err
			}
			err = tx.Create(&voucher).Error
			if err != nil {
				return err
			}
			rcd := getSameDayVipRewardRecord(tx, userID, p.ID)
			err = tx.Model(&rcd).Update("status", 2).Error
			if err != nil {
				return err
			}
			return nil
		})
		if err == nil {
			var nType, title, text string
			switch int64(p.Type) {
			case models.PromotionTypeVipRebate:
				nType = consts.Notification_Type_Rebate
				title = conf.GetI18N(lang).T(common.NOTIFICATION_REBATE_TITLE)
				text = conf.GetI18N(lang).T(common.NOTIFICATION_REBATE)
			case models.PromotionTypeVipPromotionB:
				nType = consts.Notification_Type_Vip_Promotion_Bonus
				title = conf.GetI18N(lang).T(common.NOTIFICATION_VIP_PROMOTION_BONUS_TITLE)
				text = conf.GetI18N(lang).T(common.NOTIFICATION_VIP_PROMOTION_BONUS)
			case models.PromotionTypeVipWeeklyB:
				nType = consts.Notification_Type_Weekly_Bonus
				title = conf.GetI18N(lang).T(common.NOTIFICATION_WEEKLY_BONUS_TITLE)
				text = conf.GetI18N(lang).T(common.NOTIFICATION_WEEKLY_BONUS)
			}
			common.SendNotification(userID, nType, title, text)
		}
	case models.PromotionTypeFirstDepIns, models.PromotionTypeReDepIns:
		//insert voucher only
		err = model.DB.Create(&voucher).Error
	case models.PromotionTypeCustomTemplate:
		err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) error {
			wagerChange := voucher.WagerMultiplier * rewardAmount
			err = CreateCashOrder(tx, int64(p.Type), userID, rewardAmount, wagerChange, "", p.Name)
			if err != nil {
				return err
			}
			err = tx.Create(&voucher).Error
			if err != nil {
				return err
			}
			return nil
		})
	}
	return
}

func ExtraByType(c context.Context, p models.Promotion, s models.PromotionSession, userID, progress int64, now time.Time) (extra any) {
	switch int64(p.Type) {
	case models.PromotionTypeVipReferral:
		// Get next session's reward - actually just records that cannot be claimed yet
		summaries, err := model.GetReferralAllianceSummaries(model.GetReferralAllianceSummaryCond{
			ReferrerIds:     []int64{userID},
			HasBeenClaimed:  []bool{false},
			CanClaimAfterGt: sql.NullTime{Time: now, Valid: true},
		})
		if err != nil {
			util.GetLoggerEntry(c).Error("GetReferralAllianceSummaries error", err)
			return nil
		}

		referralExtra := struct {
			NextSessionReward float64 `json:"next_session_reward"`
		}{}
		if len(summaries) == 0 {
			referralExtra.NextSessionReward = 0
		} else {
			referralExtra.NextSessionReward = float64(util.Max(summaries[0].ClaimableReward, 0)) / 100
		}
		extra = referralExtra
	}

	return
}

func CreateCashOrder(tx *gorm.DB, promoType, userId, rewardAmount, wagerChange int64, notes, name string) error {
	txType := promotionTxTypeMapping[promoType]
	sum, err := model.UserSum{}.UpdateUserSumWithDB(tx,
		userId,
		rewardAmount,
		wagerChange,
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
		WagerChange:           wagerChange,
		Name:                  name,
	}
	err = tx.Create(&dummyOrder).Error
	if err != nil {
		return err
	}
	common.SendUserSumSocketMsg(userId, sum.UserSum, "promotion", float64(rewardAmount)/100)
	return nil
}

func CraftVoucherByType(c context.Context, p models.Promotion, s models.PromotionSession, v models.VoucherTemplate, rewardAmount, userID int64, promotionRequestID int64, now time.Time, meetGapType int64, vipIncrementDetail models.VipIncrementDetail) (voucher models.Voucher) {
	endAt := earlier(v.EndAt, v.GetExpiryTimeStamp(now, p.Timezone))
	status := models.VoucherStatusReady
	isUsable := false
	suffix := buildSuffixByType(c, p, userID)
	switch int64(p.Type) {
	case models.PromotionTypeFirstDepB, models.PromotionTypeReDepB, models.PromotionTypeBeginnerB, models.PromotionTypeVipReferral, models.PromotionTypeVipRebate, models.PromotionTypeVipPromotionB, models.PromotionTypeVipWeeklyB, models.PromotionTypeVipBirthdayB, models.PromotionTypeCustomTemplate:
		status = models.VoucherStatusRedeemed
	case models.PromotionTypeFirstDepIns, models.PromotionTypeReDepIns:
		isUsable = true
	}

	voucherName := model.AmountReplace(v.Name, float64(rewardAmount)/100)
	wagerMultiplier := v.WagerMultiplier
	switch int64(p.Type) {
	case models.PromotionTypeVipRebate:
		vip, _ := model.GetVipWithDefault(c, userID)
		wagerMultiplier = vip.VipRule.BirthdayBenefitWagerMultiplier
	case models.PromotionTypeVipPromotionB:
		vip, _ := model.GetVipWithDefault(c, userID)
		wagerMultiplier = vip.VipRule.PromotionBenefitWagerMultiplier
	case models.PromotionTypeVipWeeklyB:
		vip, _ := model.GetVipWithDefault(c, userID)
		wagerMultiplier = vip.VipRule.WeeklyBenefitWagerMultiplier
	case models.PromotionTypeVipBirthdayB:
		vip, _ := model.GetVipWithDefault(c, userID)
		wagerMultiplier = vip.VipRule.BirthdayBenefitWagerMultiplier
	case models.PromotionTypeCustomTemplate:
		voucherName = p.Name
	}

	voucher = models.Voucher{

		UserID:            userID,
		Status:            status,
		StartAt:           now,
		EndAt:             endAt,
		VoucherTemplateID: v.ID,
		BrandID:           p.BrandId,
		Amount:            rewardAmount,
		// TransactionDetails
		Name:               voucherName,
		Description:        v.Description,
		PromotionType:      v.PromotionType,
		PromotionID:        p.ID,
		UsageDetails:       v.UsageDetails,
		Image:              v.Image,
		WagerMultiplier:    wagerMultiplier,
		PromotionSessionID: s.ID,
		IsUsable:           isUsable,
		// ReferenceType
		// ReferenceID
		// TransactionID
		MeetGapType:         meetGapType,
		IncludeVipIncrement: vipIncrementDetail.IncludeVipIncrement,
		UserVipLevel:        vipIncrementDetail.UserVipLevel,
		VipIncrementType:    vipIncrementDetail.VipIncrementType,
		VipIncrementValue:   vipIncrementDetail.VipIncrementValue,
		VipIncrementAmount:  vipIncrementDetail.VipIncrementAmount,
		PromotionRequestID:  promotionRequestID,
	}
	voucher.FillUniqueID(suffix)
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

func rewardVipReferral(c context.Context, userID int64, now time.Time) (reward int64) {
	summaries, err := model.GetReferralAllianceSummaries(model.GetReferralAllianceSummaryCond{
		ReferrerIds:      []int64{userID},
		HasBeenClaimed:   []bool{false},
		CanClaimAfterLte: sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		util.GetLoggerEntry(c).Error("GetReferralAllianceSummaries error", err)
		return
	}

	if len(summaries) == 0 {
		return 0
	}

	return util.Max(summaries[0].ClaimableReward, 0) // return 0 if negative
}

func claimVoucherReferralVip(c context.Context, p models.Promotion, voucher models.Voucher, userID int64, now time.Time) error {
	user := c.Value("user").(model.User)
	return model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) error {
		rewardRecords, err := claimReferralAllianceRewards(tx, userID, now)
		if err != nil {
			return fmt.Errorf("failed to claim rewards: %w", err)
		}

		var totalClaimable int64
		for _, r := range rewardRecords {
			totalClaimable += r.ClaimableAmount
		}

		var rewardRecordIds []int64
		for _, r := range rewardRecords {
			rewardRecordIds = append(rewardRecordIds, r.ID)
		}
		cashOrderNotes := util.JSON(map[string]any{
			"reward_record_ids": rewardRecordIds,
		})

		wagerChange := user.ReferralWagerMultiplier * totalClaimable
		err = CreateCashOrder(tx, int64(p.Type), user.ID, totalClaimable, wagerChange, cashOrderNotes, "")
		if err != nil {
			return fmt.Errorf("failed to create cash order: %w", err)
		}

		err = tx.Create(&voucher).Error
		if err != nil {
			return fmt.Errorf("failed to create voucher: %w", err)
		}

		return nil
	})
}

func claimReferralAllianceRewards(tx *gorm.DB, referrerId int64, now time.Time) ([]models.ReferralAllianceReward, error) {
	// Get reward records
	cond := model.GetReferralAllianceRewardsCond{
		ReferrerIds:      []int64{referrerId},
		HasBeenClaimed:   []bool{false},
		CanClaimAfterLte: sql.NullTime{Time: now, Valid: true},
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

func earlier(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func getSameDayVipRewardRecord(tx *gorm.DB, userID, prmotionID int64) models.VipRewardRecords {
	r, _ := model.GetVipRewardRecord(tx, userID, prmotionID, Today0am().UTC())
	return r
}

func getBirtdayReward(c context.Context, date time.Time, userID int64) (reward int64) {
	today := Today0am()
	defer func() {
		cache.RedisStore.Set(fmt.Sprintf(birthdayBonusRewardCacheKey, userID), reward, time.Until(today.Add(24*time.Hour)))
	}()
	if date.Month() != today.Month() || date.Day() != today.Day() {
		return
	}
	user := c.Value("user").(model.User)
	v := os.Getenv("VIP_BIRTHDAY_REWARD_MATURE_DAYS")
	days, _ := strconv.Atoi(v)
	if user.CreatedAt.After(Today0am().Add(-time.Duration(days) * 24 * time.Hour)) {
		return
	}
	vip, _ := model.GetVipWithDefault(c, userID)
	reward = vip.VipRule.BirthdayBenefit
	return
}

func buildSuffixByType(c context.Context, p models.Promotion, userID int64) string {
	today := Today0am()
	suffix := ""
	vip, _ := model.GetVipWithDefault(c, userID)
	switch int64(p.Type) {
	case models.PromotionTypeVipRebate:
		suffix = fmt.Sprintf("date-%s", today.Format(time.DateOnly))
	case models.PromotionTypeVipWeeklyB:
	case models.PromotionTypeVipBirthdayB:
		suffix = fmt.Sprintf("year-%d", today.Year())
	case models.PromotionTypeVipPromotionB:
		suffix = fmt.Sprintf("vip-%d", vip.VipRule.VIPLevel)
	}
	return suffix
}
