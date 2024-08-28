package model

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"
	"web-api/cache"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/gin-gonic/gin"

	fbService "blgit.rfdev.tech/taya/game-service/fb2/outcome_service"

	"gorm.io/gorm"
)

const (
	RedisKeyAnalystDetailTemplate = "prediction:analyst:%d"
)
type Analyst struct {
	ploutos.PredictionAnalyst

	Predictions  []Prediction `gorm:"foreignKey:AnalystId;references:ID"`
	Summaries []ploutos.PredictionAnalystSummary `gorm:"foreignKey:AnalystId;references:ID"`
}

func preloadAnalyst() *gorm.DB {
	return DB.
		Preload("PredictionAnalystSource").
		Preload("PredictionAnalystFollowers").
		Preload("Predictions", AnalystPredictionFilter).
		Preload("Summaries").
		Where("is_active", true)
}

func (Analyst) List(page, limit int, fbSportId int64) (list []Analyst, err error) {
	db := preloadAnalyst()
	db = db.
		Scopes(Paginate(page, limit)).
		Select("prediction_analysts.*, MAX(prediction_articles.published_at) as latest_publish").
		Joins("join prediction_articles on prediction_articles.analyst_id = prediction_analysts.id").
		Where("prediction_articles.is_published", true).
		Where("prediction_articles.deleted_at IS NULL").
		Group("prediction_analysts.id").
		Order("sort DESC").
		Order("latest_publish desc")


	if fbSportId != 0 {
		db = db.
			Where("prediction_articles.fb_sport_id", fbSportId)
	}

	err = db.
		Find(&list).
		Error

	return
}

func (Analyst) GetDetail(c *gin.Context, id int) (target Analyst, err error) {
	redisKey := fmt.Sprintf(RedisKeyAnalystDetailTemplate, id)

	// get data from redis 
	if util.FindFromRedis(c, cache.RedisClient, redisKey, &target); target.ID != 0 {
		return
	}

	// get data from db 
	db := preloadAnalyst()
	db = db.Where("id", id)
	if err = db.
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Order("id DESC").
		First(&target).
		Error; err != nil {
			return
		}
	// store data in redis
	util.CacheIntoRedis(c, cache.RedisClient, redisKey, 1 * time.Minute, &target)
	return
}

func GetFollowingAnalystList(c context.Context, userId int64, page, limit int) (followings []UserAnalystFollowing, err error) {
	err = DB.
		Scopes(Paginate(page, limit)).
		Preload("Analyst").
		Preload("Analyst.PredictionAnalystSource").
		Preload("Analyst.PredictionAnalystFollowers").
		Preload("Analyst.Summaries").
		Joins("JOIN prediction_analysts on prediction_analyst_followers.analyst_id = prediction_analysts.id").
		WithContext(c).
		Where("user_id = ?", userId).
		Where("prediction_analysts.is_active = ?", true).
		Order("prediction_analysts.sort desc").
		Find(&followings).Error
	return
}

func GetFollowingAnalystStatus(c context.Context, userId, analystId int64) (following UserAnalystFollowing, err error) {
	err = DB.Unscoped().WithContext(c).Where("user_id = ?", userId).Where("analyst_id = ?", analystId).Limit(1).Find(&following).Error
	return
}

func UpdateUserFollowAnalystStatus(following UserAnalystFollowing) (err error) {
	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Save(&following).Error
		return
	})

	return
}

func SoftDeleteUserFollowAnalyst(following UserAnalystFollowing) (err error) {
	err = DB.Transaction(func(tx *gorm.DB) error {
		return tx.Delete(&following).Error
	})
	return
}

func RestoreUserFollowAnalyst(following UserAnalystFollowing) (err error) {
	err = DB.Transaction(func(tx *gorm.DB) error {
		return tx.Unscoped().Model(&following).Update("DeletedAt", nil).Error
	})
	return
}

func AnalystExist(analystId int64) (exist bool, err error) {
	err = DB.Where("id", analystId).First(&Analyst{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func AnalystPredictionFilter(db *gorm.DB) *gorm.DB {
	db = db.Where("is_published", true).
		Where("deleted_at is null").
		Order("published_at desc")
	return db
}

func GetPredictionsFromAnalyst(analyst Analyst, sportId int) []Prediction {
	sortedPredictions := slices.Clone(analyst.Predictions)
	filteredSorted := []Prediction{}
	for _, p := range sortedPredictions {
		if int64(sportId) == 0 || p.FbSportId == sportId {
			filteredSorted = append(filteredSorted, p)
		}
	}

	slices.SortFunc(filteredSorted, func(a, b Prediction) int {
		return b.PublishedAt.Compare(a.PublishedAt) // newest to oldest
	})
	return filteredSorted
}

func GetOutcomesFromPredictions(predictions []Prediction) []fbService.PredictionOutcome {
	outcomes := []fbService.PredictionOutcome{}
	for _, pred := range predictions {
		predDao := GetPredictionFromPrediction(pred)
		res, err := fbService.ComputePredictionOutcomesByOrderReport(predDao)

		if err != nil {
			log.Printf("Error computing prediction, %s\n", err.Error())
		}

		outcomes = append(outcomes, res)
	}
	return outcomes
}
