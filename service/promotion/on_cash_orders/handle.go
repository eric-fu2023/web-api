package on_cash_orders

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"web-api/model"
	"web-api/service"
	"web-api/service/promotion"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

// Note: Work in progress
// CashOrderEventType, PaymentGateway is a workaround to conditionally trigger promotions. See [Handle]
// may be possible to derive args from cash order.

// CashOrderEventType
type CashOrderEventType = int64

const (
	CashOrderEventTypeClose CashOrderEventType = 2
)

type PaymentGateway = string

const (
	PaymentGatewayFinPay PaymentGateway = "finpay"
	PaymentGatewayForay  PaymentGateway = "foray"
)

type RequestMode = int64

const (
	RequestModeCallback RequestMode = 1
	// RequestModeManual via internal routes (?)
	RequestModeManual RequestMode = 2
)

// Handle
// Note: take awareness on the trigger conditions and also the sequence of process~ these will change as requirement comes in
// See also: [Note: Work in progress]c
func Handle(ctx context.Context, order model.CashOrder, transactionType models.TransactionType, eventType CashOrderEventType, gateway PaymentGateway, requestMode RequestMode) error {
	switch eventType {
	case CashOrderEventTypeClose:
	default:
		return fmt.Errorf("unsupported event type %d", eventType)
	}

	switch gateway {
	case PaymentGatewayFinPay, PaymentGatewayForay:
	default:
		return fmt.Errorf("unsupported gateway %s", gateway)
	}

	{
		var shouldHandleOneTimeBonus bool
		shouldHandleOneTimeBonus = transactionType == models.TransactionTypeCashIn

		if shouldHandleOneTimeBonus {
			HandleOneTimeBonus(ctx, order)
		}
	}

	{
		shouldHandleCashMethodPromotion := false
		switch {
		case models.TransactionTypeCashIn == transactionType:
			shouldHandleCashMethodPromotion = true
		case models.TransactionTypeCashOut == transactionType && RequestModeCallback == requestMode:
			shouldHandleCashMethodPromotion = true
		default:
			return fmt.Errorf("unknown transaction type for shouldHandleCashMethodPromotion %d", transactionType)
		}

		if shouldHandleCashMethodPromotion {
			service.HandleCashMethodPromotion(ctx, order)
		}
	}
	return nil
}

func HandleOneTimeBonus(ctx context.Context, order model.CashOrder) {
	var user model.User
	if err := model.DB.Where(`id`, order.UserId).First(&user).Error; err != nil {
		util.GetLoggerEntry(ctx).Error("get user error", err)
		return
	}
	// uaCond := model.GetUserAchievementCond{AchievementIds: []int64{
	// 	models.UserAchievementIdFirstDepositBonusTutorial,
	// }}
	// a, err := model.GetUserAchievements(order.UserId, uaCond)
	// if err != nil {
	// 	util.GetLoggerEntry(c).Error("get config error", err)
	// 	return
	// }
	// if len(a) == 0 {
	// 	return
	// }
	// if len(a) > 0 && order.CreatedAt.Sub(a[0].CreatedAt) > 1*time.Hour {
	// 	util.GetLoggerEntry(c).Info("not in reward timeframe", order.AppliedCashInAmount)
	// 	return
	// }
	amt, err := service.GetCachedConfigBranded(context.TODO(), "static_promotion_one_time_bonus_min_amount", user.BrandId)
	if err != nil {
		util.GetLoggerEntry(ctx).Error("get config error", err)
		return
	}
	minAmt, _ := strconv.Atoi(amt)
	if order.AppliedCashInAmount < int64(minAmt) {
		util.GetLoggerEntry(ctx).Info("insufficient amount", order.AppliedCashInAmount)
		return
	}
	cachedPromotionId, err := service.GetCachedConfigBranded(context.Background(), "static_promotion_one_time_bonus_id", user.BrandId)
	if err != nil {
		util.GetLoggerEntry(ctx).Error("get config error", err)
		return
	}
	promotionId, err := strconv.ParseInt(cachedPromotionId, 10, 64)
	if err != nil {
		util.GetLoggerEntry(ctx).Error("conv cachedPromotionId error", err)
		return
	}
	now := time.Now()
	p, err := model.PromotionGetActive(context.TODO(), int(user.BrandId), int64(promotionId), now)
	if err != nil {
		util.GetLoggerEntry(ctx).Error("get promotion error", err)
		return
	}
	s, err := model.PromotionSessionGetActive(context.TODO(), p.ID, now)
	if err != nil {
		util.GetLoggerEntry(ctx).Error("get promotion session error", err)
		return
	}
	_, err = promotion.Claim(ctx, time.Now(), p, s, user.ID, &user)
	if err != nil {
		util.GetLoggerEntry(ctx).Error("claim one time bonus error", err)
		return
	}
}
