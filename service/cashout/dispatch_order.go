package cashout

import (
	"errors"
	"web-api/model"
	"web-api/service/common"
	"web-api/service/exchange"
	"web-api/util"

	"blgit.rfdev.tech/taya/payment-service/finpay"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

func DispatchOrder(c *gin.Context, cashOrder model.CashOrder, user model.User, accountBinding ploutos.UserAccountBinding, retryable bool) (updatedCashOrder model.CashOrder, err error) {
	updatedCashOrder = cashOrder
	method, err := model.CashMethod{}.GetByIDWithChannel(c, cashOrder.CashMethodId)
	if err != nil {
		return
	}
	channel, nErr := model.GetNextCashMethodChannel(model.FilterCashMethodChannelsByAmount(c, updatedCashOrder.AppliedCashOutAmount, model.FilterCashMethodChannelsByVip(c, user, method.CashMethodChannels)))
	if nErr != nil {
		err = nErr
		return
	}
	stats := channel.Stats
	err = processCashOutMethod(method)
	if err != nil {
		return
	}
	var exchangeClient exchange.ExchangeClient
	er, err := exchangeClient.GetExchangeRate(c, method.Currency, false)
	if err != nil {
		return
	}

	// if currency is not USDT, round down cashoutAmount to remove decimal
	// because most payment channels don't accept decimal places, except USDT wallet
	cashoutAmount := int64(float64(updatedCashOrder.AppliedCashOutAmount) * er.AdjustedExchangeRate)
	if method.Currency != exchange.USDT && er.ExchangeRate != 1 && er.AdjustedExchangeRate != 1 {
		cashoutAmount = (cashoutAmount / 100) * 100
	}

	switch channel.Gateway {
	case "finpay":
		config := channel.GetGatewayConfig()
		var data finpay.TransferOrderResponse
		defer func() {
			result := "success"
			if errors.Is(err, finpay.ErrorGateway) {
				result = "gateway_failed"
			}
			if data.IsFailed() {
				result = "failed"
			}
			_ = model.IncrementCashMethodStats(stats, result)

		}()

		switch config.Type {
		case "BANK_CARD":
			bankInfo := accountBinding.GetBankInfo()
			data, err = finpay.FinpayClient{}.DefaultBankCardCashOutV1(c, cashoutAmount, updatedCashOrder.ID, method.Currency, string(accountBinding.AccountNumber), string(accountBinding.AccountName), bankInfo.BankBranchName, bankInfo.BankCode, bankInfo.BankName, user.Username)
		case "TRC20":
			data, err = finpay.FinpayClient{}.DefaultTRC20CashOutV1(c, cashoutAmount, updatedCashOrder.ID, string(accountBinding.AccountNumber), string(accountBinding.AccountName), user.Username)
		default:
			bankInfo := accountBinding.GetBankInfo()
			accountName := string(accountBinding.AccountName)
			if len(accountName) == 0 {
				accountName = bankInfo.FirstName + " " + bankInfo.LastName
			}
			data, err = finpay.FinpayClient{}.DefaulGenericCashOutV1(c, cashoutAmount, updatedCashOrder.ID, method.Currency, string(accountBinding.AccountNumber), accountName, bankInfo.BankBranchName, bankInfo.BankCode, bankInfo.BankName, user.Username, config.ChannelCode, config.Type, config.TypeExtra)
		}
		if data.IsSuccess() {
			updatedCashOrder.Status = ploutos.CashOrderStatusTransferring
			updatedCashOrder.TransactionId = &data.TransferOrderNo
			updatedCashOrder.Notes = ploutos.EncryptedStr(util.JSON(data))
			updatedCashOrder.CurrencyCode = method.Currency
			updatedCashOrder.ExchangeRate = er.ExchangeRate
			updatedCashOrder.ExchangeRateAdjusted = er.AdjustedExchangeRate
			err = model.DB.Debug().WithContext(c).Omit(clause.Associations).Updates(&updatedCashOrder).Error
			go func() {
				userSum, _ := model.GetByUserIDWithLockWithDB(cashOrder.UserId, model.DB)
				common.SendUserSumSocketMsg(cashOrder.UserId, userSum.UserSum, "withdraw_pending", float64(updatedCashOrder.AppliedCashOutAmount)/100)
			}()

			if err != nil {
				return
			}
		} else if retryable {
			return
		} else if data.IsFailed() || errors.Is(err, finpay.ErrorGateway) {
			updatedCashOrder, err = RevertCashOutOrder(c, updatedCashOrder.ID, util.JSON(data), "Request Failed", ploutos.CashOrderStatusFailed, model.DB)
			if err != nil {
				return
			}
		} else if data.TransferOrderNo != "" {
			updatedCashOrder.TransactionId = &data.TransferOrderNo
			err = model.DB.Debug().WithContext(c).Omit(clause.Associations).Updates(&updatedCashOrder).Error
			if err != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}
	return
}

func processCashOutMethod(m model.CashMethod) (err error) {
	if m.MethodType > 0 {
		return errors.New("cash method not permitted")
	}
	return nil
}
