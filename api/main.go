package api

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
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
	var userSum ploutos.UserSum
	if e := model.DB.Where(`user_id`, user.ID).First(&userSum).Error; e == nil {
		user.UserSum = &userSum
	}
	c.JSON(200, serializer.Response{
		Code: 0,
		Data: serializer.BuildUserInfo(c, user),
	})
}

func Heartbeat(c *gin.Context) {
	c.JSON(200, serializer.Response{
		Code: 0,
	})
}

func ErrorResponse(c *gin.Context, service any, err error) serializer.Response {
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			field := conf.T(fmt.Sprintf("Field.%s", e.Field))
			tag := conf.T(fmt.Sprintf("Tag.Valid.%s", e.Tag))
			return serializer.ParamErr(
				c, service,
				fmt.Sprintf("%s%s", field, tag),
				err,
			)
		}
	}
	if _, ok := err.(*json.UnmarshalTypeError); ok {
		return serializer.ParamErr(c, service, "JSON类型不匹配", err)
	}

	return serializer.ParamErr(c, service, "参数错误", err)
}
