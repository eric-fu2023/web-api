package api

import (
	"web-api/api"
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func StartTeamUp(c *gin.Context) {
	var service service.GetTeamupService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.StartTeamUp(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func GetTeamUpItem(c *gin.Context) {
	var service service.GetTeamupService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Get(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func ListAllTeamUp(c *gin.Context) {
	var service service.TeamupService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.List(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func SlashBet(c *gin.Context) {
	var service service.TeamupService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.SlashBet(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}
