package crown_valexy

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"web-api/model"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CrownValexy struct{}

func (c *CrownValexy) CreateWallet(user model.User, s string) error {
	//TODO implement me
	go func() {
		// fire and forget. later calls should follow up with user creation, if needed.
		service, err := util.CrownValexyFactory()
		if err == nil {
			_, _ = service.Login(context.TODO(), user.IdAsString())
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
				GameVendorId:   gameVendor.ID,
				UserId:         user.ID,
				ExternalUserId: user.IdAsString(),
			}

			err = tx.Create(&gvu).Error
			if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
				return
			}
		}
		return
	})
	return errors.New("todo")
}

type Remarks struct {
	WithdrawId string `json:"withdraw_id"`
	TransferId string `json:"transfer_id"`
}

func (r Remarks) String() string {
	bb, _ := json.Marshal(r)
	return string(bb)
}

func (c *CrownValexy) TransferFrom(tx *gorm.DB, user model.User, _curr string, _gameVendorCode string, gameVendorId int64, extra model.Extra) error {
	ctx := context.TODO()
	userId := user.IdAsString()
	client, err := util.CrownValexyFactory()
	if err != nil {
		return err
	}
	cvUser, err := client.UserDetails(ctx, userId)
	if err != nil {
		return err
	}
	toWithdraw := cvUser.Balance
	withdrawTxId := userId + strconv.FormatInt(time.Now().UnixNano(), 10) + "withdraw"

	_, wErr := client.WalletWithdraw(ctx, user.IdAsString(), toWithdraw, Remarks{
		WithdrawId: withdrawTxId,
		TransferId: "",
	}.String())
	if wErr != nil {
		return wErr
	}

	_withdrawn := toWithdraw
	var sum ploutos.UserSum
	if uErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(`user_id`, user.ID).First(&sum).Error; uErr != nil {
		return uErr
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
	tErr := tx.Create(&transaction).Error
	if tErr != nil {
		return tErr
	}
	err = tx.Model(ploutos.UserSum{}).Where(`user_id`, user.ID).Update(`balance`, gorm.Expr(`balance + ?`, withdrawn)).Error
	return err
}

func (c *CrownValexy) TransferTo(tx *gorm.DB, user model.User, sum ploutos.UserSum, s string, s2 string, gameVendorId int64, extra model.Extra) (int64, error) {
	switch {
	case sum.Balance == 0:
		return 0, nil
	case sum.Balance < 0:
		return 0, errors.New("rf user balance is not positive")
	}

	client, err := util.CrownValexyFactory()
	if err != nil {
		return 0, err
	}

	ctx := context.Background()
	userId := user.IdAsString()
	depTxId := userId + strconv.FormatInt(time.Now().UnixNano(), 10) + "deposit"
	_, wErr := client.WalletDeposit(ctx, userId, util.MoneyFloat(sum.Balance), Remarks{
		WithdrawId: "",
		TransferId: depTxId,
	}.String())
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

func (c *CrownValexy) GetGameUrl(ctx context.Context, user model.User, s string, s2 string, s3 string, i int64, extra model.Extra) (string, error) {
	client, err := util.CrownValexyFactory()
	if err != nil {
		return "", err
	}

	url, rErr := client.Login(ctx, user.IdAsString())
	if rErr != nil {
		return "", rErr
	}

	return url, nil
}

func (c *CrownValexy) GetGameBalance(user model.User, s string, s2 string, extra model.Extra) (int64, error) {
	ctx := context.Background()
	userId := user.IdAsString()

	client, err := util.CrownValexyFactory()
	if err != nil {
		return 0, err
	}

	cvUser, err := client.UserDetails(ctx, userId)
	if err != nil {
		return 0, err
	}

	return util.MoneyInt(cvUser.Balance), nil
}
