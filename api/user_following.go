package api

import (
	"github.com/gin-gonic/gin"
	"web-api/service"
)

func UserFollowingIdList(c *gin.Context) {
	var service service.UserFollowingListService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Ids(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserFollowingList(c *gin.Context) {
	var service service.UserFollowingListService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.List(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserFollowingAdd(c *gin.Context) {
	var service service.UserFollowingService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Add(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserFollowingRemove(c *gin.Context) {
	var service service.UserFollowingService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.Remove(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
