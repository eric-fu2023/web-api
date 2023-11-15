package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"web-api/service"
)

func UserFavouriteList(c *gin.Context) {
	var service service.UserFavouriteListService
	if err := c.ShouldBind(&service); err == nil {
		res, _ := service.List(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserFavouriteAdd(c *gin.Context) {
	var service service.UserFavouriteService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res, _ := service.Add(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func UserFavouriteRemove(c *gin.Context) {
	var service service.UserFavouriteService
	if err := c.ShouldBindWith(&service, binding.FormMultipart); err == nil {
		res, _ := service.Remove(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
