package api

import (
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func CashMethodList(c *gin.Context) {
	var service service.CasheMethodListService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.List(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}