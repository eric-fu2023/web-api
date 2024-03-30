package service

import (
	"web-api/model"
	"web-api/serializer"
	"web-api/util"

	"github.com/gin-gonic/gin"
)

type VipQuery struct {
}

func (s VipQuery) Get(c *gin.Context) serializer.Response {
	user := c.MustGet("user").(model.User)

	vip, err := model.GetVip(c, user.ID)
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err)
	}
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err)
	}

	return serializer.Response{
		Data: serializer.BuildVip(vip),
	}
}

type VipLoad struct {
}

func (s VipLoad) Load(c *gin.Context) serializer.Response {

	list, err := model.LoadRule(c)
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err)
	}
	if err != nil {
		return serializer.Err(c, s, serializer.CodeGeneralError, "", err)
	}

	return serializer.Response{
		Data: util.MapSlice(list, serializer.BuildVipRule),
	}
}
