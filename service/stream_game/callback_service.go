package stream_game

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
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

var GameMultiple = map[int64]map[int64]float64{
	1: { // dice
		1: 1.95, // small
		2: 1.95, // big
		3: 31,   // triple
		4: 1.95, // odd
		5: 1.95, // even
	},
}

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
	err := model.DB.Preload(`StreamGame`).Where(`id`, c.DrawId).First(&draw).Error
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
		UserId:       c.User.ID,
		OrderId:      "SG" + businessId,
		BusinessId:   businessId,
		GameType:     consts.GameVendor["stream_game"],
		InfoJson:     j,
		Bet:          util.MoneyInt(c.Amount),
		BetTime:      &now,
		Status:       4, // confirmed
		GameId:       c.DrawId,
		MaxWinAmount: fmt.Sprintf(`%d`, util.MoneyInt(c.Amount*GameMultiple[draw.StreamGameId][c.Selection])),
	}
	err = tx.Omit("id").Create(&betReport).Error
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
func (c *PlaceOrder) GetBetAmountOnly() (amount int64) {
	return 0
}

type SettleOrder struct {
	Callback
	Amount     *float64 `json:"amount" form:"amount" binding:"required"`
	BusinessId string   `json:"business_id" form:"business_id" binding:"required"`
	Username   string   `json:"username" form:"username" binding:"required"`
	BetAmount  int64    `json:"bet_amount"`
	DrawId     int64    `json:"draw_id"`
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

func (c *SettleOrder) GetBetAmountOnly() int64 {
	return util.MoneyInt(float64(c.BetAmount))
}

func (c *SettleOrder) GetWagerMultiplier() (value int64, exists bool) {
	return -1, true
}

func (c *SettleOrder) GetBetAmount() (amount int64, exists bool) {
	if *c.Amount == 0 { // lost bets
		return c.BetAmount, true
	}
	return util.MoneyInt(*c.Amount) + c.BetAmount, true // won bets; c.amount + c.BetAmount because the equation in processTransaction is betAmount - winAmount
}

type PlaceResponse struct {
	Game serializer.StreamGameSession `json:"draw"`
	Ts   int64                        `json:"ts"`
}

func Place(c *gin.Context, req PlaceOrder) (res serializer.Response, err error) {
	go common.LogGameCallbackRequest("stream_game_place_order", req)
	i18n := c.MustGet("i18n").(i18n.I18n)
	user := c.MustGet("user").(model.User)
	var draw ploutos.StreamGameSession
	err = model.DB.Preload(`StreamGame`).Where(`id`, req.DrawId).Where(`status`, ploutos.StreamGameSessionStatusOpen).First(&draw).Error
	if err != nil {
		res = serializer.ParamErr(c, req, i18n.T("invalid_stream_game_draw_id"), err)
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
		Data: PlaceResponse{
			Game: serializer.BuildStreamGameSession(c, draw),
			Ts:   time.Now().Unix(),
		},
	}
	return
}

func Settle(c *gin.Context, req SettleOrder) (res serializer.Response, err error) {
	go common.LogGameCallbackRequest("stream_game_settle_order", req)
	var br ploutos.StreamGameBetReport
	err = model.DB.Where(`business_id`, req.BusinessId).Where(`status`, 4).First(&br).Error // 4: unsettled
	if err != nil {
		res = serializer.Err(c, req, serializer.CodeGeneralError, "stream game settle error", err)
		return
	}
	req.DrawId = br.GameId
	req.BetAmount = br.Bet
	if os.Getenv("PRODUCT") == "batace" {
		err = common.ProcessTransactionBatace(&req)
	} else {
		err = common.ProcessTransaction(&req)
	}
	if err != nil {
		res = serializer.Err(c, req, serializer.CodeGeneralError, "stream game settle error", err)
		return
	}
	res = serializer.Response{
		Msg: "Success",
	}
	return
}
