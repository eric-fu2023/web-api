package internal_api

import (
	"web-api/serializer"
	cashin "web-api/service/cashin"

	"github.com/gin-gonic/gin"
)

func ManualCloseCashInOrder(c *gin.Context) {
	var service cashin.ManualCloseService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.Do(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(500, serializer.EnsureErr(c, err, res))
		}
	} else {
		c.JSON(400, serializer.ParamErr(c, service, "", err))
	}
}
