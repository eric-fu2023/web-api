package common

import (
	"context"
	"errors"
	"os"

	"web-api/model"
	"web-api/service/evo"
	"web-api/service/imone"
	"web-api/service/mancala"
	"web-api/service/mumbai"
	"web-api/service/ninewicket"
	"web-api/service/ugs"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

var GameIntegration = map[int64]GameIntegrationInterface{
	util.IntegrationIdUGS:        ugs.UGS{},
	util.IntegrationIdImOne:      &imone.ImOne{},
	util.IntegrationIdEvo:        evo.EVO{},
	util.IntegrationIdNineWicket: &ninewicket.NineWicket{},
	util.IntegrationIdMumbai: &mumbai.Mumbai{
		Merchant: os.Getenv("GAME_MUMBAI_MERCHANT_CODE"),
		Agent:    os.Getenv("GAME_MUMBAI_AGENT_CODE"),
	},

	// TODO
	//util.IntegrationIdCrownValexy: &CrownValexy{},
	util.IntegrationIdMancala: &mancala.Mancala{},
}

type CrownValexy struct{}

func (c *CrownValexy) CreateWallet(user model.User, s string) error {
	//TODO implement me
	return errors.New("todo")
}

func (c *CrownValexy) TransferFrom(db *gorm.DB, user model.User, s string, s2 string, i int64, extra model.Extra) error {
	//TODO implement me
	return errors.New("todo")
}

func (c *CrownValexy) TransferTo(db *gorm.DB, user model.User, sum ploutos.UserSum, s string, s2 string, i int64, extra model.Extra) (int64, error) {
	//TODO implement me
	return 0, errors.New("todo")
}

func (c *CrownValexy) GetGameUrl(ctx context.Context, user model.User, s string, s2 string, s3 string, i int64, extra model.Extra) (string, error) {
	//TODO implement me
	return "", errors.New("todo")
}

func (c *CrownValexy) GetGameBalance(user model.User, s string, s2 string, extra model.Extra) (int64, error) {
	//TODO implement me
	return 0, errors.New("todo")
}

type GameIntegrationInterface interface {
	CreateWallet(model.User, string) error
	TransferFrom(*gorm.DB, model.User, string, string, int64, model.Extra) error
	TransferTo(*gorm.DB, model.User, ploutos.UserSum, string, string, int64, model.Extra) (int64, error)
	GetGameUrl(context.Context, model.User, string, string, string, int64, model.Extra) (string, error)
	GetGameBalance(model.User, string, string, model.Extra) (int64, error)
}
