package api

import (
	"web-api/api"
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func ChopBet(c *gin.Context) {
	var service service.AnalystService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.ChopBet(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
