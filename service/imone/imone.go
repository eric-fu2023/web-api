package imone

import (
	"context"

	"web-api/model"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

type _gameUserInterface interface {
	CreateWallet(model.User, string) error
	TransferFrom(*gorm.DB, model.User, string, string, int64, model.Extra) error
	TransferTo(*gorm.DB, model.User, ploutos.UserSum, string, string, int64, model.Extra) (int64, error)
	GetGameUrl(context.Context, model.User, string, string, string, int64, model.Extra) (string, error)
	GetGameBalance(model.User, string, string, model.Extra) (int64, error)
}

type ImOne struct {
}

var _ _gameUserInterface = &ImOne{}
