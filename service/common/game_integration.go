package common

import (
	"context"
	"os"

	"web-api/model"
	"web-api/service/crown_valexy"
	"web-api/service/evo"
	"web-api/service/imone"
	"web-api/service/mancala"
	"web-api/service/mumbai"
	"web-api/service/ninewicket"
	"web-api/service/ugs"
	"web-api/util"

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
	util.IntegrationIdCrownValexy: &crown_valexy.CrownValexy{},
	util.IntegrationIdMancala:     &mancala.Mancala{},
}

// GameIntegrationInterface
type GameIntegrationInterface interface {
	CreateWallet(context.Context, model.User, string) error

	// TransferFrom
	// SHOULD:
	//	execute a debit action from third party account. if account not found consider a zero withdraw and return nil
	//  the db/tx instance should be orchestrated by caller
	TransferFrom(context.Context, *gorm.DB, model.User, string, string, int64, model.Extra) error
	TransferTo(context.Context, *gorm.DB, model.User, ploutos.UserSum, string, string, int64, model.Extra) (int64, error)
	GetGameUrl(context.Context, model.User, string, string, string, int64, model.Extra) (string, error)
	GetGameBalance(context.Context, model.User, string, string, model.Extra) (int64, error)
}
