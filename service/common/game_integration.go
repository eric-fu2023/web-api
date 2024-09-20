package common

import (
	"context"
	"errors"
	"os"

	"web-api/model"
	"web-api/service/evo"
	"web-api/service/imone"
	"web-api/service/mumbai"
	"web-api/service/ninewicket"
	"web-api/service/ugs"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/mancala/api"
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
	util.IntegrationIdMancala: &Mancala{},
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

func (c *CrownValexy) GetGameUrl(user model.User, s string, s2 string, s3 string, i int64, extra model.Extra) (string, error) {
	//TODO implement me
	return "", errors.New("todo")
}

func (c *CrownValexy) GetGameBalance(user model.User, s string, s2 string, extra model.Extra) (int64, error) {
	//TODO implement me
	return 0, errors.New("todo")
}

type Mancala struct{}

func (m *Mancala) CreateWallet(user model.User, tayaCurrency string) error {
	//TODO implement me
	var currencyMancala api.Currency
	if tayaCurrency == "INR" {
		currencyMancala = "INR"
	} else {
		return errors.New("mancala invalid currency")
	}

	// FIXME password to be derived from user instead of default value
	go func() {
		// fire and forget. later calls should follow up with user creation, if needed.
		service, err := util.MancalaFactory()
		if err == nil {
			_, _, _ = service.AddTransferWallet(context.TODO(), user.IdAsString(), currencyMancala)
		}
	}()

	return model.DB.Transaction(func(tx *gorm.DB) (err error) {
		var gameVendors []ploutos.GameVendor
		err = tx.Model(ploutos.GameVendor{}).
			Where(`game_vendor.game_integration_id`, util.IntegrationIdMancala).Find(&gameVendors).Error
		if err != nil {
			return
		}

		for _, gameVendor := range gameVendors {
			gvu := ploutos.GameVendorUser{
				GameVendorId:     gameVendor.ID,
				UserId:           user.ID,
				ExternalUserId:   user.IdAsString(),
				ExternalCurrency: currencyMancala,
			}

			err = tx.Create(&gvu).Error
			if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
				return
			}
		}
		return
	})
}

func (m *Mancala) TransferFrom(db *gorm.DB, user model.User, s string, s2 string, i int64, extra model.Extra) error {
	//TODO implement me
	return errors.New("todo")
}

func (m *Mancala) TransferTo(db *gorm.DB, user model.User, sum ploutos.UserSum, s string, s2 string, i int64, extra model.Extra) (int64, error) {
	//TODO implement me
	return 0, errors.New("todo")
}

func (m *Mancala) GetGameUrl(user model.User, s string, s2 string, s3 string, i int64, extra model.Extra) (string, error) {
	//TODO implement me
	return "", errors.New("todo")
}

func (m *Mancala) GetGameBalance(user model.User, s string, s2 string, extra model.Extra) (int64, error) {
	//TODO implement me
	return 0, errors.New("todo")
}

type GameIntegrationInterface interface {
	CreateWallet(model.User, string) error
	TransferFrom(*gorm.DB, model.User, string, string, int64, model.Extra) error
	TransferTo(*gorm.DB, model.User, ploutos.UserSum, string, string, int64, model.Extra) (int64, error)
	GetGameUrl(model.User, string, string, string, int64, model.Extra) (string, error)
	GetGameBalance(model.User, string, string, model.Extra) (int64, error)
}
