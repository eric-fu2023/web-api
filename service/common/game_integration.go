package common

import (
	"web-api/model"
	"web-api/service/evo"
	"web-api/service/imone"
	"web-api/service/ugs"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

var GameIntegration = map[int64]GameIntegrationInterface{
	util.IntegrationIdUGS:   ugs.UGS{},
	util.IntegrationIdImOne: &imone.ImOne{},
	util.IntegrationIdEvo:   evo.EVO{},
}

// to delete if unused
type GameIntegrationNoop struct {
}

func (g GameIntegrationNoop) CreateWallet(user model.User, s string) error {
	return nil
}

func (g GameIntegrationNoop) TransferFrom(db *gorm.DB, user model.User, s string, s2 string, i int64, extra model.Extra) error {
	return nil
}

func (g GameIntegrationNoop) TransferTo(db *gorm.DB, user model.User, sum ploutos.UserSum, s string, s2 string, i int64, extra model.Extra) (int64, error) {
	return 0, nil
}

func (g GameIntegrationNoop) GetGameUrl(user model.User, s string, s2 string, s3 string, i int64, extra model.Extra) (string, error) {
	return "gameintegration/url", nil
}

func (g GameIntegrationNoop) GetGameBalance(user model.User, s string, s2 string, extra model.Extra) (int64, error) {
	return 0, nil
}

type GameIntegrationInterface interface {
	CreateWallet(model.User, string) error
	TransferFrom(*gorm.DB, model.User, string, string, int64, model.Extra) error
	TransferTo(*gorm.DB, model.User, ploutos.UserSum, string, string, int64, model.Extra) (int64, error)
	GetGameUrl(model.User, string, string, string, int64, model.Extra) (string, error)
	GetGameBalance(model.User, string, string, model.Extra) (int64, error)
}
