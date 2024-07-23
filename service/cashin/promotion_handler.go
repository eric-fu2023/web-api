package cashin

import (
	"context"
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
	weeklyAmountRecord, dailyAmountRecord, err := model.GetWeeklyAndDailyCashMethodPromotionRecord(c, order.CashMethodId, order.UserId)
	if cashMethodPromotion.ID == 0 {
		util.GetLoggerEntry(c).Error("HandleCashMethodPromotion no CashMethodPromotion", order.ID)
		return
	}
	var weeklyAmount int64
	if len(weeklyAmountRecord) > 0 {
		weeklyAmount = weeklyAmountRecord[0].Amount
	}
	var dailyAmount int64
	if len(dailyAmountRecord) > 0 {
		dailyAmount = dailyAmountRecord[0].Amount
	}
	amount, err := model.GetMaxCashMethodPromotionAmount(c, weeklyAmount, dailyAmount, cashMethodPromotion, order.UserId, order.ActualCashInAmount, false)
	if err != nil {
		util.GetLoggerEntry(c).Error("HandleCashMethodPromotion GetMaxAmountPayment", err, order.ID)
	}

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
