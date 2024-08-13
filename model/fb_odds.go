package model

import (
	ploutosFB "blgit.rfdev.tech/taya/ploutos-object/fb/model"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

type FbOdds struct {
	ploutosFB.FbOdds

	RelatedOdds []ploutosFB.FbOdds `gorm:"foreignKey:SportsID,MatchID,MarketGroupType,MarketGroupPeriod,MarketlineValue;references:SportsID,MatchID,MarketGroupType,MarketGroupPeriod,MarketlineValue"`
	FbOddsOrderRequest FbOddsOrderRequest `gorm:"foreignKey:SportsID,MatchID,MarketGroupType,MarketGroupPeriod,MarketlineValue;references:SportsID,MatchID,MarketGroupType,MarketGroupPeriod,MarketlineValue"`
}

type FbOddsOrderRequest struct {
	ploutosFB.FbOddsOrderRequest

	FbBetReport ploutos.FbBetReport `gorm:"foreignKey:OrderID;references:BusinessId"`
}