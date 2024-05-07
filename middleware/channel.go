package middleware

import (
	"github.com/gin-gonic/gin"
)

func Channel() gin.HandlerFunc {
	return func(c *gin.Context) {
		channel := c.GetHeader("Channel")
		c.Set("_channel", channel)
		c.Next()
	}
}
