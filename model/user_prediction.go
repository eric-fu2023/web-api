package model

import (
	"errors"
	"web-api/util"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

type UserPrediction struct {
	ploutos.UserPrediction
}

type GetUserPredictionCond struct {
	DeviceId string
	UserId   int64
}

func GetUserPrediction(cond GetUserPredictionCond) ([]UserPrediction, error) {
	return GetUserPredictionWithDB(DB, cond)
}

func GetUserPredictionWithDB(tx *gorm.DB, cond GetUserPredictionCond) ([]UserPrediction, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}
	if cond.DeviceId == "" {
		return nil, errors.New("invalid uuid")
	}

	db := tx.Table(UserPrediction{}.TableName())

	db.Where("device_id = ?", cond.DeviceId)
	db.Where("user_id = ?", cond.UserId)

	now, err := util.NowGMT8()
	if err != nil {
		return nil, err
	}
	start := util.RoundDownTimeDay(now)
	end := util.RoundUpTimeDay(now)

	db.Where("created_at >= ?", start)
	db.Where("created_at < ?", end)

	var strategies []UserPrediction
	err = db.Find(&strategies).Error

	return strategies, err
}

func CreateUserPrediction(userId int64, deviceId string, predictionId int64) error {
	tx := DB.Begin()
	err := CreateUserPredictionWithDB(tx, userId, deviceId, predictionId)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func CreateUserPredictionWithDB(tx *gorm.DB, userId int64, deviceId string, predictionId int64) error {
	if tx == nil {
		return errors.New("tx is nil")
	}

	// TODO : check if strategy exist

	// obj := UserStrategy{UserStrategy: ploutos.UserStrategy{
	// 	UserId: userId,
	// 	DeviceId: deviceId,
	// 	StrategyId: strategy_id,
	// }}

	obj := ploutos.UserPrediction{
		UserId:       userId,
		DeviceId:     deviceId,
		PredictionId: predictionId,
	}

	return tx.Create(&obj).Error

}
