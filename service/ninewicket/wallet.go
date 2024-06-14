package ninewicket

import (
	"errors"
	"log"

	"web-api/model"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/ninewickets/api"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (n *NineWicket) CreateWallet(user model.User, currency string) (err error) {
	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
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
	if err != nil {
		return
	}

	return
}

func (n *NineWicket) TransferTo(tx *gorm.DB, user model.User, sum ploutos.UserSum, currency, gameCode string, gameVendorId int64, extra model.Extra) (balance int64, err error) {
	switch {
	case sum.Balance == 0:
		return 0, nil
	case sum.Balance < 0:
		return 0, errors.New("Evo::TransferTo not allowed to transfer negative sum")
	}

	client := util.NineWicketFactory()

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

func (n *NineWicket) TransferFrom(tx *gorm.DB, user model.User, currency, gameCode string, gameVendorId int64, extra model.Extra) (err error) {
	client := util.NineWicketFactory()

	userBalance, err := client.GetBalanceOneUser(api.UserId(user.ID))

	if err != nil {
		util.Log().Info("Error getting evo user balance,userID: %v ,err: %v ", user.IdAsString(), err.Error())
		return
	}

	if userBalance <= 0 {
		util.Log().Info("This user balance is smaller than / equal to 0, user: %v, balance: %v", user.IdAsString(), userBalance)
		return
	}

	//resp := ninewicket.WithdrawResponse{}
	resp, err := client.Withdraw(api.UserId(user.ID), api.WithdrawOptions{Withdraw: 1})
	util.Log().Info("9Wicket GAME INTEGRATION TRANSFER OUT game_integration_id: %d, user_id: %d, balance: %.4f, remaining balance: %.4f, tx_id: %s", util.IntegrationIdNineWicket, user.IdAsString(), resp.Withdrawn, resp.Remaining, resp.TxId)

	//res, err = client.CheckTransferRecord(userId, userId+currentTimeMillisString)
	//util.Log().Info("9Wicket GAME INTEGRATION TRANSFER IN game_integration_id: %d, user_id: %s, balance: %.4f, status: %s, tx_id: %s", util.IntegrationIdNineWicket, userId, res.Result[userId+currentTimeMillisString].Balance, res.Result[userId+currentTimeMillisString].Status, res.Result[userId+currentTimeMillisString].TsCode)
	//go handleFailedTransaction(userId, userId+currentTimeMillisString)
	if err != nil {
		util.Log().Info("Error transfer evo user balance from error,userID: %v ,err: %v ", user.IdAsString(), err.Error())
		return
	}

	//if res.Result == "N" {
	//	log.Printf("Error transfer evo user balance,userID: %v ,err: %v ", user.IdAsString(), res.ErrorMsg)
	//	// need to call another routine to fetch transactions details.
	//	go handleFailedTransaction(tx, user, userBalance.TBalance, user.IdAsString()+"_"+currentTimeMillisString, gameVendorId)
	//	return
	//}
	err = updateUserBalance(tx, user, userBalance, resp.TxId, gameVendorId)
	if err != nil {
		return err
	}
	return nil
}

// Function that returns a boolean value
func checkCondition() bool {
	// Your condition logic here
	// Returning false as an example

	return false
}

// func handleFailedTransaction(tx *gorm.DB, user model.User, userBalance float64, TransID string, gameVendorId int64) {
//func handleFailedTransaction(userId string, tsCode string) {
//	client := util.NineWicketFactory()
//
//	for {
//		res, err := client.CheckTransferRecord(userId, tsCode)
//
//		if err != nil {
//			util.Log().Info("Error fetching transaction details from 9Wicket, err: %v", err)
//			return
//		}
//		// Check the condition
//		if res.Status == "1" {
//			// Condition is true, do something
//			util.Log().Info("9Wicket GAME TRANSACTION DETAIL IN game_integration_id: %d, user_id: %s, balance: %.4f, status: %s, tx_id: %s", util.IntegrationIdNineWicket, userId, res.Result[tsCode].Balance, res.Result[tsCode].Status, res.Result[tsCode].TsCode)
//
//			fmt.Println("Condition met, proceeding with loop.")
//		} else {
//			// Condition is false, wait for 10 seconds
//			fmt.Println("Condition not met, waiting for 10 seconds.")
//			util.Log().Info("9Wicket GAME TRANSACTION DETAIL IN game_integration_id: %d, status: %s", util.IntegrationIdNineWicket, userId, res.Status)
//
//			time.Sleep(10 * time.Second)
//			// Continue the loop after waiting
//			continue
//		}
//
//		break
//	}

// sleep 2 seconds to archive 99.9% accuracy
//time.Sleep(2 * time.Second)
//client := util.EvoFactory.NewClient()
//transaction, err := client.Transactions(user.IdAsString(), TransID)
//if err != nil {
//	log.Printf("Error fetching transaction details from EVO, err: %v", err)
//	return
//}
//log.Printf("Transaction details: %v", transaction)
//
//if transaction.Result == "Y" {
//	// if transaction success, we need to deposit user balance for the transaction amount in this transaction response!
//	// NOT THE AMOUNT FROM INITIAL REQUEST, just to keep all consistent with EVO system
//	err = updateUserBalance(tx, user, transaction.Amount, TransID, gameVendorId)
//	if err != nil {
//		log.Printf("Error updating user balance, err: %v", err)
//	}
//}

//}

//func (n *NineWicket) GetGameBalance(userId string) (balance float64, err error) {
//	client := util.NineWicketFactory()
//
//	res, err := client.GetGameBalance(userId)
//	if err != nil {
//		return
//	}
//
//	balance = res.Results[0].Balance
//	//balance = util.MoneyInt(balanceFloat)
//	return balance, err
//}

func (n *NineWicket) GetGameBalance(user model.User, currency, gameCode string, extra model.Extra) (balance int64, _err error) {
	client := util.NineWicketFactory()

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
