package imone

import (
	"errors"
	"fmt"
	"time"

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

// PlatformClientToImOneMapping FIXME mapping values
var PlatformClientToImOneMapping = map[int64]string{
	_argplatformpc: "pc",
	_argh5:         "mobile",
	_argandroid:    "mobile",
	_argios:        "mobile",
}

// GetGameUrl TODO:GAMEINTEGRATIONIMONE

func (c *ImOne) GetGameUrl(user model.User, _, gameCode, subGameCode string, platform int64, extra model.Extra) (string, error) {
	productWalletCode := tayaGameCodeToImOneWalletCodeMapping[gameCode]
	if productWalletCode != imone.WalletCodePlayTech {
		return "", errors.New(fmt.Sprintf("GetGameUrl wallet type not supported by imone API v7.0. tayagameCode: %s, imonewalletCode: %s ", gameCode, productWalletCode))
	}

	// PlatformClientToImOneMapping FIXME mapping values
	platformStr, _ := PlatformClientToImOneMapping[platform]

	client := util.ImOneFactory()

	startDate := time.Now()
	endDate := startDate.Add(time.Second)
	userId := user.IdAsString()

	client.PlayTechGetPlayerGameUrl(userId, startDate, endDate, subGameCode, platformStr, productWalletCode)

	return fmt.Sprintf("dummygameurl/%s/%s/%s", userId, platformStr, subGameCode), nil
}
