package imone

import (
	"blgit.rfdev.tech/taya/game-service/imone"
	"errors"
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

func (c *ImOne) GetGameUrl(user model.User, currency, tayaGameCode, tayaSubGameCode string, _ int64, extra model.Extra) (string, error) {
	productWalletCode, exist := tayaGameCodeToImOneWalletCodeMapping[tayaGameCode]
	if !exist {
		return "", ErrGameCodeMapping
	}

	client := util.ImOneFactory()
	userId := user.IdAsString()
	exists, err := client.CheckUserExist(userId)
	if err != nil {
		return "", err
	}

	if !exists {
		createUserErr := client.CreateUser(userId, currency, defaultPassword, "")
		if createUserErr != nil && !errors.As(createUserErr, &imone.ErrCreateUserAlreadyExists{}) {
			return "", createUserErr
		}
	}

	return client.NewLaunchMobileGame(tayaSubGameCode, extra.Locale, extra.Ip, productWalletCode, "", userId)
}
