package evo

import (
	"os"
	"web-api/model"
	"web-api/util"

	"log"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EVO struct {
}

// @Description
func (e EVO) CreateWallet(user model.User, currency string) (err error) {
	//
	return nil
}

func (e EVO) GetGameUrl(user model.User, currency, gameCode, subGameCode string, platform int64, extra model.Extra) (url string, err error) {
	client := util.EvoFactory.NewClient()

	uuid := uuid.NewString()

	response, err := client.GetGameUrl(uuid, "en-GB", user.IdAsString(), currency, "1.233.1", extra.Ip, subGameCode)

	if err != nil {
		log.Printf("Error getting evo game url, err: %v ", err.Error())
	}

	url = os.Getenv("GAME_EVO_HOST") + response.EntryEmebedded

	return url, err
}

func (e EVO) GetGameBalance(user model.User, currency, gameCode string, extra model.Extra) (balance int64, err error) {
	return 0, nil
}

func (e EVO) TransferFrom(tx *gorm.DB, user model.User, currency, gameCode string, gameVendorId int64, extra model.Extra) (err error) {
	return nil
}

func (e EVO) TransferTo(tx *gorm.DB, user model.User, sum ploutos.UserSum, currency, gameCode string, gameVendorId int64, extra model.Extra) (balance int64, err error) {
	return 0, nil
}
