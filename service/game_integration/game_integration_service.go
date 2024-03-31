package game_integration

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"web-api/cache"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/ugs"
	"web-api/util"
	"web-api/util/i18n"
)

type GetUrlService struct {
	GameId   int64 `form:"game_id" json:"game_id" binding:"required"`
	Platform int64 `form:"platform" json:"platform" binding:"required"`
}

func (service *GetUrlService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	lang := c.MustGet("_language").(string)
	user := c.MustGet("user").(model.User)
	var isTestUser bool
	if user.Role == 2 {
		isTestUser = true
	}

	var subGame ploutos.SubGameC
	err = model.DB.Preload(`GameVendor`).Where(`id`, service.GameId).First(&subGame).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}

	var url string
	client := util.UgsFactory.NewClient(cache.RedisClient)
	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		var lastPlayed ploutos.GameVendorUser
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload(`GameVendor`).Where(`user_id`, user.ID).Where(`is_last_played`, true).
			Order(`updated_at DESC`).Limit(1).Find(&lastPlayed).Error
		if err != nil {
			return
		}
		var sum ploutos.UserSum
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(`user_id`, user.ID).First(&sum).Error
		if err != nil {
			return
		}
		if lastPlayed.ID != 0 && lastPlayed.GameVendorId != int64(subGame.VendorId) {
			if lastPlayed.GameVendor.GameIntegrationId == ugs.IntegrationIdUGS {
				balance, status, ptxid, e := client.TransferOut(user.ID, user.Username, lastPlayed.ExternalCurrency, lang, lastPlayed.GameVendor.GameCode, c.ClientIP(), isTestUser)
				if e != nil {
					err = e
					return
				}
				util.Log().Info("GAME INTEGRATION TRANSFER OUT game_integration_id: %d, user_id: %d, balance: %.4f, status: %d, tx_id: %s", ugs.IntegrationIdUGS, user.ID, balance, status, ptxid)
				if status == ugs.TransferStatusSuccess && balance > 0 && ptxid != "" {
					amount := util.MoneyInt(balance)
					transaction := ploutos.Transaction{
						UserId:                user.ID,
						Amount:                amount,
						BalanceBefore:         sum.Balance,
						BalanceAfter:          sum.Balance + amount,
						TransactionType:       ploutos.TransactionTypeFromUGS,
						Wager:                 0,
						WagerBefore:           sum.RemainingWager,
						WagerAfter:            sum.RemainingWager,
						ExternalTransactionId: ptxid,
					}
					err = tx.Create(&transaction).Error
					if err != nil {
						return
					}
					err = tx.Model(ploutos.UserSum{}).Where(`user_id`, user.ID).Update(`balance`, gorm.Expr(`balance + ?`, amount)).Error
					if err != nil {
						return
					}
					sum.Balance = sum.Balance + amount
				}
			}
		}
		if sum.Balance > 0 {
			status, ptxid, e := client.TransferIn(user.ID, user.Username, lastPlayed.ExternalCurrency, lang, subGame.GameVendor.GameCode, c.ClientIP(), isTestUser, util.MoneyFloat(sum.Balance))
			if e != nil {
				err = e
				return
			}
			util.Log().Info("GAME INTEGRATION TRANSFER IN game_integration_id: %d, user_id: %d, balance: %.4f, status: %d, tx_id: %s", ugs.IntegrationIdUGS, user.ID, util.MoneyFloat(sum.Balance), status, ptxid)
			if status == ugs.TransferStatusSuccess && ptxid != "" {
				transaction := ploutos.Transaction{
					UserId:                user.ID,
					Amount:                -1 * sum.Balance,
					BalanceBefore:         sum.Balance,
					BalanceAfter:          0,
					TransactionType:       ploutos.TransactionTypeToUGS,
					Wager:                 0,
					WagerBefore:           sum.RemainingWager,
					WagerAfter:            sum.RemainingWager,
					ExternalTransactionId: ptxid,
				}
				err = tx.Create(&transaction).Error
				if err != nil {
					return
				}
				err = tx.Model(ploutos.UserSum{}).Where(`user_id`, user.ID).Update(`balance`, 0).Error
				if err != nil {
					return
				}
			}
		}
		url, err = client.GetGameUrl(user.ID, user.Username, lastPlayed.ExternalCurrency, lang, subGame.GameVendor.GameCode, subGame.GameCode, c.ClientIP(), ugs.PlatformMapping[service.Platform], isTestUser)
		if err != nil {
			return
		}
		err = tx.Model(ploutos.GameVendorUser{}).Where(`id`, lastPlayed.ID).Update(`is_last_played`, false).Error
		if err != nil {
			return
		}
		err = tx.Model(ploutos.GameVendorUser{}).Where(`game_vendor_id`, subGame.GameVendor.ID).Where(`user_id`, user.ID).Update(`is_last_played`, true).Error
		if err != nil {
			return
		}
		return
	})

	r = serializer.Response{
		Data: url,
	}
	return
}
