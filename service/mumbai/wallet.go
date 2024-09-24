package mumbai

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"web-api/model"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/mumbai/api"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CURRENTLY USING HARDCODED VALUE. NOTE: WILL NEED TO CHANGE TO USE USER INPUT.
const (
	defaultTransferAmount = "50"
)

func generateTransactionNo(user model.User, mode api.TransferCheckType) string {
	return os.Getenv("GAME_MUMBAI_MERCHANT_CODE") + string(mode) + user.IdAsString() + fmt.Sprintf("x%x", time.Now().Unix())
}
func (c *Mumbai) TransferFrom(ctx context.Context, tx *gorm.DB, user model.User, currency string, gameCode string, gameVendorId int64, extra model.Extra) error {
	// get the balance from mumbai and update the db
	client, err := util.MumbaiFactory()
	if err != nil {
		return err
	}

	username := os.Getenv("GAME_MUMBAI_MERCHANT_CODE") + os.Getenv("GAME_MUMBAI_AGENT_CODE") + fmt.Sprintf("%08s", user.IdAsString())
	transactionNo := generateTransactionNo(user, api.WithdrawCheckType)
	mbBalance, err := client.CheckBalanceUser(username)
	log.Printf("(c *Mumbai) TransferFrom balance %v err %v\n", mbBalance, err)
	if err != nil {
		return err
	}
	if mbBalance == 0 {
		return nil
	}
	withdraw, err := client.WithdrawUser(username, transactionNo, fmt.Sprintf("%.2f", mbBalance))
	if err != nil {
		if err.Error() == string(api.ResponseCodeNotEnoughFundsError) {
			return ErrInsufficientMumbaiWalletBalance
		} else {
			return err
		}
	}

	// if no error means we have sufficient funds to withdraw from mumbai server
	// now we will need to update the balance in db to match with the withdraw amount.
	// parse the money from string -> float64 -> int64
	moneyToInsertDB := util.MoneyInt(withdraw)
	var sum ploutos.UserSum
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(`user_id`, user.ID).First(&sum).Error; err != nil {
		return err
	}

	transaction := ploutos.Transaction{
		UserId:                user.ID,
		Amount:                moneyToInsertDB,
		BalanceBefore:         sum.Balance,
		BalanceAfter:          sum.Balance + moneyToInsertDB,
		TransactionType:       ploutos.TransactionTypeFromGameIntegration,
		Wager:                 0,
		WagerBefore:           sum.RemainingWager,
		WagerAfter:            sum.RemainingWager,
		ExternalTransactionId: "", // empty string for now
		GameVendorId:          gameVendorId,
	}

	err = tx.Create(&transaction).Error
	if err != nil {
		return err
	}

	err = tx.Model(ploutos.UserSum{}).Where(`user_id`, user.ID).Update(`balance`, gorm.Expr(`balance + ?`, moneyToInsertDB)).Error
	return err
}

func (c *Mumbai) TransferTo(tx *gorm.DB, user model.User, sum ploutos.UserSum, _currency, gameCode string, gameVendorId int64, extra model.Extra) (_transferredBalance int64, _err error) {
	if sum.Balance == 0 {
		return 0, nil
	}

	if sum.Balance < 0 {
		return 0, ErrInsufficientUserWalletBalance
	}

	client, err := util.MumbaiFactory()
	if err != nil {
		return 0, err
	}

	// Assuming uaerName, transactionNo, and money are defined and initialized somewhere
	username := os.Getenv("GAME_MUMBAI_MERCHANT_CODE") + os.Getenv("GAME_MUMBAI_AGENT_CODE") + fmt.Sprintf("%08s", user.IdAsString())

	// Convert money to string
	moneyStr := fmt.Sprintf("%.2f", util.MoneyFloat(sum.Balance))
	transactionNo := generateTransactionNo(user, api.DepositCheckType)

	_, err = client.DepositUser(username, transactionNo, moneyStr)
	if err != nil {
		return 0, err
	}

	transaction := ploutos.Transaction{
		UserId:          user.ID,
		Amount:          -1 * sum.Balance,
		BalanceBefore:   sum.Balance,
		BalanceAfter:    0,
		TransactionType: ploutos.TransactionTypeToGameIntegration,
		Wager:           0,
		WagerBefore:     sum.RemainingWager,
		WagerAfter:      sum.RemainingWager,
		GameVendorId:    gameVendorId,
	}

	err = tx.Create(&transaction).Error
	if err != nil {
		return 0, err
	}

	err = tx.Model(ploutos.UserSum{}).Where(`user_id = ?`, user.ID).UpdateColumn(`balance`, 0).Error
	if err != nil {
		return 0, err
	}
	return sum.Balance, nil
}
