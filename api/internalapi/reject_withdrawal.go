package internal_api

import (
	"web-api/api"
	"web-api/service/cashout"

	"github.com/gin-gonic/gin"
)

func RejectWithdrawal(c *gin.Context) {
	var service cashout.CancelCashOutOrderService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Reject(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, res)
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
