package service

import (
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type SpinService struct {
}

func (service *SpinService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var spinItems []ploutos.SpinItem
	q := model.DB.Model(ploutos.SpinItem{}).Order(`id DESC`)
	err = q.Find(&spinItems).Error
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	data := make([]serializer.SpinItem, 0)
	for _, spinItem := range spinItems {
		data = append(data, serializer.BuildSpinItem(spinItem))
	}
	r = serializer.Response{
		Data: data,
	}
	return
}
