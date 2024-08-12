package model

import (
	ploutosFB "blgit.rfdev.tech/taya/ploutos-object/fb/model"
)

type FbOdds struct {
	ploutosFB.FbOdds

	RelatedOdds []ploutosFB.FbOdds `gorm:"foreignKey:SportsID,MatchID,MarketGroupType,MarketGroupPeriod,MarketlineValue;references:SportsID,MatchID,MarketGroupType,MarketGroupPeriod,MarketlineValue"`
}