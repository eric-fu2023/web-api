package api

import (
	"web-api/api"
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func ListAnalysts(c *gin.Context) {
	var service service.AnalystService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.GetAnalystList(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func ListFollowingAnalysts(c *gin.Context) {
	var service service.AnalystService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.GetFollowingAnalystList(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func GetAnalystDetail(c *gin.Context) {
	var service service.AnalystDetailService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.GetAnalyst(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func ToggleFollowAnalyst(c *gin.Context) {
	var service service.FollowToggle
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.FollowAnalystToggle(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func GetAnalystAchievement(c *gin.Context) {
	var service service.AnalystAchievementService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.GetRecord(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}