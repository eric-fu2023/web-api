package serializer

import (
	"time"

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
	case ploutos.GameCategoryIdEGames:
		return i18n.T("game_category_name_egames")
	case ploutos.GameCategoryIdCard:
		return i18n.T("game_category_name_card")
	case ploutos.GameCategoryIdESports:
		return i18n.T("game_category_name_esports")
	case ploutos.GameCategoryIdLottery:
		return i18n.T("game_category_name_lottery")
	case ploutos.GameCategoryIdFishing:
		return i18n.T("game_category_name_fishing")
	default:
		return ""
	}
}

type SubGame struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	GameId           int64  `json:"game_id,omitempty"`
	Type             int64  `json:"type"`
	WebIcon          string `json:"web_icon,omitempty"`
	AppIcon          string `json:"app_icon,omitempty"`
	IsMaintenance    bool   `json:"is_maintenance,omitempty"`
	MaintenanceStart int64  `json:"maintenance_start,omitempty"`
	MaintenanceEnd   int64  `json:"maintenance_end,omitempty"`
}

type SubGamesByCategory struct {
	Id       int64     `json:"category_id"`
	SubGames []SubGame `json:"vendor,omitempty"`
}

type SubGamesMap map[ /*categoryId*/ int64][]SubGame

func (m SubGamesMap) AsSlice() (list []SubGamesByCategory) {
	for categoryId, subGames := range m {
		list = append(list, SubGamesByCategory{
			Id:       categoryId,
			SubGames: subGames,
		})
	}
	return list
}

func BuildSubGamesByCategory(_ *gin.Context, subGamesModel []ploutos.SubGameCGameVendorBrand) []SubGamesByCategory {
	subGamesMap := make(SubGamesMap)
	for _, sg := range subGamesModel {
		cId := sg.GameVendorBrand.CategoryId
		startTime, endTime := sg.GameVendorBrand.StartTime, sg.GameVendorBrand.EndTime
		now := time.Now()

		var isMaintenance bool
		var maintenanceStart, maintenanceEnd int64

		if !sg.GameVendorBrand.StartTime.IsZero() {
			if now.After(startTime) && (now.Before(endTime) || endTime.IsZero()) {
				isMaintenance = true
				maintenanceStart = startTime.Unix()
				if !endTime.IsZero() {
					maintenanceEnd = endTime.Unix()
				}
			}
		}

		subGamesMap[cId] = append(subGamesMap[cId], SubGame{
			ID:               sg.GameVendorBrand.GameVendorId,
			Name:             sg.Name,
			GameId:           sg.ID,
			Type:             cId,
			WebIcon:          Url(sg.WebIcon),
			AppIcon:          Url(sg.AppIcon),
			IsMaintenance:    isMaintenance,
			MaintenanceStart: maintenanceStart,
			MaintenanceEnd:   maintenanceEnd,
		})
	}

	return subGamesMap.AsSlice()
}
