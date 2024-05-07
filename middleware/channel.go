package middleware

import (
	"github.com/gin-gonic/gin"
	"web-api/util"
)

func Channel() gin.HandlerFunc {
	return func(c *gin.Context) {
		var channel string
		if deviceInfo, e := util.GetDeviceInfo(c); e == nil {
			channel = deviceInfo.Channel
		}
		c.Set("_channel", channel)
		c.Next()
	}
}
