package mumbai

import (
	"errors"
	"fmt"
	"strconv"

	"web-api/model"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/mumbai/api"
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

	username := c.Merchant + c.Agent + fmt.Sprintf("%08s", user.IdAsString())

	// FIXME
	// may need to encode
	transactionNo := c.Merchant + defaultTransactionNum

	res, err := client.WithdrawUser(username, transactionNo, defaultTransferAmount)
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

var ResponseCodeTransactionNumberExistError = errors.New("transaction number exist")

func (c *Mumbai) TransferTo(tx *gorm.DB, user model.User, sum ploutos.UserSum, _currency, gameCode string, gameVendorId int64, extra model.Extra) (_transferredBalance int64, _err error) {
	switch {
	case sum.Balance == 0:
		return 0, nil
	case sum.Balance < 0:
		return 0, ResponseCodeTransactionNumberExistError
	}

	client, err := util.MumbaiFactory()
	if err != nil {
		return 0, err
	}

	// Assuming uaerName, transactionNo, and money are defined and initialized somewhere
	uaerName := "example_username"
	transactionNo := "example_transaction_no"
	money := 100 // example value

	// Convert money to string
	moneyStr := strconv.Itoa(money)

	ptxid, err := client.DepositUser(uaerName, transactionNo, moneyStr)
	if err != nil {
		return 0, err
	}

	if ptxid.Code != "0" { // Assuming ptxid has a field Code indicating success or failure
		return 0, fmt.Errorf("deposit failed with code: %s, message: %s", ptxid.Code, ptxid.Message)
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
