package model

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
)

type UserSum struct {
	ploutos.UserSum
}

func GetByUserIDWithLockWithDB(userID int64, tx *gorm.DB) (sum UserSum, err error) {
	err = tx.Debug().Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id", userID).First(&sum).Error
	return
}

func UpdateDbUserSumAndCreateTransaction(txDB *gorm.DB, userID, amount, wagerChange, withdrawableChange, transactionType int64, cashOrderID string) (sum UserSum, err error) {
	err = txDB.Clauses(dbresolver.Use("txConn")).Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id", userID).First(&sum).Error
		if err != nil {
			return
		}

		// only update deposit wager when it is a deposit transaction or make up deposit transaction
		deposit_wager_change := int64(0)
		if transactionType == models.TransactionTypeCashIn || transactionType == models.TransactionTypeMakeUpCashOrder{
			deposit_wager_change = wagerChange
		}
		
		transaction := Transaction{
			ploutos.Transaction{
				UserId:             userID,
				Amount:             amount,
				BalanceBefore:      sum.Balance,
				BalanceAfter:       sum.Balance + amount,
				TransactionType:    transactionType,
				Wager:              wagerChange,
				WagerBefore:        sum.RemainingWager,
				WagerAfter:         sum.RemainingWager + wagerChange,
				DepositWager:       deposit_wager_change,
				DepositWagerBefore: sum.DepositRemainingWager,
				DepositWagerAfter:  sum.DepositRemainingWager + deposit_wager_change,
				CashOrderID:        cashOrderID,
			},
		}
		err = tx.Create(&transaction).Error
		if err != nil {
			return
		}
		sum.Balance += amount
		sum.RemainingWager += wagerChange
		sum.DepositRemainingWager += deposit_wager_change
		sum.MaxWithdrawable += withdrawableChange
		err = tx.Save(&sum).Error
		if err != nil {
			return
		}
		return
	})
	return
}
