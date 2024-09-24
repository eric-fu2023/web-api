package ugs

import (
	"context"

	"web-api/cache"
	"web-api/model"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

type UGS struct{}

func (c UGS) CreateWallet(ctx context.Context, user model.User, currency string) (err error) {
	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		var gameVendors []ploutos.GameVendor
		err = tx.Model(ploutos.GameVendor{}).Joins(`INNER JOIN game_vendor_brand gvb ON gvb.game_vendor_id = game_vendor.id`).
			Where(`game_vendor.game_integration_id`, util.IntegrationIdUGS).Find(&gameVendors).Error
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

func (c UGS) TransferFrom(ctx context.Context, tx *gorm.DB, user model.User, currency string, gameCode string, gameVendorId int64, extra model.Extra) (err error) {
	var isTestUser bool
	if user.Role == 2 {
		isTestUser = true
	}
	client := util.UgsFactory.NewClient(cache.RedisClient)
	balance, status, ptxid, err := client.TransferOut(user.ID, user.Username, currency, extra.Locale, gameCode, extra.Ip, isTestUser)
	if err != nil {
		return
	}
	util.Log().Info("UGS GAME INTEGRATION TRANSFER OUT game_integration_id: %d, user_id: %d, balance: %.4f, status: %d, tx_id: %s", util.IntegrationIdUGS, user.ID, balance, status, ptxid)
	if status == TransferStatusSuccess && balance > 0 && ptxid != "" {
		var sum ploutos.UserSum
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(`user_id`, user.ID).First(&sum).Error
		if err != nil {
			return
		}
		amount := util.MoneyInt(balance)
		transaction := ploutos.Transaction{
			UserId:                user.ID,
			Amount:                amount,
			BalanceBefore:         sum.Balance,
			BalanceAfter:          sum.Balance + amount,
			TransactionType:       ploutos.TransactionTypeFromGameIntegration,
			Wager:                 0,
			WagerBefore:           sum.RemainingWager,
			WagerAfter:            sum.RemainingWager,
			ExternalTransactionId: ptxid,
			GameVendorId:          gameVendorId,
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

func (c UGS) TransferTo(ctx context.Context, tx *gorm.DB, user model.User, sum ploutos.UserSum, currency string, gameCode string, gameVendorId int64, extra model.Extra) (balance int64, err error) {
	var isTestUser bool
	if user.Role == 2 {
		isTestUser = true
	}
	client := util.UgsFactory.NewClient(cache.RedisClient)
	status, ptxid, err := client.TransferIn(user.ID, user.Username, currency, extra.Locale, gameCode, extra.Ip, isTestUser, util.MoneyFloat(sum.Balance))
	if err != nil {
		return
	}
	util.Log().Info("UGS GAME INTEGRATION TRANSFER IN game_integration_id: %d, user_id: %d, balance: %.4f, status: %d, tx_id: %s", util.IntegrationIdUGS, user.ID, util.MoneyFloat(sum.Balance), status, ptxid)
	if status == TransferStatusSuccess && ptxid != "" {
		transaction := ploutos.Transaction{
			UserId:                user.ID,
			Amount:                -1 * sum.Balance,
			BalanceBefore:         sum.Balance,
			BalanceAfter:          0,
			TransactionType:       ploutos.TransactionTypeToGameIntegration,
			Wager:                 0,
			WagerBefore:           sum.RemainingWager,
			WagerAfter:            sum.RemainingWager,
			ExternalTransactionId: ptxid,
			GameVendorId:          gameVendorId,
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

func (c UGS) GetGameUrl(ctx context.Context, user model.User, currency, gameCode, subGameCode string, platform int64, extra model.Extra) (url string, err error) {
	var isTestUser bool
	if user.Role == 2 {
		isTestUser = true
	}
	client := util.UgsFactory.NewClient(cache.RedisClient)
	url, err = client.GetGameUrl(user.ID, user.Username, currency, extra.Locale, gameCode, subGameCode, extra.Ip, PlatformMapping[platform], isTestUser)
	return
}

func (c UGS) GetGameBalance(ctx context.Context, user model.User, currency string, gameCode string, extra model.Extra) (balance int64, err error) {
	var isTestUser bool
	if user.Role == 2 {
		isTestUser = true
	}
	client := util.UgsFactory.NewClient(cache.RedisClient)
	balanceFloat, err := client.GetGameBalance(user.ID, user.Username, currency, extra.Locale, gameCode, extra.Ip, isTestUser)
	if err != nil {
		return
	}
	balance = util.MoneyInt(balanceFloat)
	return
}
