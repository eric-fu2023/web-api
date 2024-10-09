package crown_valexy

import (
	"context"
	"errors"
	"strconv"
	"time"

	"web-api/model"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/crownvalexy"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CrownValexy struct{}

func (c *CrownValexy) CreateWallet(ctx context.Context, user model.User, s string) error {
	go func() {
		// fire and forget. later calls should follow up with user creation, if needed.
		service, err := util.CrownValexyFactory(ctx)
		if err == nil {
			_, _ = service.Login(ctx, user.IdAsString())
		}
	}()

	return model.DB.Transaction(func(tx *gorm.DB) (err error) {
		var gameVendors []ploutos.GameVendor
		err = tx.Model(ploutos.GameVendor{}).
			Where(`game_vendor.game_integration_id`, util.IntegrationIdCrownValexy).Find(&gameVendors).Error
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
}

type Remarks struct {
	WithdrawId string `json:"withdraw_id"`
	DepositId  string `json:"transfer_id"`
}

func (r Remarks) String() string {
	if r.WithdrawId != "" {
		return r.WithdrawId
	}

	if r.DepositId != "" {
		return r.DepositId
	}

	return ""
}

func (r Remarks) MarshalJSON() ([]byte, error) {
	if r.WithdrawId != "" {
		return []byte(r.WithdrawId), nil
	}

	if r.DepositId != "" {
		return []byte(r.DepositId), nil
	}

	return nil, nil
}

func (c *CrownValexy) TransferFrom(ctx context.Context, tx *gorm.DB, user model.User, _curr string, _gameVendorCode string, gameVendorId int64, extra model.Extra) error {
	userId := user.IdAsString()
	client, err := util.CrownValexyFactory(ctx)
	if err != nil {
		return err
	}
	cvUser, err := client.UserDetails(ctx, userId)
	if err != nil {
		if errors.Is(err, crownvalexy.ErrAccountInvalid) {
			return nil
		}
		return err
	}
	toWithdraw := cvUser.Balance

	if toWithdraw == 0 {
		return nil
	}

	withdrawTxId := userId + strconv.FormatInt(time.Now().UnixNano(), 10) + "withdraw"

	_, wErr := client.WalletWithdraw(ctx, user.IdAsString(), toWithdraw, Remarks{
		WithdrawId: withdrawTxId,
		DepositId:  "",
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

func (c *CrownValexy) TransferTo(ctx context.Context, tx *gorm.DB, user model.User, sum ploutos.UserSum, s string, s2 string, gameVendorId int64, extra model.Extra) (int64, error) {
	switch {
	case sum.Balance == 0:
		return 0, nil
	case sum.Balance < 0:
		return 0, errors.New("rf user balance is not positive")
	}

	client, err := util.CrownValexyFactory(context.TODO())
	if err != nil {
		return 0, err
	}

	userId := user.IdAsString()
	depTxId := userId + strconv.FormatInt(time.Now().UnixNano(), 10) + "deposit"
	_, wErr := client.WalletDeposit(ctx, userId, util.MoneyFloat(sum.Balance), Remarks{
		WithdrawId: "",
		DepositId:  depTxId,
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
	client, err := util.CrownValexyFactory(ctx)
	if err != nil {
		return "", err
	}

	url, rErr := client.Login(ctx, user.IdAsString())
	if rErr != nil {
		return "", rErr
	}

	return url, nil
}

func (c *CrownValexy) GetGameBalance(ctx context.Context, user model.User, s string, s2 string, extra model.Extra) (int64, error) {
	userId := user.IdAsString()

	client, err := util.CrownValexyFactory(ctx)
	if err != nil {
		return 0, err
	}

	cvUser, err := client.UserDetails(ctx, userId)
	if err != nil {
		return 0, err
	}

	return util.MoneyInt(cvUser.Balance), nil
}
