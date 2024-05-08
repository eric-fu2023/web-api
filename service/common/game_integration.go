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

type GameIntegrationInterface interface {
	CreateWallet(model.User, string) error
	TransferFrom(*gorm.DB, model.User, string, string, int64, model.Extra) error
	TransferTo(*gorm.DB, model.User, ploutos.UserSum, string, string, int64, model.Extra) (int64, error)
	GetGameUrl(model.User, string, string, string, int64, model.Extra) (string, error)
	GetGameBalance(model.User, string, string, model.Extra) (int64, error)
}
