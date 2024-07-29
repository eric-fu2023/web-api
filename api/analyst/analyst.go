package api

import (
	"web-api/api"
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func GetAnalystList(c *gin.Context) {
	var service service.AnalystService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.GetList(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
