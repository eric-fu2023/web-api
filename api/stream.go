package api

import (
	"github.com/gin-gonic/gin"
	"web-api/service"
)

func StreamList(c *gin.Context) {
	var service service.StreamService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.List(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, res)
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

//func FollowingStreams(c *gin.Context) {
//	var service v2.FollowingStreamService
//	if err := c.ShouldBind(&service); err == nil {
//		if res, err := service.List(c); err == nil {
//			c.JSON(200, res)
//		} else {
//			c.JSON(500, res)
//		}
//	} else {
//		c.JSON(400, api.ErrorResponse(err))
//	}
//}
//
//func Silenced(c *gin.Context) {
//	var service v2.StreamStatusService
//	if err := c.ShouldBind(&service); err == nil {
//		if res, err := service.Get(c); err == nil {
//			c.JSON(200, res)
//		} else {
//			c.JSON(500, res)
//		}
//	} else {
//		c.JSON(400, api.ErrorResponse(err))
//	}
//}
