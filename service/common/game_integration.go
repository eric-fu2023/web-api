package common

import (
	"context"
	"errors"
	"gorm.io/gorm/clause"
	"os"
	"strconv"
	"time"

	"web-api/model"
	"web-api/service/evo"
	"web-api/service/imone"
	"web-api/service/mumbai"
	"web-api/service/ninewicket"
	"web-api/service/ugs"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/mancala/api"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

var GameIntegration = map[int64]GameIntegrationInterface{
	util.IntegrationIdUGS:        ugs.UGS{},
	util.IntegrationIdImOne:      &imone.ImOne{},
	util.IntegrationIdEvo:        evo.EVO{},
	util.IntegrationIdNineWicket: &ninewicket.NineWicket{},
	util.IntegrationIdMumbai: &mumbai.Mumbai{
		Merchant: os.Getenv("GAME_MUMBAI_MERCHANT_CODE"),
		Agent:    os.Getenv("GAME_MUMBAI_AGENT_CODE"),
	},

	// TODO
	//util.IntegrationIdCrownValexy: &CrownValexy{},
	util.IntegrationIdMancala: &Mancala{},
}

type CrownValexy struct{}

func (c *CrownValexy) CreateWallet(user model.User, s string) error {
	//TODO implement me
	return errors.New("todo")
}

func (c *CrownValexy) TransferFrom(db *gorm.DB, user model.User, s string, s2 string, i int64, extra model.Extra) error {
	//TODO implement me
	return errors.New("todo")
}

func (c *CrownValexy) TransferTo(db *gorm.DB, user model.User, sum ploutos.UserSum, s string, s2 string, i int64, extra model.Extra) (int64, error) {
	//TODO implement me
	return 0, errors.New("todo")
}

func (c *CrownValexy) GetGameUrl(user model.User, s string, s2 string, s3 string, i int64, extra model.Extra) (string, error) {
	//TODO implement me
	return "", errors.New("todo")
}

func (c *CrownValexy) GetGameBalance(user model.User, s string, s2 string, extra model.Extra) (int64, error) {
	//TODO implement me
	return 0, errors.New("todo")
}

type Mancala struct{}

var TayaCurrencyToMancalaCurrency = map[string]api.Currency{
	"INR": "INR",
}

func (m *Mancala) CreateWallet(user model.User, tayaCurrency string) error {
	//TODO implement me
	currency, ok := TayaCurrencyToMancalaCurrency[tayaCurrency]
	if !ok {
		return errors.New("mancala unknown currency mapping")
	}

	// FIXME password to be derived from user instead of default value
	go func() {
		// fire and forget. later calls should follow up with user creation, if needed.
		service, err := util.MancalaFactory()
		if err == nil {
			_, _, _ = service.AddTransferWallet(context.TODO(), user.IdAsString(), currency)
		}
	}()

	return model.DB.Transaction(func(tx *gorm.DB) (err error) {
		var gameVendors []ploutos.GameVendor
		err = tx.Model(ploutos.GameVendor{}).
			Where(`game_vendor.game_integration_id`, util.IntegrationIdMancala).Find(&gameVendors).Error
		if err != nil {
			return
		}

		for _, gameVendor := range gameVendors {
			gvu := ploutos.GameVendorUser{
				GameVendorId:     gameVendor.ID,
				UserId:           user.ID,
				ExternalUserId:   user.IdAsString(),
				ExternalCurrency: currency,
			}

			err = tx.Create(&gvu).Error
			if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
				return
			}
		}
		return
	})
}

func (m *Mancala) TransferFrom(tx *gorm.DB, user model.User, tayaCurrency, gameCode string, gameVendorId int64, extra model.Extra) error {
	currency, ok := TayaCurrencyToMancalaCurrency[tayaCurrency]
	if !ok {
		return errors.New("mancala unknown currency mapping")
	}
	userId := user.IdAsString()

	//TODO implement me
	client, err := util.MancalaFactory()
	if err != nil {
		return err
	}
	ctx := context.Background()

	balanceResponse, err := client.GetBalance(ctx, userId, currency)
	if err != nil {
		return err
	}
	toWithdraw := balanceResponse.Balance
	switch {
	case toWithdraw == 0:
		return nil
	case toWithdraw < 0:
		return errors.New("manacala user balance is not positive")
	}

	withdrawTxId := userId + strconv.FormatInt(time.Now().UnixNano(), 10)

	_, wErr := client.Withdraw(ctx, userId, currency, withdrawTxId, toWithdraw)
	if wErr != nil {
		return wErr
	}

	_withdrawn := toWithdraw
	var sum ploutos.UserSum
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(`user_id`, user.ID).First(&sum).Error; err != nil {
		return err
	}
	withdrawn := util.MoneyInt(_withdrawn)
	transaction := ploutos.Transaction{
		UserId:                user.ID,
		Amount:                withdrawn,
		BalanceBefore:         sum.Balance,
		BalanceAfter:          sum.Balance + withdrawn,
		TransactionType:       ploutos.TransactionTypeFromGameIntegration,
		Wager:                 0,
		WagerBefore:           sum.RemainingWager,
		WagerAfter:            sum.RemainingWager,
		ExternalTransactionId: withdrawTxId,
		GameVendorId:          gameVendorId,
	}
	err = tx.Create(&transaction).Error
	if err != nil {
		return err
	}
	err = tx.Model(ploutos.UserSum{}).Where(`user_id`, user.ID).Update(`balance`, gorm.Expr(`balance + ?`, withdrawn)).Error
	return err
}

func (m *Mancala) TransferTo(db *gorm.DB, user model.User, sum ploutos.UserSum, s string, s2 string, i int64, extra model.Extra) (int64, error) {
	//TODO implement me
	return 0, errors.New("todo")
}

func (m *Mancala) GetGameUrl(user model.User, s string, s2 string, s3 string, i int64, extra model.Extra) (string, error) {
	//TODO implement me
	return "", errors.New("todo")
}

func (m *Mancala) GetGameBalance(user model.User, s string, s2 string, extra model.Extra) (int64, error) {
	//TODO implement me
	return 0, errors.New("todo")
}

type GameIntegrationInterface interface {
	CreateWallet(model.User, string) error
	TransferFrom(*gorm.DB, model.User, string, string, int64, model.Extra) error
	TransferTo(*gorm.DB, model.User, ploutos.UserSum, string, string, int64, model.Extra) (int64, error)
	GetGameUrl(model.User, string, string, string, int64, model.Extra) (string, error)
	GetGameBalance(model.User, string, string, model.Extra) (int64, error)
}
