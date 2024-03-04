package dollar_jackpot_api

import (
	"github.com/gin-gonic/gin"
	"web-api/api"
	"web-api/service/dollar_jackpot"
	"web-api/util"
)

func PlaceOrder(c *gin.Context) {
	var req dollar_jackpot.PlaceOrder
	if err := c.ShouldBind(&req); err == nil {
		res, e := dollar_jackpot.Place(c, req)
		c.JSON(200, res)
		if e != nil {
			util.Log().Error("dollar jackpot place order: ", e)
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}

func SettleOrder(c *gin.Context) {
	var req dollar_jackpot.SettleOrder
	if err := c.ShouldBind(&req); err == nil {
		res, e := dollar_jackpot.Settle(c, req)
		c.JSON(200, res)
		if e != nil {
			util.Log().Error("dollar jackpot settle order: ", e)
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}
