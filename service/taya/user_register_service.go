package taya

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"fmt"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/util"
)

var ErrVendorRegister = errors.New("register user with taya failed")
var ErrOthers = errors.New("taya create user failed")

type UserRegister struct {
}

func (c *UserRegister) CreateUser(user model.User, currency string) (err error) {
	client := util.TayaFactory.NewClient()
	res, e := client.CreateUser(user.Username, []string{}, 0)
	if e != nil {
		err = fmt.Errorf("%w: %w", ErrVendorRegister, err)
		return
	}
	gpu := ploutos.GameVendorUser{
		GameVendorId:     consts.GameVendor["taya"],
		UserId:           user.ID,
		ExternalUserId:   user.Username,
		ExternalCurrency: currency,
		ExternalId:       fmt.Sprintf("%d", res),
	}
	err = model.DB.Save(&gpu).Error
	if err != nil {
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
