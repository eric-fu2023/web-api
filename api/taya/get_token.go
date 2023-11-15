package taya_api

import (
	"github.com/gin-gonic/gin"
	"web-api/api"
	"web-api/service/taya"
)

func GetToken(c *gin.Context) {
	var service taya.TokenService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
