package cashin

import (
	"context"
	"strconv"
	"time"
	"web-api/model"
	"web-api/service"
	"web-api/service/promotion"
	"web-api/util"
)

func HandlePromotion(c context.Context, order model.CashOrder) {
	HandleOneTimeB(c, order)
}

func HandleOneTimeB(c context.Context, order model.CashOrder) {
	var user model.User
	if err := model.DB.Where(`id`, order.UserId).First(&user).Error; err != nil {
		util.GetLoggerEntry(c).Error("get user error", err)
		return
	}
	uaCond := model.GetUserAchievementCond{AchievementIds: []int64{
		model.UserAchievementIdFirstDepositBonusTutorial,
	}}
	a, err := model.GetUserAchievements(order.UserId, uaCond)
	if err != nil {
		util.GetLoggerEntry(c).Error("get config error", err)
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
	_, err = promotion.Claim(context.TODO(), time.Now(), p, s, user)
	if err != nil {
		util.GetLoggerEntry(c).Error("claim one time bonus error", err)
		return
	}
}
