package cashin

import (
	"context"
	"math"
	"strconv"
	"time"
	"web-api/model"
	"web-api/service"
	"web-api/service/common"
	"web-api/service/promotion"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

func HandlePromotion(c context.Context, order model.CashOrder) {
	HandleOneTimeB(c, order)
	HandleCashMethodPromotion(c, order)
}

func HandleOneTimeB(c context.Context, order model.CashOrder) {
	var user model.User
	if err := model.DB.Where(`id`, order.UserId).First(&user).Error; err != nil {
		util.GetLoggerEntry(c).Error("get user error", err)
		return
	}
	uaCond := model.GetUserAchievementCond{AchievementIds: []int64{
		models.UserAchievementIdFirstDepositBonusTutorial,
	}}
	a, err := model.GetUserAchievements(order.UserId, uaCond)
	if err != nil {
		util.GetLoggerEntry(c).Error("get config error", err)
		return
	}
	if len(a) == 0 {
		return
	}
	if len(a) > 0 && order.CreatedAt.Sub(a[0].CreatedAt) > 1*time.Hour {
		util.GetLoggerEntry(c).Info("not in reward timeframe", order.AppliedCashInAmount)
		return
	}
	amt, err := service.GetCachedConfigBranded(context.Background(), "static_promotion_one_time_bonus_min_amount", user.BrandId)
	if err != nil {
		util.GetLoggerEntry(c).Error("get config error", err)
		return
	}
	minAmt, _ := strconv.Atoi(amt)
	if order.AppliedCashInAmount < int64(minAmt) {
		util.GetLoggerEntry(c).Info("insufficient amount", order.AppliedCashInAmount)
		return
	}
	v, err := service.GetCachedConfigBranded(context.Background(), "static_promotion_one_time_bonus_id", user.BrandId)
	if err != nil {
		util.GetLoggerEntry(c).Error("get config error", err)
		return
	}
	id, _ := strconv.Atoi(v)

	now := time.Now()
	p, err := model.PromotionGetActive(context.TODO(), int(user.BrandId), int64(id), now)
	if err != nil {
		util.GetLoggerEntry(c).Error("get promotion error", err)
		return
	}
	s, err := model.PromotionSessionGetActive(context.TODO(), p.ID, now)
	if err != nil {
		util.GetLoggerEntry(c).Error("get promotion session error", err)
		return
	}
	_, err = promotion.Claim(c, time.Now(), p, s, user.ID)
	if err != nil {
		util.GetLoggerEntry(c).Error("claim one time bonus error", err)
		return
	}
}

func HandleCashMethodPromotion(c context.Context, order model.CashOrder) {
	if order.CashMethodId == 0 {
		util.GetLoggerEntry(c).Info("HandleCashMethodPromotion order.CashMethodId == 0", order.ID)
		return
	}
	// check claim before or not
	cashMethodPromotionRecord, err := model.FindCashMethodPromotionRecordByCashOrderId(order.ID, nil)
	if err != nil {
		util.GetLoggerEntry(c).Error("HandleCashMethodPromotion FindCashMethodPromotionRecordByCashOrderId", err, order.ID)
		return
	}
	if cashMethodPromotionRecord.ID != 0 {
		util.GetLoggerEntry(c).Info("HandleCashMethodPromotion FindCashMethodPromotionRecordByCashOrderId order has been claimed", order.ID)
		return
	}

	vipRecord, err := model.GetVipWithDefault(c, order.UserId)
	if err != nil {
		util.GetLoggerEntry(c).Error("HandleCashMethodPromotion GetVipWithDefault", err, order.ID)
		return
	}
	util.GetLoggerEntry(c).Info("HandleCashMethodPromotion vipRecord.VipRule.ID", vipRecord.VipRule.ID) // wl: for staging debug

	// check cash method and vip combination has promotion or not
	cashMethodPromotion, err := model.FindCashMethodPromotionByCashMethodIdAndVipId(order.CashMethodId, vipRecord.VipRule.ID, nil)
	if err != nil {
		util.GetLoggerEntry(c).Error("HandleCashMethodPromotion Find CashMethodPromotion", err, order.ID)
		return
	}
	if cashMethodPromotion.ID == 0 {
		util.GetLoggerEntry(c).Error("HandleCashMethodPromotion no CashMethodPromotion", order.ID)
		return
	}
	util.GetLoggerEntry(c).Info("HandleCashMethodPromotion cashMethodPromotion.ID", cashMethodPromotion.ID) // wl: for staging debug

	// check over payout limit or not
	now := time.Now()
	weeklyAmount, err := model.AggreCashMethodPromotionRecordAmountByCashMethodIdAndUserId(order.CashMethodId, order.UserId, now.AddDate(0, 0, -7), now, nil)
	if err != nil {
		util.GetLoggerEntry(c).Error("HandleCashMethodPromotion AggreCashMethodPromotionRecordAmountByCashMethodIdAndUserId", order.ID)
		return
	}
	dailyAmount, err := model.AggreCashMethodPromotionRecordAmountByCashMethodIdAndUserId(order.CashMethodId, order.UserId, now.AddDate(0, 0, -1), now, nil)
	if err != nil {
		util.GetLoggerEntry(c).Error("HandleCashMethodPromotion AggreCashMethodPromotionRecordAmountByCashMethodIdAndUserId", order.ID)
		return
	}
	util.GetLoggerEntry(c).Info("HandleCashMethodPromotion weeklyAmount", weeklyAmount, order.CashMethodId, order.UserId, now.AddDate(0, 0, -7), now, order.ID) // wl: for staging debug
	util.GetLoggerEntry(c).Info("HandleCashMethodPromotion dailyAmount", dailyAmount, order.CashMethodId, order.UserId, now.AddDate(0, 0, -1), now, order.ID)   // wl: for staging debug
	if weeklyAmount >= cashMethodPromotion.WeeklyMaxPayout {
		util.GetLoggerEntry(c).Info("HandleCashMethodPromotion weeklyAmount >= cashMethodPromotion.WeeklyMaxPayout", weeklyAmount, cashMethodPromotion.WeeklyMaxPayout, order.ID)
		return
	}
	if dailyAmount >= cashMethodPromotion.DailyMaxPayout {
		util.GetLoggerEntry(c).Info("HandleCashMethodPromotion dailyAmount >= cashMethodPromotion.DailyMaxPayout", dailyAmount, cashMethodPromotion.DailyMaxPayout, order.ID)
		return
	}

	oriAmount := cashMethodPromotion.PayoutRate * float64(order.ActualCashInAmount)
	maxDailyPayoutRemaining := float64(cashMethodPromotion.DailyMaxPayout - dailyAmount)
	maxWeeklyPayoutRemaining := float64(cashMethodPromotion.WeeklyMaxPayout - weeklyAmount)

	amount := int64(math.Min(oriAmount, maxDailyPayoutRemaining))
	amount = int64(math.Min(float64(amount), maxWeeklyPayoutRemaining))
	util.GetLoggerEntry(c).Info("HandleCashMethodPromotion get min amount", oriAmount, maxDailyPayoutRemaining, maxWeeklyPayoutRemaining, amount, order.ID) // wl: for staging debug

	// cashMethodPromotion start
	err = model.DB.Clauses(dbresolver.Use("txConn")).Debug().WithContext(c).Transaction(func(tx *gorm.DB) (err error) {
		newCashMethodPromotionRecord := models.CashMethodPromotionRecord{
			CashMethodPromotionId: cashMethodPromotion.ID,
			CashMethodId:          cashMethodPromotion.CashMethodId,
			VipId:                 vipRecord.VipRule.ID,
			UserId:                order.UserId,
			PayoutRate:            cashMethodPromotion.PayoutRate,
			CashOrderId:           order.ID,
			Amount:                amount,
		}
		err = tx.Create(&newCashMethodPromotionRecord).Error
		if err != nil {
			return err
		}

		sum, err := model.UserSum{}.UpdateUserSumWithDB(tx, order.UserId, amount, amount, 0, models.TransactionTypeCashMethodPromotion, "")
		if err != nil {
			return err
		}
		notes := "dummy"

		coId := uuid.NewString()
		dummyOrder := models.CashOrder{
			ID:                    coId,
			UserId:                order.UserId,
			OrderType:             models.CashOrderTypeCashMethodPromotion,
			Status:                models.CashOrderStatusSuccess,
			Notes:                 models.EncryptedStr(notes),
			AppliedCashInAmount:   amount,
			ActualCashInAmount:    amount,
			EffectiveCashInAmount: amount,
			BalanceBefore:         sum.Balance - amount,
			WagerChange:           amount,
		}
		err = tx.Create(&dummyOrder).Error
		if err != nil {
			return err
		}
		util.GetLoggerEntry(c).Info("HandleCashMethodPromotion new co", coId) // wl: for staging debug
		common.SendUserSumSocketMsg(order.UserId, sum.UserSum, "promotion", float64(amount)/100)

		return
	})
	if err != nil {
		util.GetLoggerEntry(c).Error("HandleCashMethodPromotion start", err, order.ID)
		return
	}
}
