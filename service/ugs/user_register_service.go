package ugs

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"web-api/cache"
	"web-api/model"
	"web-api/util"
)

const IntegrationIdUGS = 1
const (
	TransferStatusFailed  = 0
	TransferStatusSuccess = 1
	TransferStatusPending = 2
)

var PlatformMapping = map[int64]int64{
	1: 1, // pc
	2: 2, // h5
	3: 2, // android
	4: 2, // ios
}

type UGS struct {
}

func (c UGS) CreateWallet(user model.User, currency string) (err error) {
	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		var gameVendors []ploutos.GameVendor
		err = tx.Model(ploutos.GameVendor{}).Joins(`INNER JOIN game_vendor_brand gvb ON gvb.game_vendor_id = game_vendor.id AND gvb.brand_id = ?`, user.BrandId).
			Where(`game_vendor.game_integration_id`, IntegrationIdUGS).Find(&gameVendors).Error
		if err != nil {
			return
		}
		for _, gameVendor := range gameVendors {
			gvu := ploutos.GameVendorUser{
				GameVendorId:     gameVendor.ID,
				UserId:           user.ID,
				ExternalCurrency: currency,
			}
			err = tx.Create(&gvu).Error
			if err != nil {
				return
			}
		}
		return
	})
	if err != nil {
		return
	}
	return
}

func (c UGS) TransferFrom(tx *gorm.DB, user model.User, currency, lang, gameCode, ip string) (err error) {
	var isTestUser bool
	if user.Role == 2 {
		isTestUser = true
	}
	client := util.UgsFactory.NewClient(cache.RedisClient)
	balance, status, ptxid, err := client.TransferOut(user.ID, user.Username, currency, lang, gameCode, ip, isTestUser)
	if err != nil {
		return
	}
	util.Log().Info("GAME INTEGRATION TRANSFER OUT game_integration_id: %d, user_id: %d, balance: %.4f, status: %d, tx_id: %s", IntegrationIdUGS, user.ID, balance, status, ptxid)
	var sum ploutos.UserSum
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(`user_id`, user.ID).First(&sum).Error
	if err != nil {
		return
	}
	if status == TransferStatusSuccess && balance > 0 && ptxid != "" {
		amount := util.MoneyInt(balance)
		transaction := ploutos.Transaction{
			UserId:                user.ID,
			Amount:                amount,
			BalanceBefore:         sum.Balance,
			BalanceAfter:          sum.Balance + amount,
			TransactionType:       ploutos.TransactionTypeFromUGS,
			Wager:                 0,
			WagerBefore:           sum.RemainingWager,
			WagerAfter:            sum.RemainingWager,
			ExternalTransactionId: ptxid,
		}
		err = tx.Create(&transaction).Error
		if err != nil {
			return
		}
		err = tx.Model(ploutos.UserSum{}).Where(`user_id`, user.ID).Update(`balance`, gorm.Expr(`balance + ?`, amount)).Error
		if err != nil {
			return
		}
	}
	return
}

func (c UGS) TransferTo(tx *gorm.DB, user model.User, sum ploutos.UserSum, currency, lang, gameCode, ip string) (balance int64, err error) {
	var isTestUser bool
	if user.Role == 2 {
		isTestUser = true
	}
	client := util.UgsFactory.NewClient(cache.RedisClient)
	status, ptxid, err := client.TransferIn(user.ID, user.Username, currency, lang, gameCode, ip, isTestUser, util.MoneyFloat(sum.Balance))
	if err != nil {
		return
	}
	util.Log().Info("GAME INTEGRATION TRANSFER IN game_integration_id: %d, user_id: %d, balance: %.4f, status: %d, tx_id: %s", IntegrationIdUGS, user.ID, util.MoneyFloat(sum.Balance), status, ptxid)
	if status == TransferStatusSuccess && ptxid != "" {
		transaction := ploutos.Transaction{
			UserId:                user.ID,
			Amount:                -1 * sum.Balance,
			BalanceBefore:         sum.Balance,
			BalanceAfter:          0,
			TransactionType:       ploutos.TransactionTypeToUGS,
			Wager:                 0,
			WagerBefore:           sum.RemainingWager,
			WagerAfter:            sum.RemainingWager,
			ExternalTransactionId: ptxid,
		}
		err = tx.Create(&transaction).Error
		if err != nil {
			return
		}
		err = tx.Model(ploutos.UserSum{}).Where(`user_id`, user.ID).Update(`balance`, 0).Error
		if err != nil {
			return
		}
	}
	balance = sum.Balance
	return
}

func (c UGS) GetGameUrl(user model.User, currency, lang, gameCode, subGameCode, ip string, platform int64) (url string, err error) {
	var isTestUser bool
	if user.Role == 2 {
		isTestUser = true
	}
	client := util.UgsFactory.NewClient(cache.RedisClient)
	url, err = client.GetGameUrl(user.ID, user.Username, currency, lang, gameCode, subGameCode, ip, PlatformMapping[platform], isTestUser)
	return
}
