package dollar_jackpot_api

import (
	"github.com/gin-gonic/gin"
	"web-api/api"
	"web-api/service/dollar_jackpot"
)

func PlaceOrder(c *gin.Context) {
	var req dollar_jackpot.PlaceOrder
	if err := c.ShouldBind(&req); err == nil {
		if res, err := dollar_jackpot.Place(c, req); err != nil {
			c.JSON(200, api.ErrorResponse(c, req, err))
		} else {
			c.JSON(200, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, req, err))
	}
}
