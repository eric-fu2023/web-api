package middleware

import (
	"strconv"
	"web-api/serializer"

	"github.com/gin-gonic/gin"
)

func BrandAgent() gin.HandlerFunc {
	return func(c *gin.Context) {
		//if c.GetHeader("Brand") == "" || c.GetHeader("Agent") == "" {
		if c.GetHeader("Brand") == "" {
			c.JSON(400, serializer.Response{
				Code:  serializer.CodeParamErr,
				Msg:   "Invalid headers",
				Error: `'Brand' header cannot be empty`,
			})
			c.Abort()
			return
		}
		brand, err := strconv.Atoi(c.GetHeader("Brand"))
		if err != nil {
			c.JSON(400, serializer.Response{
				Code:  serializer.CodeParamErr,
				Msg:   "Invalid headers",
				Error: err.Error(),
			})
			c.Abort()
			return
		}
		
		// should use agent id, but now we need to send data to pixel......
		agent, err := strconv.Atoi(c.GetHeader("Agent"))
		if err != nil {
			c.JSON(400, serializer.Response{
				Code:  serializer.CodeParamErr,
				Msg:   "Invalid headers",
				Error: err.Error(),
			})
			c.Abort()
			return
		}
		c.Set("_agent", agent)



		c.Set("_brand", brand)
		c.Next()
	}
}
