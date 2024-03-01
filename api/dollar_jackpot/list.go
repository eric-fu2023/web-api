package dollar_jackpot

import (
	"github.com/gin-gonic/gin"
	"web-api/api"
	"web-api/service/dollar_jackpot"
)

func DollarJackpotGet(c *gin.Context) {
	var service dollar_jackpot.DollarJackpotGetService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.Get(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
