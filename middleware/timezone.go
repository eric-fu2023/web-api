package middleware

import (
	"github.com/gin-gonic/gin"
	"time"
)

func Timezone() gin.HandlerFunc {
	return func(c *gin.Context) {
		tz, _ := time.LoadLocation("UTC")
		if c.GetHeader("Timezone") != "" {
			if loc, e := time.LoadLocation(c.GetHeader("Timezone")); e == nil {
				tz = loc
			}
		}
		c.Set("_tz", tz)
		c.Next()
	}
}
