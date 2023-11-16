package saba

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"errors"
	"fmt"
	"os"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/util"
)

var ErrVendorRegister = errors.New("register user with saba failed")
var ErrOthers = errors.New("saba create user failed")

type UserRegister struct {
}

func (c *UserRegister) CreateUser(user model.User, currency string) (err error) {
	client := util.SabaFactory.NewClient()
	res, e := client.CreateMember(user.Username, currency, os.Getenv("GAME_SABA_ODDS_TYPE"))
	if e != nil {
		err = fmt.Errorf("%w: %w", ErrVendorRegister, err)
		return
	}
	gpu := ploutos.GameVendorUser{
		GameVendorId:     consts.GameVendor["saba"],
		UserId:           user.ID,
		ExternalUserId:   user.Username,
		ExternalCurrency: currency,
		ExternalId:       res,
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
