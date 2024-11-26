package png

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	game_service_png "blgit.rfdev.tech/taya/game-service/png"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/go-redis/redis"

	"web-api/cache"
	"web-api/model"
	"web-api/util"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PNG struct {}

const (
	CLIENT_SESSION_REDIS_KEY = "png_tickets"
)
func (p PNG) CreateWallet(ctx context.Context, user model.User, currency string) (err error) {
	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		var gameVendors []ploutos.GameVendor
		err = tx.Model(ploutos.GameVendor{}).Joins(`INNER JOIN game_vendor_brand gvb ON gvb.game_vendor_id = game_vendor.id`).
			Where(`game_vendor.game_integration_id`, util.IntegrationIPNG).Find(&gameVendors).Error
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

func (p PNG) GetGameUrl(ctx context.Context, user model.User, currency, gameCode, subGameCode string, platform int64, extra model.Extra) (url string, err error) {
	png_service := game_service_png.PNG{}
	practice := 0
	game_name := subGameCode
	channel := "mobile"
	lang := "en"
	ticket:=""
	origin:="batce999.com"

	// get ticket
	res := cache.RedisSessionClient.Get(context.Background(), fmt.Sprintf("%s/%s",CLIENT_SESSION_REDIS_KEY,strconv.FormatInt(user.ID, 10)))
	if res.Err() == redis.Nil {
		// get ticket
		ticket, get_ticket_err :=png_service.GetTicket(os.Getenv("GAME_PNG_HOST"),"GetTicket", user.ID)
		if get_ticket_err.Error() == "UnknownUser" {
			png_service.Register(os.Getenv("GAME_PNG_HOST"), "RegisterUser",user.ID,user.Username,user.Nickname,user.CreatedAt.Format("2006-01-02"))
			ticket, get_ticket_err = png_service.GetTicket(os.Getenv("GAME_PNG_HOST"),"GetTicket", user.ID)
			if get_ticket_err!=nil{
				return "", get_ticket_err
			}
		}else if get_ticket_err!=nil{
			return "", get_ticket_err
		}

		// if get ticket api success, set token in redis.   here we do not really care if the set failed or not
		cache.RedisSessionClient.Set(context.Background(), fmt.Sprintf("%s/%s",CLIENT_SESSION_REDIS_KEY,strconv.FormatInt(user.ID, 10)), ticket, 1*time.Hour)
		
		// if set_res.Err() != nil && res.Err() != redis.Nil {
		// 	return "", set_res.Err()
		// }
	}
	// get game name
	return png_service.GetGameUrl(game_name, channel, lang, practice, ticket, origin), nil
}

func (p PNG) GetGameBalance(ctx context.Context, user model.User, currency string, gameCode string, extra model.Extra) (balance int64, err error) {
	return 0, nil
}

func (p PNG) TransferFrom(ctx context.Context, tx *gorm.DB, user model.User, _ string, _ string, gameVendorId int64, extra model.Extra) (err error) {
	png_service := game_service_png.PNG{}
	userBalance, err := png_service.GetBalance(os.Getenv(os.Getenv("GAME_PNG_HOST")),"Balance",user.ID)
	if err != nil {
		log.Printf("Error getting PNG user balance,userID: %v ,err: %v ", user.IdAsString(), err.Error())
		return
	}

	if userBalance <= 0 {
		log.Printf("This user balance is smaller than / equal to 0, user: %v, balance: %v", user.IdAsString(), userBalance)
		return
	}

	balance, tx_id, err := png_service.TransferOut(os.Getenv("GAME_PNG_HOST"),"",user.ID, userBalance, "")

	util.Log().Info("PNG GAME INTEGRATION TRANSFER OUT game_integration_id: %d, user_id: %d, balance: %.4f, tx_id: %s", util.IntegrationIdEvo, user.ID, balance, tx_id)
	if err != nil {
		log.Printf("Error transfer png user balance from error,userID: %v ,err: %v ", user.IdAsString(), err.Error())
		return
	}

	err = updateUserBalance(tx, user, balance, tx_id, gameVendorId)
	if err != nil {
		return err
	}
	return nil
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
	db_tx:=tx.Begin()

	defer db_tx.Rollback()
	err = tx.Create(&transaction).Error
	if err != nil {
		defer db_tx.Rollback()
		log.Printf("Error creating transaction, err: %v", err)
		return err
	}

	err = tx.Model(ploutos.UserSum{}).Where("user_id = ?", user.ID).Update("balance", gorm.Expr("balance + ?", amount)).Error
	if err != nil {
		defer db_tx.Rollback()
		log.Printf("Error updating user balance, err: %v", err)
		return err
	}
	db_tx.Commit()

	return nil
}

func (p PNG) TransferTo(ctx context.Context, tx *gorm.DB, user model.User, sum ploutos.UserSum, currency string, gameCode string, gameVendorId int64, extra model.Extra) (balance int64, err error) {
	switch {
	case sum.Balance == 0:
		return 0, nil
	case sum.Balance < 0:
		return 0, errors.New("Evo::TransferTo not allowed to transfer negative sum")
	}
	png_service := game_service_png.PNG{}
	_, tx_id, err := png_service.TransferIn(os.Getenv("GAME_PNG_HOST"),"CreditAccount",user.ID, util.MoneyFloat(sum.Balance),fmt.Sprintf("PNG%d",time.Now().Unix()))
	if err !=nil{
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
		ExternalTransactionId: tx_id,
		GameVendorId:          gameVendorId,
	}
	db_tx:=tx.Begin()

	defer db_tx.Rollback()
	err = db_tx.Create(&transaction).Error
	if err != nil {
		db_tx.Rollback()
		return 0, err
	}
	err = db_tx.Model(ploutos.UserSum{}).Where(`user_id`, user.ID).Update(`balance`, 0).Error
	if err != nil {
		db_tx.Rollback()
		return 0, err
	}
	db_tx.Commit()

	return sum.Balance, nil
}
