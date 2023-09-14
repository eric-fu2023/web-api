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
	Transaction struct {
		models.TransactionC
	}
)

func (UserSum) GetByIDWithLockWithDB(id int64, tx *gorm.DB) (sum UserSum, err error) {
	err = DB.First(&sum, id).Error
	return
}

func (UserSum) UpdateUserSumWithDB(txDB *gorm.DB, userID, amount, wager, transactionType, transactionID int64) (sum UserSum, err error) {
	err = txDB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id", userID).First(&sum).Error
		if err != nil {
			return
		}
		transaction := Transaction{
			models.TransactionC{
				UserId:            userID,
				Amount:            amount,
				BalanceBefore:     sum.Balance,
				BalanceAfter:      sum.Balance + amount,
				GameTransactionId: transactionID,
				Gametype:          transactionType,
				Wager:             wager,
				WagerBefore:       sum.RemainingWager,
				WagerAfter:        sum.RemainingWager + wager,
			},
		}
		err = tx.Create(&transaction).Error
		if err != nil {
			return
		}
		sum.Balance += amount
		sum.RemainingWager += wager
		err = tx.Save(&sum).Error
		if err != nil {
			return
		}
		return
	})
	return
}
