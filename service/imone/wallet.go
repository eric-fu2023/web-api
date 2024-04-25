package imone

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"web-api/model"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

// TransferFrom
func (c *ImOne) TransferFrom(tx *gorm.DB, user model.User, currency, gameCode string, gameVendorId int64, _ model.Extra) error {
	client := util.ImOneFactory()

	productWallet := tayaGameCodeToImOneWalletCodeMapping[gameCode]

	balance, err := client.GetWalletBalance(user.IdAsString(), productWallet)
	if err != nil {
		return err
	}

	ptxid, err := client.PerformTransfer(user.IdAsString(), productWallet, -1*balance, time.Now())
	if err != nil {
		return err
	}

	util.Log().Info("ImOne GAME INTEGRATION TRANSFER OUT game_integration_id: %d, user_id: %d, balance: %.4f, tx_id: %s", util.IntegrationIdImOne, user.ID, balance, ptxid)
	if balance > 0 && ptxid != "" {
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
		if err != nil {
			return err
		}
	}
	return err
}

func (c *ImOne) TransferTo(tx *gorm.DB, user model.User, sum ploutos.UserSum, _currency, gameCode string, gameVendorId int64, _ model.Extra) (_transferredBalance int64, _err error) {
	if sum.Balance < 0 {
		return 0, errors.New("ImOne::TransferTo negative balance")
	}
	productWallet := tayaGameCodeToImOneWalletCodeMapping[gameCode]

	client := util.ImOneFactory()
	ptxid, err := client.PerformTransfer(user.IdAsString(), productWallet, util.MoneyFloat(sum.Balance), time.Now())
	if err != nil {
		return 0, err
	}
	util.Log().Info("ImOne GAME INTEGRATION TRANSFER IN game_integration_id: %d, user_id: %d, balance: %.4f, status: %d, tx_id: %s", util.IntegrationIdImOne, user.ID, util.MoneyFloat(sum.Balance), ptxid)
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
