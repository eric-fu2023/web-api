package evo

import (
	"errors"
	"os"
	"strconv"
	"time"
	"web-api/model"
	"web-api/util"

	"log"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EVO struct {
}

func (e EVO) CreateWallet(user model.User, currency string) (err error) {
	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		var gameVendors []ploutos.GameVendor
		err = tx.Model(ploutos.GameVendor{}).Joins(`INNER JOIN game_vendor_brand gvb ON gvb.game_vendor_id = game_vendor.id`).
			Where(`game_vendor.game_integration_id`, util.IntegrationIdEvo).Find(&gameVendors).Error
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

func (e EVO) GetGameUrl(user model.User, currency, gameCode, subGameCode string, platform int64, extra model.Extra) (url string, err error) {
	client := util.EvoFactory.NewClient()

	uuid := uuid.NewString()
	currentTimeMillis := time.Now().UnixNano() / int64(time.Millisecond)
	currentTimeMillisString := strconv.FormatInt(currentTimeMillis, 10)
	response, err := client.GetGameUrl(uuid, extra.Locale, user.IdAsString(), currency, user.IdAsString()+"_"+currentTimeMillisString, extra.Ip, subGameCode)

	if err != nil {
		log.Printf("Error getting evo game url, err: %v ", err.Error())
	}
	url = os.Getenv("GAME_EVO_HOST") + response.EntryEmebedded

	return url, err
}

func (e EVO) GetGameBalance(user model.User, currency, gameCode string, extra model.Extra) (balance int64, err error) {
	return 0, nil
}

func (e EVO) TransferFrom(tx *gorm.DB, user model.User, currency, gameCode string, gameVendorId int64, extra model.Extra) (err error) {
	client := util.EvoFactory.NewClient()

	userBalance, err := client.GetGameBalance(user.IdAsString())

	if err != nil {
		log.Printf("Error getting evo user balance,userID: %v ,err: %v ", user.IdAsString(), err.Error())
		return
	}

	if userBalance.TBalance <= 0 {
		log.Printf("This user balance is smaller than / equal to 0, user: %v, balance: %v", user.IdAsString(), userBalance.TBalance)
		return
	}
	currentTimeMillis := time.Now().UnixNano() / int64(time.Millisecond)
	currentTimeMillisString := strconv.FormatInt(currentTimeMillis, 10)

	res, err := client.TransferOut(user.IdAsString(), userBalance.TBalance, user.IdAsString()+"_"+currentTimeMillisString)
	util.Log().Info("EVO GAME INTEGRATION TRANSFER OUT game_integration_id: %d, user_id: %d, balance: %.4f, status: %d, tx_id: %s", util.IntegrationIdEvo, user.ID, res.Balance, res.Result, res.TransID)

	if err != nil {
		log.Printf("Error transfer evo user balance from error,userID: %v ,err: %v ", user.IdAsString(), err.Error())
		return
	}

	if res.Result == "N" {
		log.Printf("Error transfer evo user balance,userID: %v ,err: %v ", user.IdAsString(), res.ErrorMsg)
		// need to call another routine to fetch transactions details.
		go handleFailedTransaction( tx, user, userBalance.TBalance, res.TransID, gameVendorId)
		return
	}
	err = updateUserBalance(tx, user, userBalance.TBalance, res.TransID, gameVendorId);
	if err != nil {
        return err
    }
    return nil
}

func handleFailedTransaction(tx *gorm.DB, user model.User, userBalance float64, TransID string, gameVendorId int64) {
	// sleep 2 seconds to archive 99.9% accuracy
	time.Sleep(2 * time.Second)
	client := util.EvoFactory.NewClient()
    transaction, err := client.Transactions(user.IdAsString(), TransID)
    if err != nil {
        log.Printf("Error fetching transaction details from EVO, err: %v", err)
        return
    }
    log.Printf("Transaction details: %v", transaction)

    if transaction.Result == "Y" {
		// if transaction success, we need to deposit user balance for the transaction amount in this transaction response!
		// NOT THE AMOUNT FROM INITIAL REQUEST, just to keep all consistent with EVO system
        err = updateUserBalance(tx, user, transaction.Amount, TransID, gameVendorId)
        if err != nil {
            log.Printf("Error updating user balance, err: %v", err)
        }
    }
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


func (e EVO) TransferTo(tx *gorm.DB, user model.User, sum ploutos.UserSum, currency, gameCode string, gameVendorId int64, extra model.Extra) (balance int64, err error) {
	switch {
	case sum.Balance == 0:
		return 0, nil
	case sum.Balance < 0:
		return 0, errors.New("Evo::TransferTo not allowed to transfer negative sum")
	}

	client := util.EvoFactory.NewClient()

	currentTimeMillis := time.Now().UnixNano() / int64(time.Millisecond)
	currentTimeMillisString := strconv.FormatInt(currentTimeMillis, 10)

	res, err := client.TransferIn(true, user.IdAsString(), util.MoneyFloat(sum.Balance), user.IdAsString()+"_"+currentTimeMillisString)
	util.Log().Info("EVO GAME INTEGRATION TRANSFER IN game_integration_id: %d, user_id: %d, balance: %.4f, status: %d, tx_id: %s", util.IntegrationIdEvo, user.ID, util.MoneyFloat(sum.Balance), res.Result, res.TransID)

	if res.Result == "N" {
		log.Printf("Error transfer evo user balance,userID: %v ,err: %v ", user.IdAsString(), err.Error())
		return 0, err
	}

	transaction := ploutos.Transaction{
		UserId:                user.ID,
		Amount:                -1 * sum.Balance,
		BalanceBefore:         sum.Balance,
		BalanceAfter:          0,
		TransactionType:       ploutos.TransactionTypeToGameIntegration,
		Wager:                 0,
		WagerBefore:           sum.RemainingWager,
		WagerAfter:            sum.RemainingWager,
		ExternalTransactionId: res.TransID,
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
