package model

import (
	"errors"
	"fmt"
	"slices"

	fbService "blgit.rfdev.tech/taya/game-service/fb2/outcome_service"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

type Prediction struct {
	ploutos.PredictionArticle

	PredictionSelections []PredictionSelection `gorm:"foreignKey:ArticleId;references:ID"`
	AnalystDetail        Analyst               `gorm:"foreignKey:AnalystId;references:ID"`
}

type ListPredictionCond struct {
	Page      int
	Limit     int
	AnalystId int64
	FbMatchId int64
	SportId   int64
}

func preloadPredictions() *gorm.DB {
	return DB.
		Preload("PredictionSelections").
		Preload("PredictionSelections.FbOdds").
		Preload("PredictionSelections.FbOdds.RelatedOdds").
		Preload("PredictionSelections.FbOdds.MarketGroupInfo").
		Preload("PredictionSelections.FbMatch").
		Preload("PredictionSelections.FbMatch.LeagueInfo").
		Preload("AnalystDetail").
		Preload("AnalystDetail.PredictionAnalystSource").
		Preload("PredictionSelections.FbMatch.HomeTeam").
		Preload("PredictionSelections.FbMatch.AwayTeam")
	// Preload("AnalystDetail.Followers").
	// Preload("AnalystDetail.Predictions").
}

func ListPredictions(cond ListPredictionCond) (preds []Prediction, err error) {

	db := preloadPredictions()
	db = db.
		Scopes(Paginate(cond.Page, cond.Limit)).
		Where("prediction_articles.deleted_at IS NULL").
		Where("prediction_articles.is_published", true).
		Joins("left join prediction_article_bets on prediction_article_bets.article_id = prediction_articles.id").
		Joins("left join fb_matches on prediction_article_bets.fb_match_id = fb_matches.match_id").
		Group("prediction_articles.id")

	if cond.AnalystId != 0 {
		db = db.Where("prediction_articles.analyst_id", cond.AnalystId)
	}

	if cond.FbMatchId != 0 {
		db = db.Where("prediction_article_bets.match_id = ?", cond.FbMatchId)
	}

	if cond.SportId != 0 {
		db = db.Where("fb_matches.sports_id = ?", cond.SportId)
	}

	err = db.Find(&preds).Error

	return
}

func GetPrediction(predictionId int64) (pred Prediction, err error) {
	err = preloadPredictions().
		Where("deleted_at IS NULL").
		Where("is_published", true).
		Where("id = ?", predictionId).
		First(&pred).Error
	return
}

func PredictionExist(predictionId int64) (exist bool, err error) {
	err = DB.Where("id", predictionId).First(&Prediction{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func GetOrderByOddFromSelection(selection PredictionSelection, oddId int64) (reports []fbService.SelectionOrder) {
	if selection.FbOdds.ID == oddId {
		for _, order := range selection.FbOdds.FbOddsOrderRequestList {
			reports = append(reports, order.FbBetReport)
		}
	}
	return
}

func GenerateMarketGroupKeyFromSelection(selection PredictionSelection) (string) {
	marketGroupKey := fmt.Sprintf("%d-%d-%d-%d-%s", selection.FbOdds.SportsID, selection.FbOdds.MatchID, selection.FbOdds.MarketGroupType, selection.FbOdds.MarketGroupPeriod, selection.FbOdds.MarketlineValue)
	return marketGroupKey
}

func GetMarketGroupOrdersByKeyFromPrediction(prediction Prediction, key string) (mg fbService.MarketGroup) {
	for _, selection := range prediction.PredictionSelections {
		mgKey := GenerateMarketGroupKeyFromSelection(selection)
		if mgKey != key {
			continue
		}
		mg.GroupType = selection.FbOdds.MarketGroupType
		orders := []fbService.SelectionOrder{}
		for _, order := range selection.FbOdds.FbOddsOrderRequestList {
			orders = append(orders, order.FbBetReport)
		}
		mg.Selections = append(mg.Selections, fbService.Selection{Orders: orders})
	}
	return
}

func GetPredictionFromPrediction(prediction Prediction) (outPred fbService.Prediction) {
	existingMgKeys := []string{}
	for _, selection := range prediction.PredictionSelections {
		mgKey := GenerateMarketGroupKeyFromSelection(selection)
		if slices.Contains(existingMgKeys, mgKey) {
			continue
		}
		outPred.MarketGroups = append(outPred.MarketGroups, GetMarketGroupOrdersByKeyFromPrediction(prediction, mgKey))
	}
	return 
}
/*
one prediction has many selection
one selection has one odd
one odd has many orderRequest - one orderRequest has one fbReport ∴ one odd has one fbReport

one prediction has many match 
one match has many market group 
one market group has many odd
*/