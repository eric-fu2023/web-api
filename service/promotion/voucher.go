package promotion

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"github.com/go-redsync/redsync/v4"
	"gorm.io/gorm"
)

const (
	userVoucherBindingKey = "user_voucher_binding_lock:%d:%d"
)

var (
	ErrVoucherUseInvalid = errors.New("invalid use of voucher")
)

type VoucherList struct {
	Filter string `form:"filter" json:"filter"`
}

func (v VoucherList) Handle(c *gin.Context) (r serializer.Response, err error) {
	now := time.Now()
	// brand := c.MustGet(`_brand`).(int)
	user := c.MustGet("user").(model.User)
	deviceInfo, _ := util.GetDeviceInfo(c)

	list, err := model.VoucherListUsableByUserFilter(c, user.ID, v.Filter, now)
	if err != nil {
		r = serializer.Err(c, v, serializer.CodeGeneralError, "", err)
		return
	}
	r.Data = util.MapSlice(list, func(a models.Voucher) serializer.Voucher {
		return serializer.BuildVoucher(a, deviceInfo.Platform)
	})
	return
}

type VoucherApplicable struct {
	Type              string          `json:"type"`
	TransactionDetail json.RawMessage `json:"transaction_detail"`
}

type FBTransactionDetail struct {
	MatchID     int64   `json:"matchId"`
	OptionType  int     `json:"optionType"`
	OddsFormat  int     `json:"oddsFormat"`
	Odds        float64 `json:"odds"`
	MarketID    int64   `json:"marketId"`
	Stake       float64 `json:"stake"`
	SportID     int     `json:"sportId"`
	MatchStatus int     `json:"matchStatus"`
	IsOutright  bool    `json:"isOutright"`
	IsParlay    bool    `json:"isParlay"`
}

func (v VoucherApplicable) GetFBTransactionDetail() (ret FBTransactionDetail) {
	_ = json.Unmarshal(v.TransactionDetail, &ret)
	return
}

func (v VoucherApplicable) Handle(c *gin.Context) (r serializer.Response, err error) {
	now := time.Now()
	deviceInfo, _ := util.GetDeviceInfo(c)

	// brand := c.MustGet(`_brand`).(int)
	user := c.MustGet("user").(model.User)
	var (
		matchType  int
		odds       float64
		betAmount  int64
		oddsFormat int
	)
	switch v.Type {
	case "fb":
		d := v.GetFBTransactionDetail()
		matchType = d.MatchStatus
		odds = d.Odds
		betAmount = int64(math.Round(d.Stake * 100))
		oddsFormat = d.OddsFormat
	}

	list, err := model.VoucherListUsableByUserFilter(c, user.ID, "valid", now)
	if err != nil {
		r = serializer.Err(c, v, serializer.CodeGeneralError, "", err)
		return
	}

	ret := []models.Voucher{}
	for _, voucher := range list {
		if ValidateVoucherUsageByType(voucher, oddsFormat, matchType, odds, betAmount) {
			ret = append(ret, voucher)
		}
	}
	r.Data = util.MapSlice(ret, func(a models.Voucher) serializer.Voucher {
		return serializer.BuildVoucher(a, deviceInfo.Platform)
	})
	return
}

type VoucherPreBinding struct {
	Type              string          `json:"type"`
	TransactionDetail json.RawMessage `json:"transaction_detail"`
	VoucherID         int64           `json:"voucher_id"`
}

func (v VoucherPreBinding) GetFBTransactionDetail() (ret FBTransactionDetail) {
	_ = json.Unmarshal(v.TransactionDetail, &ret)
	return
}

func (v VoucherPreBinding) Handle(c *gin.Context) (r serializer.Response, err error) {
	now := time.Now()
	// brand := c.MustGet(`_brand`).(int)
	user := c.MustGet("user").(model.User)
	mutex := cache.RedisLockClient.NewMutex(fmt.Sprintf(userVoucherBindingKey, user.ID, v.VoucherID), redsync.WithExpiry(5*time.Second))
	mutex.Lock()
	defer mutex.Unlock()
	var (
		matchType  int
		odds       float64
		betAmount  int64
		oddsFormat int
	)
	switch v.Type {
	case "fb":
		d := v.GetFBTransactionDetail()
		matchType = d.MatchStatus
		odds = d.Odds
		betAmount = int64(math.Round(d.Stake * 100))
		oddsFormat = d.OddsFormat
	}
	err = model.DB.WithContext(c).Transaction(func(tx *gorm.DB) error {
		voucher, err := model.VoucherActiveGetByIDUserWithDB(c, user.ID, v.VoucherID, now, tx)
		if err != nil {
			return err
		}
		if !ValidateVoucherUsageByType(voucher, oddsFormat, matchType, odds, betAmount) {
			err = ErrVoucherUseInvalid
			r = serializer.Err(c, v, serializer.CodeGeneralError, "Invalid use of voucher", err)
			return err
		}

		err = tx.WithContext(c).Model(&models.Voucher{}).Where("id", voucher.ID).Updates(
			map[string]any{
				"status":              models.VoucherStatusPending,
				"transaction_details": v,
			},
		).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		r = serializer.EnsureErr(c, err, r)
		return
	}
	r.Data = "success"
	return
}
