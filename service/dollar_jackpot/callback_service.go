package dollar_jackpot

import (
	"blgit.rfdev.tech/taya/game-service/dc/callback"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
	"web-api/model"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"
)

type Callback struct {
	GameId int64 `json:"game_id" form:"game_id" binding:"required"`
}

func (c *Callback) NewCallback(userId int64) {}

func (c *Callback) GetGameVendorId() int64 {
	return c.GameId
}

func (c *Callback) GetGameTransactionId() int64 {
	return 0
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
	DrawId int64       `json:"draw_id" form:"draw_id" binding:"required"`
	Amount float64     `json:"amount" form:"amount" binding:"required"`
	User   *model.User `json:"user"`
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
	}
	return model.DB.Omit("id").Create(&betReport).Error
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
	DrawId          int64   `json:"draw_id" form:"draw_id" binding:"required"`
	Username        string  `json:"username" form:"username" binding:"required"`
	Amount          float64 `json:"amount" form:"amount" binding:"required"`
	WagerMultiplier int64   `json:"wager_multiplier" form:"wager_multiplier" binding:"required"`
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
	return 0, true
}

func Place(c *gin.Context, req PlaceOrder) (res callback.BaseResponse, err error) {
	go common.LogGameCallbackRequest("dollar_jackpot_place_order", req)
	i18n := c.MustGet("i18n").(i18n.I18n)
	user := c.MustGet("user").(model.User)
	req.User = &user
	err = common.ProcessTransaction(&req)
	if err != nil {
		return
	}
	res = callback.BaseResponse{
		Msg: i18n.T("success"),
	}
	return
}
