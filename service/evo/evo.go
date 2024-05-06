package evo

import (
	"os"
	"web-api/model"
	"web-api/util"

	"log"

	"github.com/google/uuid"
)

type EVO struct {
}

// @Description
func (e *EVO) CreateWallet() (err error) {
	//
	return nil
}

func (e *EVO) GetGameUrl(user model.User, currency, gameCode, subGameCode string, platform int64, extra model.Extra) (url string, err error) {
	client := util.EvoFactory.NewClient()

	uuid := uuid.NewString()

	response, err := client.GetGameUrl(uuid, "en-GB", user.IdAsString(), currency, "1.233.1", extra.Ip)

	if err != nil {
		log.Printf("Error getting evo game url, err: ", err)
	}

	url = os.Getenv("GAME_EVO_HOST") + response.EntryEmebedded

	return url, err
}
