package imone

import (
	"errors"

	"web-api/model"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"

	"gorm.io/gorm"
)

var ErrImOneRegister = errors.New("register user with imone failed")
var ErrOthers = errors.New("imone create user failed")

type _userRegistration interface {
	CreateUser(model.User, string) error
	VendorRegisterError() error
	OthersError() error
}

type _gameUserInterface interface {
	CreateWallet(model.User, string) error
	TransferFrom(*gorm.DB, model.User, string, string, int64, model.Extra) error
	TransferTo(*gorm.DB, model.User, ploutos.UserSum, string, string, int64, model.Extra) (int64, error)
	GetGameUrl(model.User, string, string, string, int64, model.Extra) (string, error)
	GetGameBalance(model.User, string, string, model.Extra) (int64, error)
}

type imoner interface {
	_userRegistration
	_gameUserInterface
}

type ImOne struct {
}

func (c *ImOne) VendorRegisterError() error {
	return ErrImOneRegister
}

func (c *ImOne) OthersError() error {
	return ErrOthers
}

var _ imoner = &ImOne{}
