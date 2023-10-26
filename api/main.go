package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	validator "gopkg.in/go-playground/validator.v8"
	"time"
	"web-api/conf"
	"web-api/serializer"
	"web-api/util/i18n"
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

func Heartbeat(c *gin.Context) {
	c.JSON(200, serializer.Response{
		Code: 0,
	})
}

func ErrorResponse(c *gin.Context, service any, err error) serializer.Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
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
		return serializer.ParamErr(c, service, i18n.T("json_type_mismatch"), err)
	}

	return serializer.ParamErr(c, service, i18n.T("parameter_error"), err)
}
