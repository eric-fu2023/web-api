package api

import (
	"web-api/serializer"

	"github.com/gin-gonic/gin"
)

type Doable interface {
	Do(c *gin.Context) (serializer.Response, error)
}

func Do[T Doable](c *gin.Context) {
	var service T
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Do(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(200, serializer.EnsureErr(c, err, res))
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
