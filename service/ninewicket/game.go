package ninewicket

import (
	"context"
	"log"

	"web-api/model"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/ninewickets/api"
)

func (n *NineWicket) GetGameUrl(ctx context.Context, user model.User, currency, gameCode, subGameCode string, platform int64, extra model.Extra) (url string, err error) {
	client, err := util.NineWicketFactory()
	if err != nil {
		return "", err
	}
	userId := api.UserId(user.ID)

	// fixme
	key, err := client.PlayerGetKey(userId, user.Username, api.CurrencyINR)
	if err != nil {
		log.Printf("Error getting 9Wicket key, err: %v ", err.Error())
	}

	url, err = client.PlayerGetLoginUrl(userId, &key, api.EventTypeCricket)
	if err != nil {
		log.Printf("Error getting 9Wicket game url, err: %v ", err.Error())
	}

	return url, err
}
