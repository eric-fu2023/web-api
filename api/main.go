package api

import (
	"encoding/json"
	"fmt"
	"time"
	"web-api/conf"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"

	models "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/eclipse/paho.golang/paho"
	"github.com/gin-gonic/gin"
	validator "gopkg.in/go-playground/validator.v8"
)

func Ping(c *gin.Context) {
	i18n := i18n.I18n{}
	if err := i18n.LoadLanguages("en"); err != nil {
		fmt.Println(err)
	}
	c.Set("i18n", i18n)
	common.SendTeamupGamePopupNotificationSocketMsg(3621, int64(188), int64(1727320740), int64(60000)/100, "IMOne Slot", "https://static.tayalive.com/batace-img/icon/IMOne-min.png") // DEBUG PURPOSE
	common.SendTeamupGamePopupNotificationSocketMsg(3671, int64(188), int64(1727320740), int64(60000)/100, "IMOne Slot", "https://static.tayalive.com/batace-img/icon/IMOne-min.png") // DEBUG PURPOSE

	common.SendUserSumSocketMsg(3671, models.UserSum{}, "ADAA", 5000)
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
				QoS:     byte(0),
				Payload: j,
			}
			go func(pb *paho.Publish) {
				if e := util.Publish(pb); e != nil {
					util.Log().Error(`finpay_redirect MQTT PUBLISH ERROR %v`, e)
				}
			}(pb)
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
