package imsb_api

import (
	"blgit.rfdev.tech/taya/game-service/imsb/callback"
	"errors"
	"github.com/gin-gonic/gin"
	"web-api/api"
	"web-api/service/imsb"
)

func GetToken(c *gin.Context) {
	var service imsb.TokenService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func ValidateToken(c *gin.Context) {
	var service imsb.ValidateTokenService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.Validate(c)
		if e == nil {
			c.JSON(200, res)
		} else {
			var statusCode int64 = 400
			if errors.Is(e, imsb.ImsbErrInvalidMemberCode) {
				statusCode = 202
			}
			c.JSON(200, callback.BaseResponse{
				StatusCode: statusCode,
				StatusDesc: e.Error(),
			})
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
