package common

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
	"web-api/model"
	"web-api/service/ugs"
)

var GameIntegration = map[int64]GameIntegrationInterface{
	1: ugs.UGS{},
}

type GameIntegrationInterface interface {
	CreateWallet(model.User, string) error
	TransferFrom(*gorm.DB, model.User, string, string, int64, model.Extra) error
	TransferTo(*gorm.DB, model.User, ploutos.UserSum, string, string, int64, model.Extra) (int64, error)
	GetGameUrl(model.User, string, string, string, int64, model.Extra) (string, error)
	GetGameBalance(model.User, string, string, model.Extra) (int64, error)
}
