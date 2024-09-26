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

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

const (
	birthdayBonusRewardCacheKey = "birthday_bonus_reward_cache_key:%d"
)

func GetPromotionRewards(c context.Context, p ploutos.Promotion, userID, progress int64, now time.Time, user *model.User) (reward, vipMeetGapType int64, vipIncrementDetail ploutos.VipIncrementDetail, err error) {
	switch p.Type {
	case ploutos.PromotionTypeVipBirthdayB:
		rErr := cache.RedisStore.Get(fmt.Sprintf(birthdayBonusRewardCacheKey, userID), &reward)
		if errors.Is(rErr, persist.ErrCacheMiss) {
			if user == nil {
				err = fmt.Errorf("getting rewards for PromotionTypeVipBirthdayB userid: %d, promotion %#v ", userID, p)
			} else {
				date, _ := time.Parse(time.DateOnly, user.Birthday)
				reward = getBirtdayReward(c, date, userID)
			}
		}
		err = rErr
	case ploutos.PromotionTypeVipRebate, ploutos.PromotionTypeVipPromotionB, ploutos.PromotionTypeVipWeeklyB:
		r := getSameDayVipRewardRecord(model.DB.Debug(), userID, p.ID)
		reward = r.Amount
	case ploutos.PromotionTypeVipReferral:
		reward = rewardVipReferral(c, userID, now)
	case ploutos.PromotionTypeSpinWheel:
		var vip ploutos.VipRecord
		vip, err = model.GetVipWithDefault(c, userID)
		if err != nil {
			log.Printf("GetVipWithDefault err, %v", err)
			return
		}
		reward, vipMeetGapType, vipIncrementDetail = p.GetRewardDetails().GetReward(progress, vip.VipRule.VIPLevel)
		// need to return a non-zore rewards for voucher to be insert.
		reward = 100
	default:
		var vip ploutos.VipRecord
		vip, err = model.GetVipWithDefault(c, userID)
		if err != nil {
			log.Printf("GetVipWithDefault err, %v", err)
			return
		}
		reward, vipMeetGapType, vipIncrementDetail = p.GetRewardDetails().GetReward(progress, vip.VipRule.VIPLevel)
	}
	return
}

var ErrPromotionSessionUnknownPromotionType = errors.New("unknown promotion type for calculating progress. please explicit express promotion type")

// GetPromotionSessionProgress
func GetPromotionSessionProgress(ctx context.Context, p ploutos.Promotion, s ploutos.PromotionSession, userID int64) (int64, error) {
	switch p.Type {
	case ploutos.PromotionTypeVipWeeklyB:
		//may need to check deposit requirement + vip
		vip, _ := model.GetVipWithDefault(ctx, userID)
		return vip.VipRule.VIPLevel, nil
		// not necessary
	// case ploutos.PromotionTypeVipBirthdayB, ploutos.PromotionTypeVipPromotionB:
	// 	vip, _ := model.GetVipWithDefault(c, userID)
	// 	progress = vip.VipRule.VIPLevel
	case ploutos.PromotionTypeFirstDepB, ploutos.PromotionTypeFirstDepIns:
		order, err := model.FirstTopup(ctx, userID)
		if util.IsGormNotFound(err) {
			return 0, err
		} else if err != nil {
			return 0, err
		}
		progress := order.AppliedCashInAmount
		log.Printf("progress is %d, userid %d order %d", progress, userID, order)
		return order.AppliedCashInAmount, nil
	case ploutos.PromotionTypeReDepB:

		orders, err := model.ScopedTopupExceptAllTimeFirst(ctx, userID, s.TopupStart, s.TopupEnd)
		if err != nil {
			return 0, err
		}
		return util.Reduce(orders, func(amount int64, input model.CashOrder) int64 {
			return amount + input.AppliedCashInAmount
		}, 0), nil
	case ploutos.PromotionTypeReDepIns:
		orders, err := model.ScopedTopupExceptAllTimeFirst(ctx, userID, s.TopupStart, s.TopupEnd)
		if err != nil {
			return 0, err
		}

		var progress int64 = 0
		for _, o := range orders {
			if progress < o.AppliedCashInAmount {
				progress = o.AppliedCashInAmount
			}
		}
		return progress, nil
	}

	return 0, nil
}

