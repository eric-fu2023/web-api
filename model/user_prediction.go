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
	// if cond.UserId == 0 {
	// 	db.Where("user_id = 0")
	// }
	// db.Where("user_id = 0 OR user_id = ?", cond.UserId)

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
	exist, err := todayHasId(predictionId, deviceId)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}
	return CreateUserPredictionWithDB(DB, userId, deviceId, predictionId)
}

func CreateUserPredictionWithDB(tx *gorm.DB, userId int64, deviceId string, predictionId int64) error {
	if tx == nil {
		return errors.New("tx is nil")
	}

	exist, err := PredictionExist(predictionId)

	if err != nil {
		return err
	}

	if !exist {
		return errors.New("prediction does not exist")
	}

	obj := ploutos.UserPrediction{
		UserId:       userId,
		DeviceId:     deviceId,
		PredictionId: predictionId,
	}

	return tx.Create(&obj).Error

}

func GetUserPredictionCount(deviceId string) (count int64, err error) {
	db := DB.Table(UserPrediction{}.TableName())

	db.Where("device_id = ?", deviceId)

	now, err := util.NowGMT8()
	if err != nil {
		return
	}
	start := util.RoundDownTimeDay(now)
	end := util.RoundUpTimeDay(now)

	db.Where("created_at >= ?", start)
	db.Where("created_at < ?", end)
	db.Where("deleted_at IS NULL")

	err = db.Count(&count).Error

	return
}

func todayHasId(predictionId int64, deviceId string) (exist bool, err error) {
	db := DB.Table(UserPrediction{}.TableName())

	now, err := util.NowGMT8()
	if err != nil {
		return
	}
	start := util.RoundDownTimeDay(now)
	end := util.RoundUpTimeDay(now)

	db.Where("created_at >= ?", start)
	db.Where("created_at < ?", end)
	db.Where("deleted_at IS NULL")

	db.Where("prediction_id = ?", predictionId)
	db.Where("device_id = ?", deviceId)

	err = db.First(&ploutos.UserPrediction{}).Error

	if err != nil {
		exist = false
		return exist, nil
	}
	exist = true

	return

}
