package api

import (
	"github.com/gin-gonic/gin"
	"web-api/service"
)

func RtcToken(c *gin.Context) {
	var service service.RtcTokenService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.Get(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func RtcTokens(c *gin.Context) {
	var service service.RtcTokensService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.Get(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
