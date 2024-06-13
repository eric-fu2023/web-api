package util

import (
	"blgit.rfdev.tech/taya/game-service/ninewicket"
	"fmt"
	"os"

	"web-api/conf/consts"

	gameservicecommon "blgit.rfdev.tech/taya/game-service/common"
	"blgit.rfdev.tech/taya/game-service/dc"
	"blgit.rfdev.tech/taya/game-service/evo"
	"blgit.rfdev.tech/taya/game-service/fb"
	"blgit.rfdev.tech/taya/game-service/imone"
	"blgit.rfdev.tech/taya/game-service/imsb"
	"blgit.rfdev.tech/taya/game-service/saba"
	"blgit.rfdev.tech/taya/game-service/ugs"
)

const (
	IntegrationIdUGS        = 1
	IntegrationIdImOne      = 2
	IntegrationIdEvo        = 3
	IntegrationIdNineWicket = 4
)

var (
	TayaFactory       fb.FB
	FBFactory         fb.FB
	SabaFactory       saba.Saba
	DCFactory         dc.Dc
	IMFactory         imsb.IM
	UgsFactory        ugs.UGS
	EvoFactory        evo.EVO
	NineWicketFactory func() ninewicket.ClientOperations
	ImOneFactory      func() imone.GeneralApi
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

// ImonePlayer FIXME: move to relevant pkg
type ImonePlayer struct {
	BaseId string
	Prefix string
}

func (p ImonePlayer) Token() string {
	//TODO implement me
	return ""
}

func (p ImonePlayer) SetToken(token string) {
	//TODO implement me
}

func (p ImonePlayer) Id() string {
	return p.Prefix + fmt.Sprintf("%08s", p.BaseId)
}

func NewPlayer(prefix string) func(string) imone.Playable {
	return func(baseId string) imone.Playable {
		return &ImonePlayer{
			BaseId: baseId,
			Prefix: prefix,
		}
	}
}

func InitImOneFactory() {
	baseUrl := os.Getenv("GAME_IMONE_BASE_URL")
	merchantCode := os.Getenv("GAME_IMONE_MERCHANT_CODE")
	prefix := os.Getenv("GAME_IMONE_PLAYER_PREFIX")

	ImOneFactory = imone.NewGeneralService(baseUrl, merchantCode, NewPlayer(prefix))
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

func InitNineWicketFactory() {
	//NineWicketFactory = ninewicket.NineWicket{
	//	ApiServerHost: os.Getenv("GAME_NINE_WICKET_API_HOST"),
	//	ExchangeHost:  os.Getenv("GAME_NINE_WICKET_EX_HOST"),
	//	Domain:        os.Getenv("GAME_NINE_WICKET_DOMAIN"),
	//	Cert:          os.Getenv("GAME_NINE_WICKET_CERT"),
	//}

	f := ninewicket.NewClientFactory(os.Getenv("GAME_NINE_WICKET_CERT"), os.Getenv("GAME_NINE_WICKET_DOMAIN"), os.Getenv("GAME_NINE_WICKET_WEBSITE"))
	NineWicketFactory = func() ninewicket.ClientOperations {
		client := f()
		d, _ := client.GetDomains()
		client.SetDomains(d.Domains)
		client.SetPrivateDomains(d.PrivateDomains)
		return client
	}
}
