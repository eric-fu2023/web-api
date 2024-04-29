package imone

import (
	"errors"
	"fmt"

	"web-api/model"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/imone"
)

// InitImOneFactory TODO:GAMEINTEGRATIONIMONE
const (
	_ int64 = iota
	_argplatformpc
	_argh5
	_argandroid
	_argios
)

// GetGameUrl TODO:GAMEINTEGRATIONIMONE
func (c *ImOne) GetGameUrl(user model.User, _, gameCode, subGameCode string, _ int64, extra model.Extra) (string, error) {
	productWalletCode := tayaGameCodeToImOneWalletCodeMapping[gameCode]
	if productWalletCode != imone.WalletCodePlayTech {
		return "", errors.New(fmt.Sprintf("GetGameUrl wallet type not supported by imone API v7.0. tayagameCode: %s, imonewalletCode: %s ", gameCode, productWalletCode))
	}

	client := util.ImOneFactory()
	return client.NewLaunchMobileGame(subGameCode, extra.Locale, extra.Ip, productWalletCode, "")
}
