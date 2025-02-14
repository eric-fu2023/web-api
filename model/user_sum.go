package model

import (
	"context"
	"log"

	"blgit.rfdev.tech/taya/common-function/rfcontext"

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

func UpdateDbUserSumAndCreateTransaction(ctx context.Context, txDB *gorm.DB, userId, amount, wagerChange, withdrawableChange, transactionType int64, cashOrderId string) (sum UserSum, err error) {
	ctx = rfcontext.AppendCallDesc(ctx, "UpdateDbUserSumAndCreateTransaction")
	ctx = rfcontext.AppendParams(ctx, "", map[string]interface{}{
		"userId":             userId,
		"amount":             amount,
		"wagerChange":        wagerChange,
		"withdrawableChange": withdrawableChange,
		"transactionType":    transactionType,
		"cashOrderId":        cashOrderId,
	})

	{
		log.Printf(rfcontext.Fmt(ctx))
	}

	err = txDB.Clauses(dbresolver.Use("txConn")).Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id", userId).First(&sum).Error
		if err != nil {
			return
		}

		// only update deposit wager when it is a deposit transaction or make up deposit transaction
		deposit_wager_change := int64(0)
		if transactionType == ploutos.TransactionTypeCashIn || transactionType == ploutos.TransactionTypeMakeUpCashOrder {
			deposit_wager_change = wagerChange
		}

		transaction := Transaction{
			ploutos.Transaction{
				UserId:             userId,
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
				CashOrderID:        cashOrderId,
			},
		}

		{
			ctx = rfcontext.AppendParams(ctx, "before save transaction", map[string]interface{}{
				"transaction": transaction,
			})
			log.Printf(rfcontext.Fmt(ctx))
		}

		err = tx.Create(&transaction).Error
		if err != nil {
			return
		}
		sum.Balance += amount
		sum.RemainingWager += wagerChange
		sum.DepositRemainingWager += deposit_wager_change
		sum.MaxWithdrawable += withdrawableChange

		{
			ctx = rfcontext.AppendParams(ctx, "before save user sum", map[string]interface{}{
				"sum": sum,
			})
			log.Printf(rfcontext.Fmt(ctx))
		}

		err = tx.Save(&sum).Error
		if err != nil {
			return
		}
		return
	})
	return
}
