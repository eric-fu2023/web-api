package mumbai

import (
	"fmt"
	"strconv"
	"web-api/model"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CURRENTLY USING HARDCODED VALUE. NOTE: WILL NEED TO CHANGE TO USE USER INPUT.
const (
	defaultTransactionNum = "12345678"
	defaultTransferAmount = "50"
)

func (c *Mumbai) TransferFrom(tx *gorm.DB, user model.User, currency, gameCode string, gameVendorId int64, extra model.Extra) error {
	// get the balance from mumbai and update the db
	client, err := util.MumbaiFactory()
	if err != nil {
		return err
	}

	username := c.Merchant + c.Agent + user.IdAsString()
	transactionNo := c.Merchant + defaultTransactionNum

	res, err := client.WithdrawUser(username, transactionNo, defaultTransferAmount)

	if err != nil {
		if err.Error() == string(ResponseCodeNotEnoughFundsError) {
			return ErrInsufficientMumbaiWalletBalance
		} else {
			return err
		}
	}

	// if no error means we have sufficient funds to withdraw from mumbai server
	// now we will need to update the balance in db to match with the withdraw amount.
	// parse the money from string -> float64 -> int64
	money, err := strconv.ParseFloat(res.Result.Money, 64)
	moneyToInsertDB := util.MoneyInt(money)

	if err != nil {
		return err
	}

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
	return 0, fmt.Errorf("not implemented")
}
