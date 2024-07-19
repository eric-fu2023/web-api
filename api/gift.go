package api

import (
	"web-api/serializer"
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func GiftList(c *gin.Context) {
	var service service.GiftListService
	if err := c.ShouldBind(&service); err == nil {
		res, err := service.List(c)
		if err != nil {
			c.JSON(500, serializer.Err(c, "", 50000, err.Error(), err))
		}
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
