package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type AvatarListService struct {
	Ids string `form:"ids" json:"ids"`
}

func (service *AvatarListService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)

	var ids []int64
	for _, s := range strings.Split(service.Ids, ",") {
		if i, e := strconv.Atoi(strings.TrimSpace(s)); e == nil {
			ids = append(ids, int64(i))
		}
	}

	var users []ploutos.User
	err = model.DB.Model(ploutos.User{}).Scopes(model.ByIds(ids), model.SortById).Find(&users).Error
	if err != nil {
		r = serializer.DBErr(c, service, i18n.T("general_error"), err)
		return
	}

	var list []serializer.UserAvatar
	for _, u := range users {
		list = append(list, serializer.BuildUserAvatar(c, u))
	}

	r = serializer.Response{
		Data: list,
	}
	return
}
