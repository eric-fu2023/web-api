package api

import (
	"github.com/gin-gonic/gin"
	"web-api/service"
)

func GameList(c *gin.Context) {
	var service service.GameListService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.List(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
