package dollar_jackpot_api

import (
	"errors"
	"web-api/api"
	"web-api/service/dollar_jackpot"

	"github.com/gin-gonic/gin"
)

func DollarJackpotGet(c *gin.Context) {
	var service dollar_jackpot.DollarJackpotGetService
	if err := c.Bind(&service); err == nil {
		if service.StreamerId < 1 {
			err := errors.New("Please using valid streamer id")
			c.JSON(400, api.ErrorResponse(c, service, err))
			return
		}
		res, e := service.Get(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func DollarJackpotWinners(c *gin.Context) {
	var service dollar_jackpot.DollarJackpotWinnersService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.List(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func DollarJackpotBetReport(c *gin.Context) {
	var service dollar_jackpot.DollarJackpotBetReportService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.List(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
