package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	validator "gopkg.in/go-playground/validator.v8"
	"time"
	"web-api/conf"
	"web-api/model"
	"web-api/serializer"
)

func Ping(c *gin.Context) {
	country, _ := c.Get("_country")
	city, _ := c.Get("_city")
	c.JSON(200, serializer.Response{
		Code: 0,
		Msg:  "Pong",
		Data: map[string]interface{}{
			"country": country,
			"city":    city,
		},
	})
}

func Ts(c *gin.Context) {
	c.JSON(200, serializer.Response{
		Code: 0,
		Data: time.Now().Unix(),
	})
}

func Me(c *gin.Context) {
	u, _ := c.Get("user")
	user := u.(model.User)
	c.JSON(200, serializer.Response{
		Code: 0,
		Data: serializer.BuildUserInfo(c, user),
	})
}

func ErrorResponse(err error) serializer.Response {
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			field := conf.T(fmt.Sprintf("Field.%s", e.Field))
			tag := conf.T(fmt.Sprintf("Tag.Valid.%s", e.Tag))
			return serializer.ParamErr(
				fmt.Sprintf("%s%s", field, tag),
				err,
			)
		}
	}
	if _, ok := err.(*json.UnmarshalTypeError); ok {
		return serializer.ParamErr("JSON类型不匹配", err)
	}

	return serializer.ParamErr("参数错误", err)
}
