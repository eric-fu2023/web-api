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

func HandlePromotion(order model.CashOrder) {
	HandleOneTimeB(order)
}

func HandleOneTimeB(order model.CashOrder) {
	v, err := service.GetCachedConfig(context.Background(), "static_promotion_one_time_bonus_id")
	if err != nil {
		util.Log().Error("get config error", err)
		return
	}
	id, _ := strconv.Atoi(v)
	var user model.User
	if err = model.DB.Where(`id`, order.UserId).First(&user).Error; err != nil {
		util.Log().Error("get user error", err)
		return
	}
	now := time.Now()
	p, err := model.PromotionGetActive(context.TODO(), int(user.BrandId), int64(id), now)
	if err != nil {
		util.Log().Error("get promotion error", err)
		return
	}
	s, err := model.PromotionSessionGetActive(context.TODO(), p.ID, now)
	if err != nil {
		util.Log().Error("get promotion session error", err)
		return
	}
	_, err = promotion.Claim(context.TODO(), time.Now(), p, s, user)
	if err != nil {
		util.Log().Error("claim one time bonus error", err)
		return
	}
}