func GetPromotionSessionClaimStatus(c context.Context, p ploutos.Promotion, s ploutos.PromotionSession, userID int64, now time.Time) (claim serializer.ClaimStatus) {
	claim.ClaimStart = s.ClaimStart.Unix()
	claim.ClaimEnd = s.ClaimEnd.Unix()
	switch p.Type {
	case ploutos.PromotionTypeVipRebate, ploutos.PromotionTypeVipPromotionB, ploutos.PromotionTypeVipWeeklyB, ploutos.PromotionTypeVipBirthdayB:
		v, err := model.VoucherGetByUniqueID(c, ploutos.GenerateVoucherUniqueId(p.Type, p.ID, s.ID, userID, 0, buildSuffixByType(c, p, userID)))
		if err == nil && v.ID != 0 {
			claim.HasClaimed = true
		}
	case ploutos.PromotionTypeFirstDepB, ploutos.PromotionTypeFirstDepIns:
		v, err := model.GetVoucherByUserAndPromotionSession(c, userID, s.ID)
		if err == nil && v.ID != 0 {
			claim.HasClaimed = true
		} else {
			order, err := model.FirstTopup(c, userID)
			if err == nil {
				claim.ClaimEnd = order.CreatedAt.Add(7 * 24 * time.Hour).Unix()
			}
		}
	case ploutos.PromotionTypeVipReferral, ploutos.PromotionTypeSpinWheel:
		// since this promotion is able to claim multiple times, thus, we need to treat it as not claimed.
		// and both of the promotion has their own management for has_claimed
		claim.HasClaimed = false
	default:
		v, err := model.GetVoucherByUserAndPromotionSession(c, userID, s.ID)
		if err != nil {
			log.Printf("GetVoucherByUserAndPromotionSession err, %v", err)
		}
		if err == nil && v.ID != 0 {
			claim.HasClaimed = true
		}
	}
	return
}

