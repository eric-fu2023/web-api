package api

import (
	"web-api/api"
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func StartGameTeamUp(c *gin.Context) {
	var service service.GetTeamupService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.StartTeamUp(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func ListAllGameTeamUp(c *gin.Context) {
	var service service.TeamupService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.List(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
