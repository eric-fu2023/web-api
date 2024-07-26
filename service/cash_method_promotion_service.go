package service

import (
	"context"
	"web-api/model"
	"web-api/service/common"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

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
	util.GetLoggerEntry(c).Info("HandleCashMethodPromotion vipRecord.VipRule.ID", vipRecord.VipRule.ID, order.ID) // wl: for staging debug

	// check cash method and vip combination has promotion or not
	cashMethodPromotion, err := model.FindActiveCashMethodPromotionByCashMethodIdAndVipId(order.CashMethodId, vipRecord.VipRule.ID, &order.CreatedAt, nil)
	if err != nil {
		util.GetLoggerEntry(c).Error("HandleCashMethodPromotion Find CashMethodPromotion", err, order.ID)
		return
	}
	if cashMethodPromotion.ID == 0 {
		util.GetLoggerEntry(c).Error("HandleCashMethodPromotion no CashMethodPromotion", order.ID)
		return
	}
	util.GetLoggerEntry(c).Info("HandleCashMethodPromotion cashMethodPromotion.ID", cashMethodPromotion.ID, order.ID) // wl: for staging debug

	// check over payout limit or not
	weeklyAmountRecord, dailyAmountRecord, err := model.GetWeeklyAndDailyCashMethodPromotionRecord(c, order.CashMethodId, order.UserId)
	if err != nil {
		util.GetLoggerEntry(c).Error("HandleCashMethodPromotion GetWeeklyAndDailyCashMethodPromotionRecord", err, order.ID)
		return
	}
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

	coAmount := order.AppliedCashInAmount
	if order.OrderType == -1 {
		coAmount = order.AppliedCashOutAmount
	}

	amount, err := model.GetMaxCashMethodPromotionAmount(c, weeklyAmount, dailyAmount, cashMethodPromotion, order.UserId, coAmount, false)
	if err != nil {
		util.GetLoggerEntry(c).Error("HandleCashMethodPromotion GetMaxAmountPayment", err, order.ID)
	}

	if amount == 0 {
		util.GetLoggerEntry(c).Info("HandleCashMethodPromotion amount == 0", order.ID)
		return
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
		util.GetLoggerEntry(c).Info("HandleCashMethodPromotion new co", coId, order.ID) // wl: for staging debug
		common.SendUserSumSocketMsg(order.UserId, sum.UserSum, "promotion", float64(amount)/100)

		return
	})
	if err != nil {
		util.GetLoggerEntry(c).Error("HandleCashMethodPromotion start", err, order.ID)
		return
	}
}
