package middleware

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

func AB() gin.HandlerFunc {
	return func(c *gin.Context) {
		isA, err := strconv.ParseBool(c.GetHeader("IsA"))
		if err != nil {
			isA = false
		}

		c.Set("_isA", isA)
		c.Next()
	}
}
