package fb_api

import (
	"github.com/gin-gonic/gin"
	"web-api/serializer"
)

func CallbackHealth(c *gin.Context) {
	c.JSON(200, serializer.Response{
		Code: 0,
	})
}
