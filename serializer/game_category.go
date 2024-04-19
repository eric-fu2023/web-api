package serializer

import (
	"web-api/util/i18n"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
)

type GameCategory struct {
	Id     int64             `json:"category_id"`
	Vendor []GameVendorBrand `json:"vendor,omitempty"`
}

func BuildGameCategory(c *gin.Context, a ploutos.GameCategory, gameIds []int64) (b GameCategory) {
	b = GameCategory{
		Id: a.ID,
	}
	if len(a.GameVendorBrand) > 0 {
		for i, v := range a.GameVendorBrand {
			if len(gameIds) > i {
				t := BuildGameVendorBrandForCategory(v, gameIds[i])
				b.Vendor = append(b.Vendor, t)
			}
		}
	} else {
		return GameCategory{}
	}
	return
}

func FormatGameCategoryName(c *gin.Context, id int64) string {
	i18n := c.MustGet("i18n").(i18n.I18n)

	switch id {
	case ploutos.GameCategoryIdSports:
		return i18n.T("game_category_name_sports")
	case ploutos.GameCategoryIdLive:
		return i18n.T("game_category_name_live")
	case ploutos.GameCategoryIdElectronic:
		return i18n.T("game_category_name_electronic")
	case ploutos.GameCategoryIdCard:
		return i18n.T("game_category_name_card")
	case ploutos.GameCategoryIdEsports:
		return i18n.T("game_category_name_esports")
	case ploutos.GameCategoryIdLottery:
		return i18n.T("game_category_name_lottery")
	case ploutos.GameCategoryIdFishing:
		return i18n.T("game_category_name_fishing")
	default:
		return ""
	}
}
