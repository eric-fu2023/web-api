package model

import (
	"errors"
	"fmt"
	"slices"
	"time"
	"web-api/cache"
	"web-api/util"

	fbService "blgit.rfdev.tech/taya/game-service/fb2/outcome_service"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	RedisKeyPredictionDetailTemplate = "prediction:article:%d"
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

type BrandId int
const (
	BrandIdAha BrandId = 3001
	BrandIdBatace BrandId = 1002
)

func preloadPredictions() *gorm.DB {
	return DB.
		Preload("PredictionSelections").
		Preload("PredictionSelections.FbOdds").
		Preload("PredictionSelections.FbOdds.RelatedOdds", SortFbOddsByShortName).
		Preload("PredictionSelections.FbOdds.FbOddsOrderRequestList").
		Preload("PredictionSelections.FbOdds.FbOddsOrderRequestList.TayaBetReport").
		Preload("PredictionSelections.FbOdds.MarketGroupInfo").
		Preload("PredictionSelections.FbMatch").
		Preload("PredictionSelections.FbMatch.LeagueInfo").
		Preload("AnalystDetail").
		Preload("AnalystDetail.PredictionAnalystSource").
		Preload("AnalystDetail.Predictions", AnalystPredictionFilter).
		Preload("AnalystDetail.Summaries").
		Preload("PredictionSelections.FbMatch.HomeTeam").
		Preload("PredictionSelections.FbMatch.AwayTeam").
		Joins("join prediction_analysts on prediction_analysts.id = prediction_articles.analyst_id").
		Where("prediction_analysts.is_active", true).Where("prediction_analysts.deleted_at IS null")
	// Preload("AnalystDetail.Followers").
	// Preload("AnalystDetail.Predictions").
}

func ListPredictions(cond ListPredictionCond) (preds []Prediction, err error) {

	db := preloadPredictions()
	db = db.
		Scopes(Paginate(cond.Page, cond.Limit)).
		Select(`
			prediction_articles.*, 
			case prediction_articles.prediction_result when 0 then 0 else 1 end as is_settle,
			(COALESCE(cast(pas.recent_win as float) / NULLIF(cast(pas.recent_total as float), 0), 0) * 100 * 0.5) + (pas.accuracy * 0.5) as weight
		`).
		Where("prediction_articles.is_published", true).
		Joins("left join prediction_article_bets on prediction_article_bets.article_id = prediction_articles.id").
		Joins("left join fb_matches on prediction_article_bets.fb_match_id = fb_matches.match_id").
		Joins("left join prediction_analyst_summary pas on prediction_articles.analyst_id = pas.analyst_id and pas.fb_sport_id = 0").
		Group("prediction_articles.id, pas.id").
		Order("is_settle asc").
		Order("weight desc").
		Order("prediction_articles.published_at DESC")

	if cond.AnalystId == 0 { 
		_y, _m, _d := time.Now().AddDate(0, 0, -7).Date()
		weekAgo := time.Date(_y, _m, _d, 0, 0, 0, 0, time.Now().Location())

		db = db.Where("prediction_articles.published_at > ?", weekAgo)
	}

	if cond.AnalystId != 0 {
		db = db.Where("prediction_articles.analyst_id", cond.AnalystId)
	}

	if cond.FbMatchId != 0 {
		db = db.Where("prediction_articles.fb_match_id = ?", cond.FbMatchId)
	}

	if cond.SportId != 0 {
		db = db.Where("prediction_articles.fb_sport_id = ?", cond.SportId)
	}

	err = db.Find(&preds).Error

	return
}

func GetPrediction(c *gin.Context, predictionId int64) (pred Prediction, err error) {
	redisKey := fmt.Sprintf(RedisKeyPredictionDetailTemplate, predictionId)
   
	// get data from redis
	if util.FindFromRedis(c, cache.RedisClient, redisKey, &pred); pred.ID != 0 {
		return
	}

	// get data form db 
	if err = preloadPredictions().
		Where("is_published", true).
		Where("prediction_articles.id = ?", predictionId).
		First(&pred).
		Error; err != nil {
			return 
		}
	// store data in redis 
	util.CacheIntoRedis(c, cache.RedisClient, redisKey, 1 * time.Minute, &pred)
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
			reports = append(reports, order.TayaBetReport)
		}
	}
	return
}

func GenerateMarketGroupKeyFromSelection(selection PredictionSelection) string {
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
			orders = append(orders, order.TayaBetReport)
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

func IncreasePredictionViewCountBy1(prediction Prediction) error {
	err := DB.Transaction(func(tx *gorm.DB) error {
		return tx.Model(&prediction.PredictionArticle).Update("Views", prediction.Views+1).Error
	})
	return err
}

func GetPredictionSportId(p Prediction) int64 {
	if len(p.PredictionSelections) == 0 {
		return 0
	} else {
		return int64(p.PredictionSelections[0].FbMatch.SportsID)
	}
}