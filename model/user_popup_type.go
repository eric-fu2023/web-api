package model

import (
	"context"
	"errors"
	"strconv"
	"time"
	"web-api/cache"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm/logger"
)

func ShouldPopupWinLose(user User) (bool, error) {
	now := time.Now()
	key := "popup/win_lose/"+now.Format("2006-01-02")
	// here we need to use db2 to get the task system redis data
	res := cache.RedisClient.HGet(context.Background(), key, strconv.FormatInt(user.ID, 10))
	if res.Err() != nil {
		if res.Err() == redis.Nil{
			return false, nil
		}else{
			return false, res.Err()
		}
	}
	return true, nil
}

func ShouldPopupVIP(user User) (bool, error) {
	now := time.Now()
	TodayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// if not displayed today
	var previous_vip ploutos.PopupRecord
	err := DB.Model(ploutos.PopupRecord{}).Where("user_id = ?  AND type = 2", user.ID).
		Order("created_at DESC").
		First(&previous_vip).Error

	if errors.Is(err, logger.ErrRecordNotFound) {
		err = nil
		// if no vip level up record, we check if user vip level is more than 1
		vip, err := GetVipWithDefault(nil, user.ID)
		if errors.Is(err, logger.ErrRecordNotFound) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		currentVipRule := vip.VipRule
		if currentVipRule.VIPLevel > 1{
			return true, nil
		}
	}
	if err != nil {
		return false, err
	}
	if previous_vip.CreatedAt.Before(TodayStart) {
		// check if user has VIP lvl up yesterday
		vip, err := GetVipWithDefault(nil, user.ID)
		if errors.Is(err, logger.ErrRecordNotFound) {
			err = nil
		}
		if err != nil {
			return false, err
		}
		currentVipRule := vip.VipRule
		// if there is a level up
		if previous_vip.VipLevel < currentVipRule.VIPLevel {
			return true, nil
		} else if previous_vip.VipLevel > currentVipRule.VIPLevel{
			// if there is a vip downgrade, we need to update the deleted_at for the record
			err = DB.Model(ploutos.PopupRecord{}).
				Where("user_id = ? AND vip_level > ?  AND type = 2", user.ID, currentVipRule.VIPLevel).
				Update("deleted_at", now).Error
		}
	}
	return false,err
}

func GetPopupList(condition int64) (resp_list []ploutos.Popups, err error) {
	err = DB.Model(ploutos.Popups{}).Where("condition = ?", condition).
		Find(&resp_list).Error
	return
}