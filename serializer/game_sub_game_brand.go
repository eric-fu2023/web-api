// package serializer
// game_sub_game_brand mainly for url path /sub_games
package serializer

import (
	"errors"
	"slices"
	"time"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type BrandSubGame struct {
	GameVendorId         int64 `json:"id"`
	GameVendorCategoryId int64 `json:"game_vendor_category_id"`

	SubGameId int64 `json:"game_id,omitempty"`

	Name                 string `json:"name"`
	Type                 string `json:"type"`
	WebIcon              string `json:"web_icon,omitempty"`
	AppIcon              string `json:"app_icon,omitempty"`
	IsMaintenance        bool   `json:"is_maintenance,omitempty"`
	MaintenanceStartUnix int64  `json:"maintenance_start,omitempty"`
	MaintenanceEndUnix   int64  `json:"maintenance_end,omitempty"`
	SortRanking          int64  `json:"sort,omitempty"`

	SupportIFrame bool `json:"support_iframe"` // from `sub_game`
}

// SubGamesBrandsByGameType
// game_type to category_id:  Many-to-1
type SubGamesBrandsByGameType struct {
	Type string `json:"game_type"`

	GameVendorCategoryId int64          `json:"category_id"`
	SubGames             []BrandSubGame `json:"vendor,omitempty"`
}

type SubGameBrandsMap map[string][]BrandSubGame

func (m SubGameBrandsMap) AsSlice(gameTypeOrder map[string]int) (list []SubGamesBrandsByGameType, err error) {
	if gameTypeOrder == nil {
		return
	}
	for subGameType, subGames := range m {
		if len(subGames) == 0 {
			return list, errors.New("build SubGamesBrandsByGameType aborted. cannot accept empty subGames")
		}
		list = append(list, SubGamesBrandsByGameType{
			Type:                 subGameType,
			GameVendorCategoryId: subGames[0].GameVendorCategoryId,
			SubGames:             subGames,
		})
	}

	slices.SortFunc(list, func(a, b SubGamesBrandsByGameType) int {
		return gameTypeOrder[a.Type] - gameTypeOrder[b.Type]
	})
	return list, nil
}

// SubGameBrand for (service *SubGameService) List
type SubGameBrandRaw struct {
	ploutos.BASE
	Name          string    `json:"name" form:"name" gorm:"column:name"`
	VendorBrandId int64     `json:"vendor_brand_id" form:"vendor_brand_id" gorm:"column:vendor_brand_id"`
	WebIcon       string    `json:"web_icon" form:"web_icon" gorm:"column:web_icon"`
	AppIcon       string    `json:"app_icon" form:"app_icon" gorm:"column:app_icon"`
	Web           int64     `json:"web" form:"web" gorm:"column:web"`
	H5            int64     `json:"h5" form:"h5" gorm:"column:h5"`
	Ios           int64     `json:"ios" form:"ios" gorm:"column:ios"`
	Android       int64     `json:"android" form:"android" gorm:"column:android"`
	Status        int64     `json:"status" form:"status" gorm:"column:status"`
	BrandId       int64     `json:"brand_id" form:"brand_id" gorm:"column:brand_id"`
	SubGameId     int64     `json:"sub_game_id" form:"sub_game_id" gorm:"column:sub_game_id"`
	GameCode      string    `json:"game_code" form:"game_code" gorm:"column:game_code"`
	GameType      string    `json:"game_type" form:"game_type" gorm:"column:game_type"`
	IsFeatured    bool      `json:"is_featured" form:"is_featured" gorm:"column:is_featured"`
	StartTime     time.Time `json:"start_time" form:"start_time" gorm:"column:start_time"`
	EndTime       time.Time `json:"end_time" form:"end_time" gorm:"column:end_time"`
	SortRanking   int64     `json:"sort" form:"sort" gorm:"column:sort"`

	GameVendorBrand *ploutos.GameVendorBrand `json:"game_vendor_brand,omitempty" form:"-" gorm:"references:VendorBrandId;foreignKey:ID"`

	SubGame *ploutos.SubGameC `json:"game_vendor_brand,omitempty" form:"-" gorm:"references:SubGameId;foreignKey:ID"`
}

func BuildBrandSubGamesByGameType(sgbs []SubGameBrandRaw, gameTypeOrder map[string]int) ([]SubGamesBrandsByGameType, error) {
	m := make(SubGameBrandsMap)
	for _, sgb := range sgbs {
		gameType := sgb.GameType
		gvCategoryId := sgb.GameVendorBrand.CategoryId
		gvbMaintenanceStartTime, gvbMaintenanceEndTime := sgb.GameVendorBrand.StartTime, sgb.GameVendorBrand.EndTime

		var supportIFrame bool

		if sgb.SubGame != nil {
			supportIFrame = sgb.SubGame.SupportIframe
		}

		now := time.Now()
		var isGvbMaintenance bool
		if !sgb.GameVendorBrand.StartTime.IsZero() {
			if now.After(gvbMaintenanceStartTime) && (now.Before(gvbMaintenanceEndTime) || gvbMaintenanceEndTime.IsZero()) {
				isGvbMaintenance = true
			}
		}

		m[gameType] = append(m[gameType], BrandSubGame{
			GameVendorId:         sgb.GameVendorBrand.GameVendorId,
			GameVendorCategoryId: gvCategoryId,
			SubGameId:            sgb.SubGameId,
			Name:                 sgb.Name,
			Type:                 sgb.GameType,
			WebIcon:              Url(sgb.WebIcon),
			AppIcon:              Url(sgb.AppIcon),
			IsMaintenance:        isGvbMaintenance,
			MaintenanceStartUnix: gvbMaintenanceStartTime.Unix(),
			MaintenanceEndUnix:   gvbMaintenanceEndTime.Unix(),
			SortRanking:          sgb.SortRanking,
			SupportIFrame:        supportIFrame,
		})
	}

	return m.AsSlice(gameTypeOrder)
}
