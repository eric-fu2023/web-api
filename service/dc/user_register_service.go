package dc

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"gorm.io/gorm"
	"web-api/conf/consts"
	"web-api/model"
)

var ErrVendorRegister = errors.New("register user with dc failed")
var ErrOthers = errors.New("dc create user failed")

type UserRegister struct {
}

func (c *UserRegister) CreateUser(user model.User, currency string) (err error) {
	gpu := ploutos.GameVendorUser{
		GameVendorId:     consts.GameVendor["dc"],
		UserId:           user.ID,
		ExternalUserId:   user.Username,
		ExternalCurrency: currency,
	}
	err = model.DB.Save(&gpu).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			err = c.VendorRegisterError()
		}
		return
	}
	return
}

func (c *UserRegister) VendorRegisterError() error {
	return ErrVendorRegister
}

func (c *UserRegister) OthersError() error {
	return ErrOthers
}
