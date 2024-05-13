package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.golang/paho"
	"github.com/gin-gonic/gin"
	validator "gopkg.in/go-playground/validator.v8"
	"time"
	"web-api/conf"
	"web-api/serializer"
	"web-api/util"
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

func FinpayRedirect(c *gin.Context) {
	data := map[string]any{}
	c.MultipartForm()
	for key, value := range c.Request.PostForm {
		data[key] = value[0]
	}
	for key, value := range c.Request.URL.Query() {
		data[key] = value[0]
	}
	j, _ := json.Marshal(data)
	if v, exists := data["user_id"]; exists {
		if vv, ok := v.(string); ok && vv != "" {
			pb := &paho.Publish{
				Topic:   fmt.Sprintf(`finpay_redirect/user/%s`, vv),
				QoS:     byte(1),
				Payload: j,
			}
			util.MQTTClient.Publish(context.Background(), pb)
		}
	}
	c.Status(200)
}

func Heartbeat(c *gin.Context) {
	c.JSON(200, serializer.Response{
		Code: 0,
	})
}

func ErrorResponse(c *gin.Context, service any, err error) serializer.Response {
	t, exists := c.Get("i18n")
	if !exists {
		return serializer.Err(c, service, serializer.CodeParamErr, "", err)
	}
	i18n := t.(i18n.I18n)
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

func ErrorResponseWithMsg(c *gin.Context, service any, err error, msg string) serializer.Response {
	res := ErrorResponse(c, service, err)
	res.Msg = msg
	return res
}
