package imone

import (
	"web-api/model"
	"web-api/util"
)

const (
	_ int64 = iota
	_argplatformpc
	_argh5
	_argandroid
	_argios
)

func (c *ImOne) GetGameUrl(user model.User, _, gameCode, subGameCode string, _ int64, extra model.Extra) (string, error) {
	productWalletCode := tayaGameCodeToImOneWalletCodeMapping[gameCode]

	client := util.ImOneFactory()
	return client.NewLaunchMobileGame(subGameCode, extra.Locale, extra.Ip, productWalletCode, "")
}
