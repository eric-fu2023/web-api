package api

import (
	"github.com/gin-gonic/gin"
	"web-api/service"
)

func CategoryList(c *gin.Context) {
	var service service.CategoryListService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.List(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
