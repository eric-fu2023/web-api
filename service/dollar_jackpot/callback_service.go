package dollar_jackpot

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"
)

type Callback struct {
	GameId int64   `json:"game_id" form:"game_id" binding:"required"`
	DrawId int64   `json:"draw_id" form:"draw_id" binding:"required"`
	Amount float64 `json:"amount" form:"amount" binding:"required"`
}

func (c *Callback) NewCallback(userId int64) {}

func (c *Callback) GetGameVendorId() int64 {
	return c.GameId
}

func (c *Callback) GetGameTransactionId() int64 {
	return c.DrawId
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
	User *model.User `json:"user"`
}

func (c *PlaceOrder) GetExternalUserId() string {
	return c.User.Username
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
	businessId := fmt.Sprintf(`%d%d%d%d`, c.GameId, c.DrawId, c.User.ID, now.Unix())
	betReport := ploutos.DollarJackpotBetReport{
		UserId:     c.User.ID,
		OrderId:    "DJ" + businessId,
		BusinessId: businessId,
		GameType:   c.GameId,
		InfoJson:   j,
		Bet:        util.MoneyInt(c.Amount),
		BetTime:    &now,
		Status:     4, // confirmed
		GameId:     c.DrawId,
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
	Username        string `json:"username" form:"username" binding:"required"`
	WagerMultiplier int64  `json:"wager_multiplier" form:"wager_multiplier" binding:"required"`
}

func (c *SettleOrder) GetExternalUserId() string {
	return c.Username
}

func (c *SettleOrder) SaveGameTransaction(tx *gorm.DB) error {
	return nil
}

func (c *SettleOrder) GetAmount() int64 {
	return util.MoneyInt(c.Amount)
}

func (c *SettleOrder) GetWagerMultiplier() (value int64, exists bool) {
	return c.WagerMultiplier, true
}

func (c *SettleOrder) GetBetAmount() (amount int64, exists bool) {
	return util.MoneyInt(c.Amount) * 2, true
}

func Place(c *gin.Context, req PlaceOrder) (res serializer.Response, err error) {
	go common.LogGameCallbackRequest("dollar_jackpot_place_order", req)
	i18n := c.MustGet("i18n").(i18n.I18n)
	user := c.MustGet("user").(model.User)
	brand := c.MustGet(`_brand`).(int)
	var djd ploutos.DollarJackpotDraw
	err = model.DB.Joins(`JOIN dollar_jackpots ON dollar_jackpots.id = dollar_jackpot_draws.dollar_jackpot_id AND dollar_jackpots.status = 1 AND dollar_jackpots.brand_id = ?`, brand).
		Where(`dollar_jackpot_draws.id`, req.DrawId).Where(`dollar_jackpot_draws.status`, 0).First(&djd).Error
	if err != nil {
		res = serializer.ParamErr(c, req, i18n.T("invalid_draw_id"), err)
		return
	}
	req.User = &user
	err = common.ProcessTransaction(&req)
	if err != nil {
		res = serializer.Err(c, req, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	res = serializer.Response{
		Msg: i18n.T("success"),
	}
	return
}

func Settle(c *gin.Context, req SettleOrder) (res serializer.Response, err error) {
	go common.LogGameCallbackRequest("dollar_jackpot_place_order", req)
	err = common.ProcessTransaction(&req)
	if err != nil {
		res = serializer.Err(c, req, serializer.CodeGeneralError, "Error", err)
		return
	}
	res = serializer.Response{
		Msg: "Success",
	}
	return
}
