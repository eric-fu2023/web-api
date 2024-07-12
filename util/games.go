package util

import (
	"os"

	"web-api/conf/consts"

	gameservicecommon "blgit.rfdev.tech/taya/game-service/common"
	"blgit.rfdev.tech/taya/game-service/dc"
	"blgit.rfdev.tech/taya/game-service/evo"
	"blgit.rfdev.tech/taya/game-service/fb"
	"blgit.rfdev.tech/taya/game-service/imone"
	"blgit.rfdev.tech/taya/game-service/imsb"
	"blgit.rfdev.tech/taya/game-service/ninewickets"
	"blgit.rfdev.tech/taya/game-service/saba"
	"blgit.rfdev.tech/taya/game-service/ugs"
	ploutosmodel "blgit.rfdev.tech/taya/ploutos-object"
)

const (
	IntegrationIdUGS        = ploutosmodel.GAME_INTEGRATION_UGS
	IntegrationIdImOne      = ploutosmodel.GAME_INTEGRATION_IMONE
	IntegrationIdEvo        = ploutosmodel.GAME_INTEGRATION_EVO
	IntegrationIdNineWicket = ploutosmodel.GAME_INTEGRATION_NINEWICKETS
	IntegrationIdMumbai     = ploutosmodel.GAME_INTEGRATION_MUMBAI
)

var (
	TayaFactory       fb.FB
	FBFactory         fb.FB
	SabaFactory       saba.Saba
	DCFactory         dc.Dc
	IMFactory         imsb.IM
	UgsFactory        ugs.UGS
	EvoFactory        evo.EVO
	NineWicketFactory func() (ninewickets.ClientOperations, error)
	ImOneFactory      func() imone.GeneralApi
	// TODO @Seng
)

var VendorIdToGameClient = make(map[int64]gameservicecommon.TransferWalletInterface)

func InitTayaFactory() {
	TayaFactory = fb.FB{
		BaseUrl:           os.Getenv("GAME_TAYA_BASE_URL"),
		MerchantId:        os.Getenv("GAME_TAYA_MERCHANT_ID"),
		MerchantApiSecret: os.Getenv("GAME_TAYA_MERCHANT_API_SECRET"),
		IsSandbox:         true,
	}
	VendorIdToGameClient[consts.GameVendor["taya"]] = TayaFactory
}

func InitFbFactory() {
	FBFactory = fb.FB{
		BaseUrl:           os.Getenv("GAME_FB_BASE_URL"),
		MerchantId:        os.Getenv("GAME_FB_MERCHANT_ID"),
		MerchantApiSecret: os.Getenv("GAME_FB_MERCHANT_API_SECRET"),
		IsSandbox:         true,
	}
	VendorIdToGameClient[consts.GameVendor["fb"]] = FBFactory
}

func InitSabaFactory() {
	SabaFactory = saba.Saba{
		BaseUrl:    os.Getenv("GAME_SABA_BASE_URL"),
		VendorId:   os.Getenv("GAME_SABA_VENDOR_ID"),
		OperatorId: os.Getenv("GAME_SABA_OPERATOR_ID"),
		IsSandbox:  true,
	}
}

func InitDcFactory() {
	DCFactory = dc.Dc{
		BaseUrl:   os.Getenv("GAME_DC_BASE_URL"),
		BrandId:   os.Getenv("GAME_DC_BRAND_ID"),
		ApiKey:    os.Getenv("GAME_DC_API_KEY"),
		IsSandbox: true,
	}
}

func InitImFactory() {
	IMFactory = imsb.IM{
		BaseUrl:        os.Getenv("GAME_IMSB_BASE_URL"),
		AccessCode:     os.Getenv("GAME_IMSB_ACCESS_CODE"),
		CommonWalletIv: os.Getenv("GAME_IMSB_COMMON_WALLET_IV"),
	}
}

func InitUgsFactory() {
	UgsFactory = ugs.UGS{
		BaseUrl:      os.Getenv("GAME_UGS_BASE_URL"),
		ClientId:     os.Getenv("GAME_UGS_CLIENT_ID"),
		ClientSecret: os.Getenv("GAME_UGS_CLIENT_SECRET"),
	}

}

func InitImOneFactory() {
	baseUrl := os.Getenv("GAME_IMONE_BASE_URL")
	merchantCode := os.Getenv("GAME_IMONE_MERCHANT_CODE")
	prefix := os.Getenv("GAME_IMONE_PLAYER_PREFIX")

	ImOneFactory = imone.NewGeneralService(baseUrl, merchantCode, imone.NewPrefixedPlayer(prefix))
}

func InitEvoFactory() {
	EvoFactory = evo.EVO{
		Host:                  os.Getenv("GAME_EVO_HOST"),
		CasinoId:              os.Getenv("GAME_EVO_CASINO_ID"),
		UA2Token:              os.Getenv("GAME_EVO_UA2_TOKEN"),
		ECToken:               os.Getenv("GAME_EVO_EC_TOKEN"),
		GameHistoryApiToken:   os.Getenv("GAME_EVO_HISTORY_API_TOKEN"),
		ExternalLobbyApiToken: os.Getenv("GAME_EVO_LOBBY_API_TOKEN"),
	}
}

func InitNineWicketsFactory() {
	cert := os.Getenv("GAME_NINE_WICKETS_CERT")
	initPrivateDomain := os.Getenv("GAME_NINE_WICKETS_DOMAIN")
	website := os.Getenv("GAME_NINE_WICKETS_WEBSITE")
	agentId := os.Getenv("GAME_NINE_WICKETS_AGENT_ID")
	apiServerHost := os.Getenv("GAME_NINE_WICKETS_API_HOST")
	exchHost := os.Getenv("GAME_NINE_WICKETS_EX_HOST")

	NineWicketFactory = ninewickets.NewClientFactory(cert, initPrivateDomain, website, apiServerHost, exchHost, agentId, true)
}
