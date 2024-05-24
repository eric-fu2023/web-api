package serializer

import (
	"errors"
	"slices"
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
	GameVendorId         int64 `json:"id"`
	GameVendorCategoryId int64 `json:"game_vendor_category_id"`

	Name                 string `json:"name"`
	Id                   int64  `json:"game_id,omitempty"`
	Type                 string `json:"type"`
	WebIcon              string `json:"web_icon,omitempty"`
	AppIcon              string `json:"app_icon,omitempty"`
	IsMaintenance        bool   `json:"is_maintenance,omitempty"`
	MaintenanceStartUnix int64  `json:"maintenance_start,omitempty"`
	MaintenanceEndUnix   int64  `json:"maintenance_end,omitempty"`
	SortRanking          int64  `json:"sort,omitempty"`
}

// SubGamesByGameType
// game_type M to category_id 1
type SubGamesByGameType struct {
	Type string `json:"game_type"`

	GameVendorCategoryId int64     `json:"category_id"`
	SubGames             []SubGame `json:"vendor,omitempty"`
}

type SubGamesMap map[string][]SubGame

func (m SubGamesMap) AsSlice(gameTypeOrder map[string]int) (list []SubGamesByGameType, err error) {

	if gameTypeOrder == nil {
		return
	}
	for subGameType, subGames := range m {
		if len(subGames) == 0 {
			return list, errors.New("build SubGamesByGameType aborted. cannot accept empty subGames")
		}
		list = append(list, SubGamesByGameType{
			Type:                 subGameType,
			GameVendorCategoryId: subGames[0].GameVendorCategoryId,
			SubGames:             subGames,
		})
	}

	slices.SortFunc(list, func(a, b SubGamesByGameType) int {
		return gameTypeOrder[a.Type] - gameTypeOrder[b.Type]
	})
	return list, nil
}

func BuildSubGamesByGameType(subGamesModel []ploutos.SubGameBrand, gameTypeOrder map[string]int) ([]SubGamesByGameType, error) {
	subGamesMap := make(SubGamesMap)
	for _, sg := range subGamesModel {
		gameType := sg.GameType
		gvCategoryId := sg.GameVendorBrand.CategoryId
		gvbMaintenanceStartTime, gvbMaintenanceEndTime := sg.GameVendorBrand.StartTime, sg.GameVendorBrand.EndTime

		now := time.Now()
		var isMaintenance bool
		if !sg.GameVendorBrand.StartTime.IsZero() {
			if now.After(gvbMaintenanceStartTime) && (now.Before(gvbMaintenanceEndTime) || gvbMaintenanceEndTime.IsZero()) {
				isMaintenance = true
			}
		}

		subGamesMap[gameType] = append(subGamesMap[gameType], SubGame{
			GameVendorId:         sg.GameVendorBrand.GameVendorId,
			GameVendorCategoryId: gvCategoryId,

			Name:                 sg.Name,
			Id:                   sg.SubGameId,
			Type:                 sg.GameType,
			WebIcon:              Url(sg.WebIcon),
			AppIcon:              Url(sg.AppIcon),
			IsMaintenance:        isMaintenance,
			MaintenanceStartUnix: gvbMaintenanceStartTime.Unix(),
			MaintenanceEndUnix:   gvbMaintenanceEndTime.Unix(),
			SortRanking:          sg.SortRanking,
		})
	}

	return subGamesMap.AsSlice(gameTypeOrder)
}