func ClaimVoucherByType(c context.Context, p ploutos.Promotion, s ploutos.PromotionSession, v ploutos.VoucherTemplate, userID, promotionRequestID int64, rewardAmount int64, now time.Time, meetGapType int64, vipIncrementDetail ploutos.VipIncrementDetail) (voucher ploutos.Voucher, err error) {
	voucher = CraftVoucherByType(c, p, s, v, rewardAmount, userID, promotionRequestID, now, meetGapType, vipIncrementDetail)
	lang := model.GetUserLang(userID)

	switch p.Type {
	case ploutos.PromotionTypeFirstDepB, ploutos.PromotionTypeReDepB, ploutos.PromotionTypeBeginnerB, ploutos.PromotionTypeOneTimeDepB:
		//add money and insert voucher
		// add cash order
		err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) error {
			wagerChange := voucher.WagerMultiplier * rewardAmount
			err = CreateCashOrder(tx, p.Type, userID, rewardAmount, wagerChange, "", "")
			if err != nil {
				return err
			}
			err = tx.Create(&voucher).Error
			if err != nil {
				return err
			}
			if p.Type == ploutos.PromotionTypeBeginnerB {
				err = model.CreateUserAchievement(userID, ploutos.UserAchievementIdFirstAppLoginReward)
				if err != nil {
					return err
				}
			}
			if p.Type == ploutos.PromotionTypeOneTimeDepB {
				err = model.CreateUserAchievement(userID, ploutos.UserAchievementIdFirstDepositBonusReward)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err == nil {
			common.SendCashNotificationWithoutCurrencyId(userID, consts.Notification_Type_Deposit_Bonus, common.NOTIFICATION_DEPOSIT_BONUS_SUCCESS_TITLE, common.NOTIFICATION_DEPOSIT_BONUS_SUCCESS, rewardAmount)
		}
	case ploutos.PromotionTypeVipReferral:
		err = claimVoucherReferralVip(c, p, voucher, userID, now)
		if err == nil {
			common.SendNotification(userID, consts.Notification_Type_Referral_Alliance, conf.GetI18N(lang).T(common.NOTIFICATION_REFERRAL_ALLIANCE_TITLE), conf.GetI18N(lang).T(common.NOTIFICATION_REFERRAL_ALLIANCE))
		}
	case ploutos.PromotionTypeVipBirthdayB:
		err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) error {
			wagerChange := voucher.WagerMultiplier * rewardAmount
			err = CreateCashOrder(tx, p.Type, userID, rewardAmount, wagerChange, "", "")
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
	case ploutos.PromotionTypeVipRebate, ploutos.PromotionTypeVipPromotionB, ploutos.PromotionTypeVipWeeklyB:
		err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) error {
			wagerChange := voucher.WagerMultiplier * rewardAmount
			err = CreateCashOrder(tx, p.Type, userID, rewardAmount, wagerChange, "", "")
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
			switch p.Type {
			case ploutos.PromotionTypeVipRebate:
				nType = consts.Notification_Type_Rebate
				title = conf.GetI18N(lang).T(common.NOTIFICATION_REBATE_TITLE)
				text = conf.GetI18N(lang).T(common.NOTIFICATION_REBATE)
			case ploutos.PromotionTypeVipPromotionB:
				nType = consts.Notification_Type_Vip_Promotion_Bonus
				title = conf.GetI18N(lang).T(common.NOTIFICATION_VIP_PROMOTION_BONUS_TITLE)
				text = conf.GetI18N(lang).T(common.NOTIFICATION_VIP_PROMOTION_BONUS)
			case ploutos.PromotionTypeVipWeeklyB:
				nType = consts.Notification_Type_Weekly_Bonus
				title = conf.GetI18N(lang).T(common.NOTIFICATION_WEEKLY_BONUS_TITLE)
				text = conf.GetI18N(lang).T(common.NOTIFICATION_WEEKLY_BONUS)
			}
			common.SendNotification(userID, nType, title, text)
		}
	case ploutos.PromotionTypeFirstDepIns, ploutos.PromotionTypeReDepIns:
		//insert voucher only
		err = model.DB.Create(&voucher).Error
	case ploutos.PromotionTypeCustomTemplate:
		err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) error {
			wagerChange := voucher.WagerMultiplier * rewardAmount
			err = CreateCashOrder(tx, p.Type, userID, rewardAmount, wagerChange, "", p.Name)
			if err != nil {
				return err
			}
			err = tx.Create(&voucher).Error
			if err != nil {
				return err
			}
			return nil
		})
	case ploutos.PromotionTypeSpinWheel:
		fmt.Println("promotion.PromotionTypeSpinWheel ")
		// TODO move the cash order to here as well 
		var spin_items []ploutos.SpinItem

		// Build the GORM query
		model.DB.
			Table("spin_items si").
			Joins("INNER JOIN spins sp ON si.spin_id = sp.id").
			Joins("INNER JOIN spin_results sr ON si.id = sr.spin_result").
			Where("sp.promotion_id = ?", p.ID).
			Where("sr.user_id = ?", userID).
			Where("sr.redeemed = ?", false).
			Find(&spin_items)
		

		for _,spin_item:=range spin_items{
			voucher.Amount = int64(spin_item.Amount)
			voucher.WagerMultiplier = spin_item.Wager
			voucher.ReferenceID = strconv.FormatInt(spin_item.ID, 10)
			voucher.Status = ploutos.VoucherStatusRedeemed
			err = model.DB.Create(&voucher).Error
			if err != nil {
				fmt.Println("promotion.PromotionTypeSpinWheel creation failed")
				return 
			}
		}
		fmt.Println("promotion.PromotionTypeSpinWheel creation ends")
	}
	return
}

