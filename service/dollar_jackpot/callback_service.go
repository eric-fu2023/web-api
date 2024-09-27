package dollar_jackpot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
	"web-api/cache"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Callback struct{}

func (c *Callback) NewCallback(userId int64) {}

func (c *Callback) GetGameVendorId() int64 {
	return consts.GameVendor["dollar_jackpot"]
}

func (c *Callback) ShouldProceed() bool {
	return true // dc doesn't have wager that shouldn't proceed
}

func (c *Callback) IsAdjustment() bool {
	return false
}

func (c *Callback) ApplyInsuranceVoucher(userId int64, betAmount int64, betExists bool) (err error) {
	// Voucher application not done
	return
}

type PlaceOrder struct {
	Callback
	Amount float64     `json:"amount" form:"amount" binding:"required"`
	DrawId int64       `json:"draw_id" form:"draw_id" binding:"required"`
	User   *model.User `json:"user"`
}

func (c *PlaceOrder) GetExternalUserId() string {
	return c.User.Username
}

func (c *PlaceOrder) GetGameTransactionId() int64 {
	return c.DrawId
}

func (c *PlaceOrder) SaveGameTransaction(tx *gorm.DB) error {
	var djd ploutos.DollarJackpotDraw
	err := model.DB.Preload(`DollarJackpot`).Where(`id`, c.DrawId).First(&djd).Error
	if err != nil {
		return err
	}
	j, err := json.Marshal(&djd)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	businessId := fmt.Sprintf(`%d%d%d`, c.DrawId, c.User.ID, now.Unix())
	betReport := ploutos.DollarJackpotBetReport{
		UserId:       c.User.ID,
		OrderId:      "DJ" + businessId,
		BusinessId:   businessId,
		GameType:     consts.GameVendor["dollar_jackpot"],
		InfoJson:     j,
		Bet:          util.MoneyInt(c.Amount),
		BetTime:      &now,
		Status:       4, // confirmed
		GameId:       c.DrawId,
		MaxWinAmount: fmt.Sprintf(`%d`, djd.DollarJackpot.Prize),
	}
	err = model.DB.Transaction(func(tx2 *gorm.DB) (err error) {
		err = tx2.Omit("id").Create(&betReport).Error
		if err != nil {
			return err
		}
		r := cache.RedisClient.IncrBy(context.TODO(), fmt.Sprintf(DollarJackpotRedisKey, c.DrawId), util.MoneyInt(c.Amount))
		if r.Err() != nil {
			return err
		}
		return nil
	})
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (c *PlaceOrder) GetAmount() int64 {
	return -1 * util.MoneyInt(c.Amount)
}

func (c *PlaceOrder) GetWagerMultiplier() (value int64, exists bool) {
	return 0, false
}

func (c *PlaceOrder) GetBetAmount() (amount int64, exists bool) {
	return 0, false
}

type SettleOrder struct {
	Callback
	Amount          *float64 `json:"amount" form:"amount" binding:"required"`
	BusinessId      string   `json:"business_id" form:"business_id" binding:"required"`
	Username        string   `json:"username" form:"username" binding:"required"`
	WagerMultiplier int64    `json:"wager_multiplier" form:"wager_multiplier" binding:"required"`
	BetAmount       int64    `json:"bet_amount"`
	DrawId          int64    `json:"draw_id"`
}

func (c *SettleOrder) GetExternalUserId() string {
	return c.Username
}

func (c *SettleOrder) GetGameTransactionId() int64 {
	return c.DrawId
}

func (c *SettleOrder) SaveGameTransaction(tx *gorm.DB) error {
	return nil
}

func (c *SettleOrder) GetAmount() int64 {
	return util.MoneyInt(*c.Amount)
}

func (c *SettleOrder) GetWagerMultiplier() (value int64, exists bool) {
	if *c.Amount == 0 { // lost bets
		return -1, true
	}
	return c.WagerMultiplier, true
}

func (c *SettleOrder) GetBetAmount() (amount int64, exists bool) {
	if *c.Amount == 0 { // lost bets
		return c.BetAmount, true
	}
	return util.MoneyInt(*c.Amount) * 2, true // won bets; c.amount * 2 because the equation in processTransaction is betAmount - winAmount
}

func Place(c *gin.Context, req PlaceOrder) (res serializer.Response, err error) {
	go common.LogGameCallbackRequest("dollar_jackpot_place_order", req)
	i18n := c.MustGet("i18n").(i18n.I18n)
	user := c.MustGet("user").(model.User)
	brand := c.MustGet(`_brand`).(int)
	var djd ploutos.DollarJackpotDraw
	//err = model.DB.Joins(`JOIN dollar_jackpots ON dollar_jackpots.id = dollar_jackpot_draws.dollar_jackpot_id AND dollar_jackpots.status = 1 AND dollar_jackpots.brand_id = ?`, brand).
	//	Where(`dollar_jackpot_draws.id`, req.DrawId).Where(`dollar_jackpot_draws.status`, 0).First(&djd).Error
	err = model.DB.InnerJoins(`DollarJackpot`, model.DB.Where(&ploutos.DollarJackpot{Status: 1, BrandId: int64(brand)})).
		Where(`dollar_jackpot_draws.id`, req.DrawId).Where(`dollar_jackpot_draws.status`, 0).First(&djd).Error
	if err != nil {
		res = serializer.ParamErr(c, req, i18n.T("invalid_draw_id"), err)
		return
	}
	var sum model.ContributionSum
	err = model.DB.Model(ploutos.DollarJackpotBetReport{}).Scopes(model.GetContribution(user.ID, djd.ID)).Find(&sum).Error
	if err != nil {
		res = serializer.Err(c, req, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	if djd.DollarJackpot != nil {
		limit := float64(djd.DollarJackpot.Prize) * model.ContributionLimitPercent
		totalContrib := sum.Sum + util.MoneyInt(req.Amount)
		if limit < float64(totalContrib) {
			res = serializer.ParamErr(c, req, i18n.T("contribution_limit_reached"), err)
			return
		}
	}
	var total_sum model.ContributionSum
	err = model.DB.Model(ploutos.DollarJackpotBetReport{}).Scopes(model.GetTotalContribution(djd.ID)).Find(&total_sum).Error
	if djd.DollarJackpot != nil {
		limit := float64(djd.DollarJackpot.Prize)
		totalContrib := total_sum.Sum + util.MoneyInt(req.Amount)
		if limit < float64(totalContrib) {
			res = serializer.ParamErr(c, req, i18n.T("prize_limit_reached"), err)
			return
		}
	}

	req.User = &user
	err = common.ProcessTransaction(&req)
	if err != nil {
		if !errors.Is(err, common.ErrGameVendorUserInvalid) {
			res = serializer.Err(c, req, serializer.CodeGeneralError, i18n.T("general_error"), err)
			return
		}
		// if error is due to user not being registered with the game, retry registration
		var currency ploutos.CurrencyGameVendor
		err = model.DB.Where(`game_vendor_id`, consts.GameVendor["dollar_jackpot"]).Where(`currency_id`, user.CurrencyId).First(&currency).Error
		if err != nil {
			res = serializer.Err(c, req, serializer.CodeGeneralError, i18n.T("empty_currency_id"), err)
			return
		}
		var game UserRegister
		err = game.CreateUser(user, currency.Value)
		if err != nil {
			res = serializer.Err(c, req, serializer.CodeGeneralError, i18n.T("dollar_jackpot_create_user_failed"), err)
			return
		}
		err = common.ProcessTransaction(&req)
		if err != nil {
			res = serializer.Err(c, req, serializer.CodeGeneralError, i18n.T("general_error"), err)
			return
		}
	}
	
	// staging mqtt got problem, need to fix that 1st
	// pb := &paho.Publish{
	// 	Topic:   fmt.Sprintf(`/dollar_jackpot/%s`, djd.DollarJackpotId),
	// 	QoS:     byte(1),
	// 	Payload: []byte("{}"),
	// }
	// _, err = util.MQTTClient.Publish(context.Background(), pb)
	// if err != nil {
	// 	util.Log().Error("dollar_jackpot send refresh message send error", err)
	// 	return
	// }

	res = serializer.Response{
		Msg: i18n.T("success"),
	}
	return
}

func Settle(c *gin.Context, req SettleOrder) (res serializer.Response, err error) {
	go common.LogGameCallbackRequest("dollar_jackpot_settle_order", req)
	var br ploutos.DollarJackpotBetReport
	err = model.DB.Where(`business_id`, req.BusinessId).Where(`status`, 4).First(&br).Error // 4: unsettled
	if err != nil {
		res = serializer.Err(c, req, serializer.CodeGeneralError, "dollar jackpot settle error", err)
		return
	}
	req.DrawId = br.GameId
	req.BetAmount = br.Bet
	if os.Getenv("PRODUCT") == "batace"{
		err = common.ProcessTransactionBatace(&req)
	} else {
		err = common.ProcessTransaction(&req)
	}
	if err != nil {
		res = serializer.Err(c, req, serializer.CodeGeneralError, "dollar jackpot settle error", err)
		return
	}
	res = serializer.Response{
		Msg: "Success",
	}
	return
}
