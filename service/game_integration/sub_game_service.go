package game_integration

import (
	"fmt"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"github.com/gin-gonic/gin"
)

type SubGameService struct {
	Platform int64 `form:"platform" json:"platform" binding:"required"`
}

func (service *SubGameService) List(c *gin.Context) (serializer.Response, error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	brandId := c.MustGet("_brand").(int)

	platform, ok := consts.PlatformIdToGameVendorColumn[service.Platform]
	if !ok {
		r := serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("invalid_platform"), nil)
		return r, nil
	}

	var subGames []ploutos.SubGameCGameVendorBrand
	tx := model.DB.Model(ploutos.SubGameCGameVendorBrand{}).Preload("GameVendorBrand").Joins(fmt.Sprintf(`LEFT JOIN game_vendor_brand gvb on gvb.game_vendor_id = %s.vendor_id`, ploutos.SubGameCGameVendorBrand{}.TableName())).Where("gvb.brand_id = ?", brandId).Where(fmt.Sprintf("gvb.%s = ?", platform), 1).Find(&subGames)
	if err := tx.Error; err != nil {
		return serializer.Response{
			Data: []serializer.SubGamesByGameType{},
		}, err
	}

	data, err := serializer.BuildSubGamesByGameType(subGames)
	if err != nil {
		return serializer.Response{
			Data: []serializer.SubGamesByGameType{},
		}, err
	}

	return serializer.Response{
		Data: data,
	}, nil
}
