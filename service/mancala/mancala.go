package mancala

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"web-api/model"
	"web-api/util"

	"blgit.rfdev.tech/taya/common-function/rfcontext"
	"blgit.rfdev.tech/taya/game-service/mancala"
	"blgit.rfdev.tech/taya/game-service/mancala/api"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Mancala struct{}

var TayaCurrencyToMancalaCurrency = map[string]api.Currency{
	"INR": "INR",
}

func (m *Mancala) CreateWallet(ctx context.Context, user model.User, tayaCurrency string) error {
	currency, ok := TayaCurrencyToMancalaCurrency[tayaCurrency]
	if !ok {
		return errors.New("mancala unknown currency mapping")
	}

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

func (m *Mancala) TransferFrom(ctx context.Context, tx *gorm.DB, user model.User, tayaCurrency string, gameCode string, gameVendorId int64, extra model.Extra) error {
	currency, ok := TayaCurrencyToMancalaCurrency[tayaCurrency]
	if !ok {
		return errors.New("mancala unknown currency mapping")
	}
	userId := user.IdAsString()

	client, err := util.MancalaFactory()
	if err != nil {
		return err
	}

	balanceResponse, err := client.GetBalance(ctx, userId, currency)
	if err != nil {
		if errors.Is(err, mancala.ErrNotFound) {
			return nil
		}
		return err
	}
	toWithdraw := balanceResponse.Balance
	switch {
	case toWithdraw == 0:
		return nil
	case toWithdraw < 0:
		return errors.New("mancala user balance is not positive")
	}

	withdrawTxId := userId + strconv.FormatInt(time.Now().UnixNano(), 10) + "withdraw"

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

func (m *Mancala) TransferTo(ctx context.Context, tx *gorm.DB, user model.User, sum ploutos.UserSum, tayaCurrency string, gameCode string, gameVendorId int64, extra model.Extra) (_transferredBalance int64, _err error) {
	switch {
	case sum.Balance == 0:
		return 0, nil
	case sum.Balance < 0:
		return 0, errors.New("rf user balance is not positive")
	}

	client, err := util.MancalaFactory()
	if err != nil {
		return 0, err
	}
	currency, ok := TayaCurrencyToMancalaCurrency[tayaCurrency]
	if !ok {
		return 0, errors.New("mancala unknown currency mapping")
	}

	userId := user.IdAsString()

	depTxId := userId + strconv.FormatInt(time.Now().UnixNano(), 10) + "deposit"
	_, wErr := client.Deposit(ctx, userId, currency, depTxId, util.MoneyFloat(sum.Balance))
	if wErr != nil {
		return 0, wErr
	}

	if depTxId != "" {
		transaction := ploutos.Transaction{
			UserId:                user.ID,
			Amount:                -1 * sum.Balance,
			BalanceBefore:         sum.Balance,
			BalanceAfter:          0,
			TransactionType:       ploutos.TransactionTypeToGameIntegration,
			Wager:                 0,
			WagerBefore:           sum.RemainingWager,
			WagerAfter:            sum.RemainingWager,
			ExternalTransactionId: depTxId,
			GameVendorId:          gameVendorId,
		}
		err = tx.Create(&transaction).Error
		if err != nil {
			return 0, err
		}
		err = tx.Model(ploutos.UserSum{}).Where(`user_id`, user.ID).Update(`balance`, 0).Error
		if err != nil {
			return 0, err
		}
	}
	return sum.Balance, nil
}

func (m *Mancala) GetGameUrl(ctx context.Context, user model.User, tayaCurrency, _, subGameCode string, _ int64, extra model.Extra) (string, error) {
	ctx = rfcontext.AppendParams(ctx, "GetGameUrl", map[string]interface{}{
		"user":         user,
		"subGameCode":  subGameCode,
		"tayaCurrency": tayaCurrency,
	})
	userId := user.IdAsString()
	currency, ok := TayaCurrencyToMancalaCurrency[tayaCurrency]
	if !ok {
		return "", errors.New("mancala unknown currency mapping")
	}

	client, err := util.MancalaFactory()
	if err != nil {
		return "", err
	}

	_, exists, aErr := client.AddTransferWallet(ctx, user.IdAsString(), TayaCurrencyToMancalaCurrency[tayaCurrency])
	if aErr != nil {
		ctx = rfcontext.AppendError(ctx, aErr, "AddTransferWallet.a")
		log.Printf(rfcontext.Fmt(ctx))
		return "", aErr
	}

	if !exists {
		_err := errors.New("mancala account missing after create if not exists")
		ctx = rfcontext.AppendError(ctx, aErr, "AddTransferWallet.e")
		log.Printf(rfcontext.Fmt(ctx))
		return "", _err
	}

	gameId, err := strconv.Atoi(subGameCode)
	if err != nil {
		return "", err
	}

	resp, gErr := client.GetToken(ctx, userId, int64(gameId), currency, api.LanguageGB)
	if gErr != nil {
		return "", gErr
	}

	url := resp.IframeUrl
	if url == "" {
		return "", errors.New("no err but frame url empty")
	}
	return url, nil
}

func (m *Mancala) GetGameBalance(ctx context.Context, user model.User, tayaCurrency string, _ string, extra model.Extra) (int64, error) {
	currency, ok := TayaCurrencyToMancalaCurrency[tayaCurrency]
	if !ok {
		return 0, errors.New("mancala unknown currency mapping")
	}
	userId := user.IdAsString()

	client, err := util.MancalaFactory()
	if err != nil {
		return 0, err
	}
	balanceResponse, err := client.GetBalance(ctx, userId, currency)
	if err != nil {
		return 0, err
	}
	// isit necessary to convert?
	return util.MoneyInt(balanceResponse.Balance), err
}
