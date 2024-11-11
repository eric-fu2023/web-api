package middleware

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"web-api/serializer"
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
		//agent, err := strconv.Atoi(c.GetHeader("Agent"))
		//if err != nil {
		//	c.JSON(400, serializer.Response{
		//		Code:  serializer.CodeParamErr,
		//		Msg:   "Invalid headers",
		//		Error: err.Error(),
		//	})
		//	c.Abort()
		//	return
		//}
		c.Set("_brand", brand)
		//c.Set("_agent", agent)
		c.Next()
	}
}
