package game_integration

import (
	"fmt"
	"slices"

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

var gameTypeOrdering = map[string]int{
	"LIVE":   0,
	"CRASH":  1,
	"FLASH":  2,
	"SPRIBE": 3,
	"BOARD":  4,
	"SLOTS":  5,
	"TABLE":  6,
}

func (service *SubGameService) List(c *gin.Context) (serializer.Response, error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	brandId := c.MustGet("_brand").(int)

	platform, ok := consts.PlatformIdToGameVendorColumn[service.Platform]
	if !ok {
		r := serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("invalid_platform"), nil)
		return r, nil
	}

	var subGames []ploutos.SubGameBrand
	tx := model.DB.Model(ploutos.SubGameBrand{}).Preload("GameVendorBrand").Joins(fmt.Sprintf(`LEFT JOIN game_vendor_brand gvb on gvb.id = %s.vendor_brand_id`, ploutos.SubGameBrand{}.TableName())).Where(fmt.Sprintf("%s.brand_id = %d", ploutos.SubGameBrand{}.TableName(), brandId)).Where(fmt.Sprintf("%s.%s = ?", ploutos.SubGameBrand{}.TableName(), platform), 1).Find(&subGames)
	if err := tx.Error; err != nil {
		return serializer.Response{
			Data: []serializer.SubGamesBrandsByGameType{},
		}, err
	}

	slices.SortFunc(subGames, func(a, b ploutos.SubGameBrand) int {
		return int(a.SortRanking - b.SortRanking)
	})

	data, err := serializer.BuildSubGamesByGameType(subGames, gameTypeOrdering)
	if err != nil {
		return serializer.Response{
			Data: []serializer.SubGamesBrandsByGameType{},
		}, err
	}

	return serializer.Response{
		Data: data,
	}, nil
}
