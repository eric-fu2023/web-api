package imone

import (
	"web-api/model"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TransferFrom
func (c *ImOne) TransferFrom(tx *gorm.DB, user model.User, currency, gameCode string, gameVendorId int64, extra model.Extra) error {
	client := util.ImOneFactory()

	productWallet, exist := tayaGameCodeToImOneWalletCodeMapping[gameCode]
	if !exist {
		return ErrGameCodeMapping
	}
	balance, err := client.GetWalletBalance(user.IdAsString(), productWallet)

	if err != nil {
		return err
	}

	switch {
	case balance == 0:
		return nil
	case balance < 0:
		return ErrInsufficientImoneWalletBalance
	}

	now, err := util.NowGMT8()
	if err != nil {
		return err
	}

	ptxid, err := client.PerformTransfer(user.IdAsString(), productWallet, -1*balance, now)
	if err != nil {
		return err
	}

	var sum ploutos.UserSum
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(`user_id`, user.ID).First(&sum).Error; err != nil {
		return err
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
		return err
	}
	err = tx.Model(ploutos.UserSum{}).Where(`user_id`, user.ID).Update(`balance`, gorm.Expr(`balance + ?`, amount)).Error
	return err
}

func (c *ImOne) TransferTo(tx *gorm.DB, user model.User, sum ploutos.UserSum, _currency, gameCode string, gameVendorId int64, extra model.Extra) (_transferredBalance int64, _err error) {
	switch {
	case sum.Balance == 0:
		return 0, nil
	case sum.Balance < 0:
		return 0, ErrTransferNegativeBalance
	}

	productWallet, exist := tayaGameCodeToImOneWalletCodeMapping[gameCode]
	if !exist {
		return 0, ErrGameCodeMapping
	}

	client := util.ImOneFactory()

	now, err := util.NowGMT8()
	if err != nil {
		return 0, err
	}

	ptxid, err := client.PerformTransfer(user.IdAsString(), productWallet, util.MoneyFloat(sum.Balance), now)
	if err != nil {
		return 0, err
	}

	if ptxid != "" {
		transaction := ploutos.Transaction{
			UserId:                user.ID,
			Amount:                -1 * sum.Balance,
			BalanceBefore:         sum.Balance,
			BalanceAfter:          0,
			TransactionType:       ploutos.TransactionTypeToGameIntegration, /*ploutos.TransactionTypeToUGS*/
			Wager:                 0,
			WagerBefore:           sum.RemainingWager,
			WagerAfter:            sum.RemainingWager,
			ExternalTransactionId: ptxid,
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
