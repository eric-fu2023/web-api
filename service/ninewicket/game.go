package ninewicket

import (
	"blgit.rfdev.tech/taya/game-service/ninewicket/api"
	"log"
	"web-api/model"
	"web-api/util"
)

func (n *NineWicket) GetGameUrl(user model.User, currency, gameCode, subGameCode string, platform int64, extra model.Extra) (url string, err error) {
	client := util.NineWicketFactory()

	key, err := client.PlayerGetKey(user.IdAsString(), user.Username, "testbainrma", api.CurrencyINR)
	if err != nil {
		log.Printf("Error getting 9Wicket key, err: %v ", err.Error())
	}
	url, err = client.PlayerGetLoginUrl(user.IdAsString(), &key, api.EventTypeCricket)

	if err != nil {
		log.Printf("Error getting 9Wicket game url, err: %v ", err.Error())
	}

	return url, err
}
