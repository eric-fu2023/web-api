package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"strconv"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"
)

var ErrInsufficientBalance = errors.New("insufficient balance")
var ErrUpdateFailed = errors.New("update failed")

type TransferToService struct {
	VendorId int64 `form:"vendor_id" json:"vendor_id" binding:"required"`
}

func (service *TransferToService) Do(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	err = transferTo(user, service.VendorId)
	if err != nil {
		if errors.Is(err, ErrInsufficientBalance) {
			r = serializer.Err(c, "", serializer.CodeGeneralError, i18n.T("insufficient_balance"), err)
			return
		}
		r = serializer.Err(c, "", serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	r = serializer.Response{
		Msg: i18n.T("success"),
	}
	return
}

type TransferFromService struct {
	VendorId int64 `form:"vendor_id" json:"vendor_id" binding:"required"`
}

func (service *TransferFromService) Do(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	err = transferFrom(user, service.VendorId)
	if err != nil {
		r = serializer.Err(c, "", serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	r = serializer.Response{
		Msg: i18n.T("success"),
	}
	return
}

type TransferBackService struct {
}

func (service *TransferBackService) Do(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	err = transferBack(user)
	if err != nil {
		r = serializer.Err(c, "", serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	r = serializer.Response{
		Msg: i18n.T("success"),
	}
	return
}

func transferTo(user model.User, vendorId int64) (err error) {
	err = model.DB.Clauses(dbresolver.Use("txConn")).Transaction(func(tx *gorm.DB) (err error) {
		gpu, balance, _, _, err := common.GetUserAndSum(tx, vendorId, user.Username)
		if err != nil {
			return
		}
		if balance <= 0 {
			err = ErrInsufficientBalance
			return
		}

		txId := fmt.Sprintf(`%s-%d`, user.Username, time.Now().Unix())
		currency, err := strconv.Atoi(gpu.ExternalCurrency)
		if err != nil {
			return
		}
		rows := tx.Model(ploutos.UserSum{}).Where(`user_id`, user.ID).Update(`balance`, gorm.Expr(`balance - ?`, balance)).RowsAffected
		if rows == 0 {
			return
		}

		transfer := ploutos.Transfer{
			UserId:         user.ID,
			Amount:         balance * -1,
			BalanceBefore:  balance,
			BalanceAfter:   0,
			TxId:           txId,
			GameVendorId:   vendorId,
			ExternalUserId: user.Username,
		}
		err = tx.Save(&transfer).Error
		if err != nil {
			return
		}

		_, err = util.VendorIdToGameClient[vendorId].TransferTo(user.Username, txId, util.MoneyFloat(balance), int64(currency))
		if err != nil {
			return
		}
		return
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	return
}

func transferFrom(user model.User, vendorId int64) (err error) {
	balances, err := util.VendorIdToGameClient[vendorId].GetWalletBalance(user.Username)
	if err != nil {
		return
	}

	err = model.DB.Clauses(dbresolver.Use("txConn")).Transaction(func(tx *gorm.DB) (err error) {
		gpu, origBalance, _, _, err := common.GetUserAndSum(tx, vendorId, user.Username)
		if err != nil {
			return
		}

		balance := balances[gpu.ExternalCurrency]
		if balance == 0 { // if balance is zero, don't return error and return
			return
		}
		txId := fmt.Sprintf(`%s-%d`, user.Username, time.Now().Unix())
		currency, err := strconv.Atoi(gpu.ExternalCurrency)
		if err != nil {
			return
		}

		// TODO: update wager
		rows := tx.Model(ploutos.UserSum{}).Where(`user_id`, user.ID).Update(`balance`, gorm.Expr(`balance + ?`, util.MoneyInt(balance))).RowsAffected
		if rows == 0 {
			err = ErrUpdateFailed
			return
		}

		transfer := ploutos.Transfer{
			UserId:         user.ID,
			Amount:         util.MoneyInt(balance),
			BalanceBefore:  origBalance,
			BalanceAfter:   origBalance + util.MoneyInt(balance),
			TxId:           txId,
			GameVendorId:   vendorId,
			ExternalUserId: user.Username,
		}
		err = tx.Save(&transfer).Error
		if err != nil {
			return
		}

		_, err = util.VendorIdToGameClient[vendorId].TransferFrom(user.Username, txId, balance, int64(currency))
		if err != nil {
			return
		}
		return
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	return
}

func transferBack(user model.User) (err error) {
	var transfer ploutos.Transfer
	err = model.DB.Clauses(dbresolver.Use("txConn")).Where(`user_id`, user.ID).Order(`created_at DESC`).Limit(1).Find(&transfer).Error
	if err != nil {
		return
	}
	if transfer.ID == 0 { // no transfer record
		return
	}
	if transfer.Amount >= 0 { // if transfer is to game provider, do nothing and return
		return
	}
	err = transferFrom(user, transfer.GameVendorId)
	return
}
