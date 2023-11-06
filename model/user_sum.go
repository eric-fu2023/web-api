package model

import (
	models "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	UserSum struct {
		models.UserSumC
	}
)

func (UserSum) GetByUserIDWithLockWithDB(userID int64, tx *gorm.DB) (sum UserSum, err error) {
	err = DB.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id", userID).First(&sum).Error
	return
}

func (UserSum) UpdateUserSumWithDB(txDB *gorm.DB, userID, amount, wagerChange, withdrawableChange, transactionType int64, cashOrderID string) (sum UserSum, err error) {
	err = txDB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id", userID).First(&sum).Error
		if err != nil {
			return
		}
		transaction := Transaction{
			models.TransactionC{
				UserId:          userID,
				Amount:          amount,
				BalanceBefore:   sum.Balance,
				BalanceAfter:    sum.Balance + amount,
				TransactionType: transactionType,
				Wager:           wagerChange,
				WagerBefore:     sum.RemainingWager,
				WagerAfter:      sum.RemainingWager + wagerChange,
				CashOrderID:     cashOrderID,
			},
		}
		err = tx.Create(&transaction).Error
		if err != nil {
			return
		}
		sum.Balance += amount
		sum.RemainingWager += wagerChange
		sum.MaxWithdrawable += withdrawableChange
		err = tx.Save(&sum).Error
		if err != nil {
			return
		}
		return
	})
	return
}
