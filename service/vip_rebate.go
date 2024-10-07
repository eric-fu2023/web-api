package service

import (
	"web-api/model"
	"web-api/serializer"

	"github.com/gin-gonic/gin"
)

type VipRebateQuery struct{}

func (v VipRebateQuery) Load(c *gin.Context) (r serializer.Response, err error) {
	list, err := model.LoadVipRebateRules(c)
	u, _ := c.Get("user")
	user := u.(model.User)
	if err != nil {
		r = serializer.Err(c, "", serializer.CodeGeneralError, "", err)
		return
	}
	vips, err := model.LoadVipRule(c)
	if err != nil {
		r = serializer.Err(c, "", serializer.CodeGeneralError, "", err)
		return
	}
	desc, err := GetCachedConfig(c, "rebate_description")
	if err != nil {
		r = serializer.Err(c, "", serializer.CodeGeneralError, "", err)
		return
	}
	r.Data = serializer.BuildVipRebateDetails(list, desc, vips, model.GetUserLang(user.ID))
	return
}
