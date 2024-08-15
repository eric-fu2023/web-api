package model

import (
	"fmt"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	ploutosFB "blgit.rfdev.tech/taya/ploutos-object/fb/model"

	gameApi "blgit.rfdev.tech/taya/game-service/fb2/client/api"
)

type FbOdds struct {
	ploutosFB.FbOdds

	RelatedOdds []ploutosFB.FbOdds `gorm:"foreignKey:SportsID,MatchID,MarketGroupType,MarketGroupPeriod,MarketlineValue;references:SportsID,MatchID,MarketGroupType,MarketGroupPeriod,MarketlineValue"`
	FbOddsOrderRequestList []FbOddsOrderRequest `gorm:"foreignKey:SportsID,MatchID,MarketGroupType,MarketGroupPeriod,MarketlineValue;references:SportsID,MatchID,MarketGroupType,MarketGroupPeriod,MarketlineValue"`
}

type FbOddsOrderRequest struct {
	ploutosFB.FbOddsOrderRequest

	FbBetReport FbBetReport `gorm:"foreignKey:OrderID;references:BusinessId"`
}

type FbBetReport ploutos.FbBetReport

func (r FbBetReport) SettledStatus() (gameApi.SettledStatus, error) {
	reportInfo, uErr := ploutos.UnmarshalFBBetRawS(r.InfoJson)
	if uErr != nil {
		return gameApi.SettledStatusUnsettle, fmt.Errorf("unmarshal report history for SettledStatus fail %v", uErr)
	}

	return gameApi.SettledStatus(reportInfo.SettleStatus), nil
}

func (r FbBetReport) Outcome() (gameApi.Outcome, error) {
	reportInfo, uErr := ploutos.UnmarshalFBBetRawS(r.InfoJson)
	if uErr != nil {
		return gameApi.OutcomeNoResulted, fmt.Errorf("unmarshal report history for FbOutcome fail %v", uErr)
	}
	return gameApi.Outcome(reportInfo.SettleResult), nil
}