func GetPromotionExtraDetails(c context.Context, p ploutos.Promotion, userID int64, now time.Time) any {
	var extra any
	switch p.Type {
	case ploutos.PromotionTypeVipReferral:
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

	return extra
}

func CreateCashOrder(tx *gorm.DB, promoType, userId, rewardAmount, wagerChange int64, notes, name string) error {
	txType := promotionTypeToTransactionTypeMapping[promoType]
	sum, err := model.UpdateDbUserSumAndCreateTransaction(tx,
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
	orderType := promotionTypeToCashOrderType[promoType]
	dummyOrder := ploutos.CashOrder{
		ID:                    uuid.NewString(),
		UserId:                userId,
		OrderType:             orderType,
		Status:                ploutos.CashOrderStatusSuccess,
		Notes:                 ploutos.EncryptedStr(notes),
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

func CraftVoucherByType(c context.Context, p ploutos.Promotion, ps ploutos.PromotionSession, v ploutos.VoucherTemplate, rewardAmount, userID int64, promotionRequestID int64, now time.Time, meetGapType int64, vipIncrementDetail ploutos.VipIncrementDetail) (voucher ploutos.Voucher) {
	endAt := earlier(v.EndAt, v.GetExpiryTimeStamp(now, p.Timezone))
	status := ploutos.VoucherStatusReady
	isUsable := false
	suffix := buildSuffixByType(c, p, userID)
	switch p.Type {
	case ploutos.PromotionTypeFirstDepB, ploutos.PromotionTypeReDepB, ploutos.PromotionTypeBeginnerB, ploutos.PromotionTypeVipReferral, ploutos.PromotionTypeVipRebate, ploutos.PromotionTypeVipPromotionB, ploutos.PromotionTypeVipWeeklyB, ploutos.PromotionTypeVipBirthdayB, ploutos.PromotionTypeCustomTemplate:
		status = ploutos.VoucherStatusRedeemed
	case ploutos.PromotionTypeFirstDepIns, ploutos.PromotionTypeReDepIns:
		isUsable = true
	}

	voucherName := model.AmountReplace(v.Name, float64(rewardAmount)/100)
	wagerMultiplier := v.WagerMultiplier
	switch p.Type {
	case ploutos.PromotionTypeVipRebate:
		vip, _ := model.GetVipWithDefault(c, userID)
		wagerMultiplier = vip.VipRule.BirthdayBenefitWagerMultiplier
	case ploutos.PromotionTypeVipPromotionB:
		vip, _ := model.GetVipWithDefault(c, userID)
		wagerMultiplier = vip.VipRule.PromotionBenefitWagerMultiplier
	case ploutos.PromotionTypeVipWeeklyB:
		vip, _ := model.GetVipWithDefault(c, userID)
		wagerMultiplier = vip.VipRule.WeeklyBenefitWagerMultiplier
	case ploutos.PromotionTypeVipBirthdayB:
		vip, _ := model.GetVipWithDefault(c, userID)
		wagerMultiplier = vip.VipRule.BirthdayBenefitWagerMultiplier
	case ploutos.PromotionTypeCustomTemplate:
		voucherName = p.Name
	}

	voucher = ploutos.Voucher{
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
		PromotionSessionID: ps.ID,
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

func ValidateVoucherUsageByType(v ploutos.Voucher, oddsFormat, matchType int, odds float64, betAmount int64, isParley bool) (ret bool) {
	ret = false
	switch v.PromotionType {
	case ploutos.PromotionTypeFirstDepIns, ploutos.PromotionTypeReDepIns:
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

func ValidateUsageDetailsByType(v ploutos.Voucher, matchType int, odds float64, betAmount int64) (ret bool) {
	ret = false
	switch v.PromotionType {
	case ploutos.PromotionTypeFirstDepIns, ploutos.PromotionTypeReDepIns:
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

func claimVoucherReferralVip(c context.Context, p ploutos.Promotion, voucher ploutos.Voucher, userID int64, now time.Time) error {
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
		err = CreateCashOrder(tx, p.Type, user.ID, totalClaimable, wagerChange, cashOrderNotes, "")
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

func claimReferralAllianceRewards(tx *gorm.DB, referrerId int64, now time.Time) ([]ploutos.ReferralAllianceReward, error) {
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

func getSameDayVipRewardRecord(tx *gorm.DB, userID, prmotionID int64) ploutos.VipRewardRecords {
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

func buildSuffixByType(c context.Context, p ploutos.Promotion, userID int64) string {
	today := time.Now()
	suffix := ""
	vip, _ := model.GetVipWithDefault(c, userID)
	switch p.Type {
	case ploutos.PromotionTypeVipRebate:
		suffix = fmt.Sprintf("date-%s", today.Format(time.DateOnly))
	case ploutos.PromotionTypeVipWeeklyB:
	case ploutos.PromotionTypeVipBirthdayB:
		suffix = fmt.Sprintf("year-%d", today.Year())
	case ploutos.PromotionTypeVipPromotionB:
		suffix = fmt.Sprintf("vip-%d", vip.VipRule.VIPLevel)
	case ploutos.PromotionTypeSpinWheel, ploutos.PromotionTypeVipReferral:
		suffix = fmt.Sprintf("time-%s", today.Format("2006-01-02 15:04:05"))
	}
	return suffix
}
