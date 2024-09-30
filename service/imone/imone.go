package imone

import (
	"context"

	"web-api/model"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

type _gameUserInterface interface {
	CreateWallet(context.Context, model.User, string) error
	TransferFrom(context.Context, *gorm.DB, model.User, string, string, int64, model.Extra) error
	TransferTo(context.Context, *gorm.DB, model.User, ploutos.UserSum, string, string, int64, model.Extra) (int64, error)
	GetGameUrl(context.Context, model.User, string, string, string, int64, model.Extra) (string, error)
	GetGameBalance(context.Context, model.User, string, string, model.Extra) (int64, error)
}

type ImOne struct {
}

var _ _gameUserInterface = &ImOne{}
