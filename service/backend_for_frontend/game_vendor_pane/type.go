package game_vendor_pane

import ploutos "blgit.rfdev.tech/taya/ploutos-object"

type GamesPaneType = int64

const (
	GamesPaneType1      GamesPaneType = 1
	GamesPaneTypeSports               = GamesPaneType1
	GamesPaneType2      GamesPaneType = 2
	GamesPaneTypeCasino               = GamesPaneType2
)

var GameVendorIdsByPaneType = map[int64][]int64{
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
