package on_cash_orders

import (
	"context"
	"strconv"
	"time"

	"web-api/model"
	"web-api/service"
	"web-api/service/promotion"
	"web-api/util"
)

func OneTimeBonusPromotion(ctx context.Context, order model.CashOrder) {
	var user model.User
	if err := model.DB.Where(`id`, order.UserId).First(&user).Error; err != nil {
		util.GetLoggerEntry(ctx).Error("get user error", err)
		return
	}
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
