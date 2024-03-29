package common

import (
	"web-api/model"
	"web-api/service/ugs"
)

var GameIntegration = map[int64]GameIntegrationInterface{
	1: ugs.UGS{},
}

type GameIntegrationInterface interface {
	CreateWallet(model.User, string) error
}
