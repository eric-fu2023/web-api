package dc

import (
	"blgit.rfdev.tech/taya/game-service/dc/callback"
	"github.com/gin-gonic/gin"
	"web-api/conf/consts"
	"web-api/service/common"
)

func LoginCallback(c *gin.Context, req callback.LoginRequest) (res callback.BaseResponse, err error) {
	gpu, balance, _, _, err := common.GetUserAndSum(consts.GameVendor["dc"], req.BrandUid)
	if err != nil {
		return
	}
	if err != nil {
		return
	}
	res = callback.BaseResponse{
		Code: 1000,
		Data: callback.LoginResponse{
			BrandUid: gpu.ExternalUserId,
			Currency: gpu.ExternalCurrency,
			Balance:  float64(balance) / 100,
		},
	}
	return
}
