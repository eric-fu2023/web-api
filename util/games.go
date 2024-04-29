package util

import (
	"os"
	"web-api/conf/consts"

	gameservicecommon "blgit.rfdev.tech/taya/game-service/common"
	"blgit.rfdev.tech/taya/game-service/dc"
	"blgit.rfdev.tech/taya/game-service/fb"
	"blgit.rfdev.tech/taya/game-service/imone"
	"blgit.rfdev.tech/taya/game-service/imsb"
	"blgit.rfdev.tech/taya/game-service/saba"
	"blgit.rfdev.tech/taya/game-service/ugs"
)

const (
	IntegrationIdUGS   = 1
	IntegrationIdImOne = 2
)

var (
	TayaFactory  fb.FB
	FBFactory    fb.FB
	SabaFactory  saba.Saba
	DCFactory    dc.Dc
	IMFactory    imsb.IM
	UgsFactory   ugs.UGS
	ImOneFactory func() *imone.ImOne
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
	baseUrl := os.Getenv("GAME_IMONE_BASE_URL")                // baseUrl
	merchantCode := os.Getenv("GAME_IMONE_BASE_MERCHANT_CODE") // merchantCode
	prefix := os.Getenv("GAME_IMONE_BASE_PLAYER_PREFIX")       // merchantCode

	ImOneFactory = imone.NewFactory(baseUrl, merchantCode, imone.NewDefaultPlayer(prefix))
}
