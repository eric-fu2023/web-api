package service

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type GameListService struct {
	Platform int64 `form:"platform" json:"platform" binding:"required"`
	Type     int64 `form:"type" json:"type"`
}

func (service *GameListService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	brand := c.MustGet(`_brand`).(int)
	var gvb []ploutos.GameVendorBrand
	if err = model.DB.Model(ploutos.GameVendorBrand{}).Scopes(model.ByGameTypeAndBrand(service.Type, int64(brand)), model.ByPlatformExpended(service.Platform), model.ByStatus).Find(&gvb).Error; err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	var list []serializer.GameVendorBrand
	for _, g := range gvb {
		list = append(list, serializer.BuildGameVendorBrand(g))
	}

	r = serializer.Response{
		Data: list,
	}
	return
}
