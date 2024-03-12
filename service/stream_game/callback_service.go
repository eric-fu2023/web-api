package stream_game

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"
)

type BetObj struct {
	GameSession ploutos.StreamGameSession `json:"game_session"`
	Selection   int64                     `json:"selection"`
}

type Callback struct{}

func (c *Callback) NewCallback(userId int64) {}

func (c *Callback) GetGameVendorId() int64 {
	return consts.GameVendor["stream_game"]
}

func (c *Callback) ShouldProceed() bool {
	return true
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
	DrawId    int64       `json:"draw_id" form:"draw_id" binding:"required"`
	Selection int64       `json:"selection" form:"selection" binding:"required"`
	Amount    float64     `json:"amount" form:"amount" binding:"required"`
	User      *model.User `json:"user"`
}

func (c *PlaceOrder) GetExternalUserId() string {
	return c.User.Username
}

func (c *PlaceOrder) GetGameTransactionId() int64 {
	return c.DrawId
}

func (c *PlaceOrder) SaveGameTransaction(tx *gorm.DB) error {
	var draw ploutos.StreamGameSession
	err := model.DB.Where(`id`, c.DrawId).First(&draw).Error
	if err != nil {
		return err
	}
	bet := BetObj{
		GameSession: draw,
		Selection:   c.Selection,
	}
	j, err := json.Marshal(&bet)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	businessId := fmt.Sprintf(`%d%d%d`, c.DrawId, c.User.ID, now.Unix())
	betReport := ploutos.StreamGameBetReport{
		UserId:     c.User.ID,
		OrderId:    "SG" + businessId,
		BusinessId: businessId,
		GameType:   consts.GameVendor["stream_game"],
		InfoJson:   j,
		Bet:        util.MoneyInt(c.Amount),
		BetTime:    &now,
		Status:     4, // confirmed
		GameId:     draw.StreamGameId,
		RefId:      c.DrawId,
	}
	err = tx.Omit("id").Create(&betReport).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (c *PlaceOrder) GetAmount() int64 {
	return util.MoneyInt(c.Amount)
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
	go common.LogGameCallbackRequest("stream_game_place_order", req)
	i18n := c.MustGet("i18n").(i18n.I18n)
	user := c.MustGet("user").(model.User)
	var draw ploutos.StreamGameSession
	err = model.DB.Where(`id`, req.DrawId).Where(`status`, ploutos.StreamGameSessionStatusOpen).First(&draw).Error
	if err != nil {
		res = serializer.ParamErr(c, req, i18n.T("invalid_draw_id"), err)
		return
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
		err = model.DB.Where(`game_vendor_id`, consts.GameVendor["stream_game"]).Where(`currency_id`, user.CurrencyId).First(&currency).Error
		if err != nil {
			res = serializer.Err(c, req, serializer.CodeGeneralError, i18n.T("empty_currency_id"), err)
			return
		}
		var game UserRegister
		err = game.CreateUser(user, currency.Value)
		if err != nil {
			res = serializer.Err(c, req, serializer.CodeGeneralError, i18n.T("stream_game_create_user_failed"), err)
			return
		}
		err = common.ProcessTransaction(&req)
		if err != nil {
			res = serializer.Err(c, req, serializer.CodeGeneralError, i18n.T("general_error"), err)
			return
		}
	}
	res = serializer.Response{
		Msg: i18n.T("success"),
	}
	return
}

func Settle(c *gin.Context, req SettleOrder) (res serializer.Response, err error) {
	go common.LogGameCallbackRequest("dollar_jackpot_place_order", req)
	var br ploutos.DollarJackpotBetReport
	err = model.DB.Where(`business_id`, req.BusinessId).Where(`status`, 4).First(&br).Error // 4: unsettled
	if err != nil {
		res = serializer.Err(c, req, serializer.CodeGeneralError, "dollar jackpot settle error", err)
		return
	}
	req.DrawId = br.GameId
	req.BetAmount = br.Bet
	err = common.ProcessTransaction(&req)
	if err != nil {
		res = serializer.Err(c, req, serializer.CodeGeneralError, "dollar jackpot settle error", err)
		return
	}
	res = serializer.Response{
		Msg: "Success",
	}
	return
}
