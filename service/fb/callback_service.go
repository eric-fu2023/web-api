package fb

import (
	"blgit.rfdev.tech/taya/game-service/fb/callback"
	"fmt"
	"github.com/gin-gonic/gin"
	"web-api/conf/consts"
	"web-api/model"
)

func BalanceCallback(c *gin.Context, req callback.BalanceRequest) (res callback.BaseResponse, err error) {
	var gpu model.GameProviderUser
	err = model.DB.Where(`game_provider_id`, consts.GameProvider["fb"]).Where(`external_user_id`, req.MerchantUserId).First(&gpu).Error
	if err != nil {
		res = callback.BaseResponse{
			Code: 1,
			Message: err.Error(),
		}
		return
	}

	var userSum model.UserSum
	err = model.DB.Where(`user_id`, gpu.UserId).First(&userSum).Error
	if err != nil {
		res = callback.BaseResponse{
			Code: 1,
			Message: err.Error(),
		}
		return
	}

	data := callback.BalanceResponse{
		Balance: fmt.Sprintf("%.2f", float64(userSum.Balance) / 100),
		CurrencyId: gpu.ExternalCurrencyId,
	}
	res = callback.BaseResponse{
		Code: 0,
		Data: data,
	}
	return
}

func OrderPayCallback(c *gin.Context, req callback.OrderPayRequest) (res callback.BaseResponse, err error) {
	res = callback.BaseResponse{
		Code: 0,
	}
	return
}