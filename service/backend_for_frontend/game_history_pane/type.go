package game_history_pane

import (
	"fmt"
	"sync"
	"web-api/model"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type GamesHistoryPaneType = int64

const (
	GamesPaneDefault    GamesHistoryPaneType = 0
	GamesPaneAll                             = GamesPaneDefault
	GamesPaneType1      GamesHistoryPaneType = 1
	GamesPaneTypeSports                      = GamesPaneType1
	GamesPaneType2      GamesHistoryPaneType = 2
	GamesPaneTypeCasino                      = GamesPaneType2
	_                   GamesHistoryPaneType = 3 // reserved for gift
)

// GamePaneHistoryTypes derived from keys of [gameVendorIdsByPaneType_baseline]
var GamePaneHistoryTypes = sync.OnceValue(
	func() (paneTypes []GamesHistoryPaneType) {
		for paneType, _ := range gameVendorIdsByPaneType_baseline {
			paneTypes = append(paneTypes, paneType)
		}
		return paneTypes
	})

// gameVendorIdsByPaneType_baseline to be read in conjunction GetGameVendorIdsByPaneType for completeness
var gameVendorIdsByPaneType_baseline = map[GamesHistoryPaneType][]int64{
	GamesPaneAll: {},
	/*
		Equivalence  Oct 2024
				AHA:
					GamesPaneTypeSports: {ploutos.GAME_FB, ploutos.GAME_SABA, ploutos.GAME_TAYA, ploutos.GAME_IMSB, ploutos.GAME_DB_SPORT},
				Batace:
					GamesPaneTypeSports: {NA, NA , ploutos.GAME_SelfSports, ploutos.GAME_InplayMatrixSportsbook, NA},
	*/
	GamesPaneTypeSports: {ploutos.GAME_FB, ploutos.GAME_SABA, ploutos.GAME_TAYA, ploutos.GAME_IMSB, ploutos.GAME_DB_SPORT},

	/*
		Equivalence Oct 2024
				AHA:
					GamesPaneTypeCasino: {ploutos.GAME_HACKSAW, ploutos.GAME_DOLLAR_JACKPOT, ploutos.GAME_STREAM_GAME},
				Batace:
					GamesPaneTypeCasino: {NA, ploutos.GAME_DollarJackpot, NA},
	*/
	GamesPaneTypeCasino: {ploutos.GAME_HACKSAW, ploutos.GAME_DOLLAR_JACKPOT, ploutos.GAME_STREAM_GAME},
}

func GetGameVendorIdsByPaneType(pType GamesHistoryPaneType) ([]int64, error) {
	if pType == 3 {
		return []int64{}, fmt.Errorf("3 is reserved for gift (not a game / bet report feature)")
	}

	gameVendorIds, ok := gameVendorIdsByPaneType_baseline[pType]
	if !ok {
		gameVendorIds = []int64{}
	}

	// additions
	switch pType {
	case GamesPaneAll:
		forSports, _ := GetGameVendorIdsByPaneType(GamesPaneTypeSports)
		forCasino, _ := GetGameVendorIdsByPaneType(GamesPaneTypeCasino)
		return append(forSports, forCasino...), nil
	case GamesPaneTypeCasino: // assumes all game vendors of game integrations are of the same type and include them in gameVendorIds
		var gi []ploutos.GameIntegration
		err := model.DB.Model(ploutos.GameIntegration{}).Preload(`GameVendors`).Find(&gi).Error
		if err != nil {
			return []int64{}, err
		}
		for _, g := range gi {
			for _, v := range g.GameVendors {
				gameVendorIds = append(gameVendorIds, v.ID)
			}
		}
	}

	return gameVendorIds, nil
}
