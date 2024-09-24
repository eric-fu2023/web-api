package ninewicket

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"web-api/model"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/ninewickets"
	"blgit.rfdev.tech/taya/game-service/ninewickets/api"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (n *NineWicket) CreateWallet(user model.User, currency string) error {
	return model.DB.Transaction(func(tx *gorm.DB) (err error) {
		var gameVendors []ploutos.GameVendor
		err = tx.Model(ploutos.GameVendor{}).Joins(`INNER JOIN game_vendor_brand gvb ON gvb.game_vendor_id = game_vendor.id`).
			Where(`game_vendor.game_integration_id`, util.IntegrationIdNineWicket).Find(&gameVendors).Error
		if err != nil {
			return
		}
		for _, gameVendor := range gameVendors {
			gvu := ploutos.GameVendorUser{
				GameVendorId:     gameVendor.ID,
				UserId:           user.ID,
				ExternalUserId:   user.Username,
				ExternalCurrency: currency,
			}
			err = tx.Create(&gvu).Error
			if err != nil {
				return
			}
		}
		return
	})
}

func (n *NineWicket) TransferTo(tx *gorm.DB, user model.User, sum ploutos.UserSum, currency, gameCode string, gameVendorId int64, extra model.Extra) (balance int64, err error) {
	switch {
	case sum.Balance == 0:
		return 0, nil
	case sum.Balance < 0:
		return 0, errors.New("9Wicket::TransferTo not allowed to transfer negative sum")
	}

	client, err := util.NineWicketFactory()
	if err != nil {
		return 0, err
	}

	tsCode, err := client.Deposit(api.UserId(user.ID), util.MoneyFloat(sum.Balance))
	util.Log().Info("9Wicket GAME INTEGRATION TRANSFER IN game_integration_id: %d, user_id: %d, balance: %.4f, tx_id: %s", util.IntegrationIdNineWicket, user.IdAsString(), util.MoneyFloat(sum.Balance), tsCode)

	//go handleFailedTransaction(userId, userId+currentTimeMillisString)
	transaction := ploutos.Transaction{
		UserId:                user.ID,
		Amount:                -1 * sum.Balance,
		BalanceBefore:         sum.Balance,
		BalanceAfter:          0,
		TransactionType:       ploutos.TransactionTypeToGameIntegration,
		Wager:                 0,
		WagerBefore:           sum.RemainingWager,
		WagerAfter:            sum.RemainingWager,
		ExternalTransactionId: tsCode,
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

	return sum.Balance, nil
}

func (n *NineWicket) TransferFrom(ctx context.Context, tx *gorm.DB, user model.User, currency string, gameCode string, gameVendorId int64, extra model.Extra) (err error) {
	client, err := util.NineWicketFactory()
	if err != nil {
		return err
	}

	userBalance, err := client.GetBalanceOneUser(api.UserId(user.ID))
	if err != nil {
		util.Log().Info("Error getting 9Wicket user balance,userID: %v ,err: %v ", user.IdAsString(), err.Error())
		return
	}

	if userBalance <= 0 {
		util.Log().Info("This user balance is smaller than / equal to 0, user: %v, balance: %v", user.IdAsString(), userBalance)
		return
	}

	resp, err := client.Withdraw(api.UserId(user.ID), api.WithdrawOptions{Withdraw: 1})

	if err != nil {
		util.Log().Info("Error transfer 9Wicket user balance from 9Wicket error,userID: %v ,err: %v ", user.IdAsString(), err.Error())
		go handleFailedTransaction(tx, user, userBalance, resp.TxId, gameVendorId, client)
		return
	}

	util.Log().Info("9Wicket GAME INTEGRATION TRANSFER OUT game_integration_id: %d, user_id: %d, balance: %.4f, remaining balance: %.4f, tx_id: %s", util.IntegrationIdNineWicket, user.IdAsString(), resp.Withdrawn, resp.Remaining, resp.TxId)
	err = updateUserBalance(tx, user, userBalance, resp.TxId, gameVendorId)
	return err
}

func handleFailedTransaction(tx *gorm.DB, user model.User, userBalance float64, TransID string, gameVendorId int64, client ninewickets.ClientOperations) {
	var count = 0
	for {
		res, err := client.CheckTransferRecord(api.UserId(user.ID), TransID)
		if err != nil {
			util.Log().Info("Error fetching transaction details from 9Wicket, err: %v", err)
			return
		}

		if count == 3 {
			break
		}
		// Check the condition
		if res.Status == "1" {
			// Condition is true, do something
			if res.Result[TransID].Status == "1012" {
				util.Log().Info("9Wicket GAME TRANSACTION DETAIL IN game_integration_id: %d, user_id: %s, balance: %.4f, status: %s, tx_id: %s", util.IntegrationIdNineWicket, api.UserId(user.ID), res.Result[TransID].Balance, res.Result[TransID].Status, res.Result[TransID].TsCode)
				err = updateUserBalance(tx, user, userBalance, TransID, gameVendorId)
				if err != nil {
					log.Printf("Error updating user balance, err: %v", err)
				}
			} else {
				time.Sleep(10 * time.Second)
				util.Log().Info("9Wicket GAME TRANSACTION DETAIL IN game_integration_id: %d, user_id: %s, status: %s, tx_id: %s", util.IntegrationIdNineWicket, api.UserId(user.ID), res.Result[TransID].Status, TransID)
				count++
				continue
			}
			//Condition met, proceeding with loop.
		} else {
			// Condition is false, wait for 10 seconds
			fmt.Println("Condition not met, waiting for 10 seconds.")
			util.Log().Info("9Wicket GAME TRANSACTION DETAIL IN game_integration_id: %d, status: %s", util.IntegrationIdNineWicket, api.UserId(user.ID), res.Status)

			time.Sleep(10 * time.Second)
			// Continue the loop after waiting
			count++
			continue
		}
		break
	}
}

func (n *NineWicket) GetGameBalance(user model.User, currency, gameCode string, extra model.Extra) (balance int64, _err error) {
	client, err := util.NineWicketFactory()
	if err != nil {
		return 0, err
	}
	userBalance, err := client.GetBalanceOneUser(api.UserId(user.ID))
	if err != nil {
		util.Log().Info("Error getting 9wicket user balance,userID: %v ,err: %v ", user.IdAsString(), err.Error())
		return
	}

	if userBalance <= 0 {
		util.Log().Info("This user balance is smaller than / equal to 0, user: %v, balance: %v", user.IdAsString(), userBalance)
		return
	}

	return util.MoneyInt(userBalance), nil
}

func updateUserBalance(tx *gorm.DB, user model.User, TBalance float64, transID string, gameVendorId int64) error {
	var sum ploutos.UserSum
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", user.ID).First(&sum).Error
	if err != nil {
		log.Printf("Error fetching user balance, err: %v", err)
		return err
	}

	amount := util.MoneyInt(TBalance)
	transaction := ploutos.Transaction{
		UserId:                user.ID,
		Amount:                amount,
		BalanceBefore:         sum.Balance,
		BalanceAfter:          sum.Balance + amount,
		TransactionType:       ploutos.TransactionTypeFromGameIntegration,
		Wager:                 0,
		WagerBefore:           sum.RemainingWager,
		WagerAfter:            sum.RemainingWager,
		ExternalTransactionId: transID,
		GameVendorId:          gameVendorId,
	}
	err = tx.Create(&transaction).Error
	if err != nil {
		log.Printf("Error creating transaction, err: %v", err)
		return err
	}

	err = tx.Model(ploutos.UserSum{}).Where("user_id = ?", user.ID).Update("balance", gorm.Expr("balance + ?", amount)).Error
	if err != nil {
		log.Printf("Error updating user balance, err: %v", err)
		return err
	}

	return nil
}
