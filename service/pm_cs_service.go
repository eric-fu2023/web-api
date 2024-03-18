package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.golang/paho"
	"github.com/gin-gonic/gin"
	"os"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"
)

type CsSendService struct {
	Type    int64  `form:"type" json:"type" binding:"required"`
	Message string `form:"message" json:"message" binding:"required"`
}

func (service *CsSendService) Send(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var userRef string
	u, isUser := c.Get("user")
	if isUser {
		user := u.(model.User)
		userRef = fmt.Sprintf(`%d`, user.ID)
	} else {
		deviceInfo, e := util.GetDeviceInfo(c)
		if e != nil || deviceInfo.Uuid == "" {
			r = serializer.ParamErr(c, service, i18n.T("missing_device_uuid"), e)
			return
		}
		userRef = deviceInfo.Uuid
	}
	pm := ploutos.PrivateMessage{
		Message: service.Message,
		Type:    service.Type,
		UserRef: userRef,
	}
	err = model.DB.Create(&pm).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var data *serializer.PrivateMessage
	if pm.ID != 0 {
		t := serializer.BuildPrivateMessage(pm)
		data = &t
	}
	if data != nil {
		go sendMqtt(userRef, *data)
	}

	r = serializer.Response{
		Data: data,
	}
	return
}

type CsHistoryService struct {
	common.PageById
}

func (service *CsHistoryService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var userRef string
	u, isUser := c.Get("user")
	if isUser {
		user := u.(model.User)
		userRef = fmt.Sprintf(`%d`, user.ID)
	} else {
		deviceInfo, e := util.GetDeviceInfo(c)
		if e != nil || deviceInfo.Uuid == "" {
			r = serializer.ParamErr(c, service, i18n.T("missing_device_uuid"), e)
			return
		}
		userRef = deviceInfo.Uuid
	}
	var messages []ploutos.PrivateMessage
	q := model.DB.Model(ploutos.PrivateMessage{}).Where(`user_ref = ? OR user_ref = '0'`, userRef).
		Order(`created_at DESC, id DESC`).Limit(service.PageById.Limit)
	if service.PageById.IdFrom != 0 {
		q = q.Where(`id < ?`, service.PageById.IdFrom)
	}
	err = q.Find(&messages).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var data []serializer.PrivateMessage
	for _, message := range messages {
		data = append(data, serializer.BuildPrivateMessage(message))
	}

	r = serializer.Response{
		Data: data,
	}
	return
}

func sendMqtt(userRef string, message serializer.PrivateMessage) (err error) {
	j, err := json.Marshal(message)
	if err != nil {
		util.Log().Error("private message send mqtt error", err)
		return
	}
	pb := &paho.Publish{
		Topic:   fmt.Sprintf(`%s/cs/user/%s`, os.Getenv("MQTT_PREFIX"), userRef),
		QoS:     byte(1),
		Payload: j,
	}
	_, err = util.MQTTClient.Publish(context.Background(), pb)
	if err != nil {
		util.Log().Error("private message send mqtt error", err)
		return
	}
	return
}
