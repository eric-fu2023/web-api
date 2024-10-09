package service

import (
	"fmt"
	"os"
	"strconv"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type ServiceWager struct {}
type CheckGameWalletStruct struct {
	GameBalance int64 `json:"game_balance"`
	UserId      int64 `json:"user_id"`
}
type MainWalletStruct struct {
	Balance int64 `json:"game_balance"`
	UserId  int64 `json:"user_id"`
}

func (service_wager *ServiceWager) WagerClear(c *gin.Context) (serializer.Response ){

	min_bet_string := os.Getenv("MIN_BET_VALUE")
	min_bet_int, err := strconv.ParseInt(min_bet_string, 10, 64)
	if err != nil {
		util.Log().Error("Err parse MIN_BET_VALUE env", err.Error())
	}
	// check user balance < env value
	u, _ := c.Get("user")
	user, _ := u.(model.User)

	// check if user total balance is below threshold
	balance_below_threshold := CheckWallet(user.ID, min_bet_int)

	if balance_below_threshold{
		// check if there are pending bet_report
		should_clear := CheckBetReport(user.ID)

		// for redis checking
		if should_clear {
			ResetWager(user.ID)
		}
	}
	return serializer.Response{
		Data: nil,
	}
}

func CheckWallet(user_id int64, min_bet_value int64) (should_clear_wager bool) {
	var main_wallet MainWalletStruct
	err := model.DB.Table("user_sums").Select("user_id, balance").
		Where("balance < ?", min_bet_value).
		Where("remaining_wager > 0").
		Where("user_id", user_id).
		Find(&main_wallet).Error
	if err != nil {
		util.Log().Error("Err checking users' main wallets", err.Error())
		return false
	}

	if main_wallet.UserId==0{
		return false
	}

	var check_game_wallet CheckGameWalletStruct
	err = model.DB.Table("users").Select("SUM(game_vendor_users.balance) as game_balance, users.id as user_id").
		Joins("LEFT JOIN game_vendor_users ON users.id = game_vendor_users.user_id").
		Where("users.id", user_id).
		Group("users.id").
		Find(&check_game_wallet).Error
	if err != nil {
		util.Log().Error("Err checking users' game wallets", err.Error())
		return false
	}

	if (check_game_wallet.GameBalance + main_wallet.Balance) < min_bet_value {
		return true
	}

	return false
}

func CheckBetReport(user_id int64) (should_clear bool) {
	var users_bet_reports []models.BetReport
	terminated_status := []int64{2, 3, 4, 5, 6}
	err := model.DB.Table("bet_report").Where("user_id", user_id).Where("status not in ?", terminated_status).Find(&users_bet_reports).Error
	if err != nil {
		util.Log().Error("Err checking users' bet report", err.Error())
		return false
	}

	// var OrderStatusMap = map[int64]string{
	// 	0: "Created",
	// 	1: "Confirming",
	// 	2: "Rejected",
	// 	3: "Canceled",
	// 	4: "Confirmed",
	// 	5: "Settled",
	// 	6: "EarlySettled",
	// }
	
	if len(users_bet_reports) != 0{
		return false
	}
	util.Log().Info("should clear wager for user :", user_id)
	return true
}

func ResetWager (user_id int64){
	//
	tx := model.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var usData models.UserSum
	// find the balance before/after and wager before/after from user_sums
	if err := tx.Model(&models.UserSum{}).Where("user_id = ?", user_id).First(&usData).Error; err != nil {
		util.Log().Error("Err find the balance before/after and wager before/after from user_sums", err.Error())
		tx.Rollback()
	}

	// log.Println("from update db spin settlement function:  ", usData)

	tsx := models.Transaction{
		Amount:               0,
		Wager:                -usData.RemainingWager,
		DepositWager:         -usData.DepositRemainingWager,
		TransactionType:      models.TransactionTypeClearWager,
		GameVendorId:         0,
		ForeignTransactionId: 0,
		IsAdjustment:         false,
		BalanceBefore:        usData.Balance,
		BalanceAfter:         usData.Balance,
		WagerBefore:          usData.RemainingWager,
		WagerAfter:           0,
		DepositWagerBefore:   usData.DepositRemainingWager,
		DepositWagerAfter:    0,
		UserId:               user_id,
	}

	// insert into transaction table.
	if err := tx.Create(&tsx).Error; err != nil {
		util.Log().Error("Err insert into transaction", err.Error())
		tx.Rollback()
	}
	fmt.Println("insert into transaction table for user ", user_id)

	// update user_sum table's balance and remaining_wager for that particular userId.
	if err := tx.Model(&models.UserSum{}).Where("user_id = ?", user_id).UpdateColumn("remaining_wager", 0).UpdateColumn("deposit_remaining_wager", 0).Error; err != nil {
		util.Log().Error("Err update remaining wager", err.Error())
		tx.Rollback()
	}
	fmt.Println("update user_sum table's balance and remaining_wager for that particular userId. ", user_id)

	risk_tag_id_string := os.Getenv("WAGER_AUTO_CLEAR_RISK_TAG_ID")
	risk_tag_id, read_env_err := strconv.ParseInt(risk_tag_id_string, 10, 64)
	if read_env_err != nil {
		util.Log().Error("Err parse risk tag id", read_env_err.Error())
	}
	// add user risk tag
	risk_tag := models.UserTagConn{
		UserId:    user_id,
		UserTagId: risk_tag_id,
	}
	if err := tx.Where("user_id = ? AND user_tag_id = ?", user_id, risk_tag_id).FirstOrCreate(&risk_tag).Error; err != nil {
		util.Log().Error("Err inserting or finding user risk tag", err.Error())
		tx.Rollback()
	}

	err := tx.Commit().Error
	if err != nil {
		util.Log().Error("Err tx commit error", err.Error())
		tx.Rollback()
	}
}
