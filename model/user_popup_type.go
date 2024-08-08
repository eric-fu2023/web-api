package model

import (
	"context"
	"errors"
	"fmt"
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

func ShouldPopupTeamUp(user User) (bool, error) {
	now := time.Now()
	TodayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var team_up ploutos.Teamup
	// status = 2 is success,    status = 0 is onging
	err := DB.Model(ploutos.Teamup{}).Where("user_id = ? AND created_at < ? AND status in (2,0)", user.ID, TodayStart).Order("status DESC, total_teamup_deposit DESC").First(&team_up).Error
		if errors.Is(err, logger.ErrRecordNotFound) {
			err = nil
			// if no team up record, we return nil
			return false, err
		}
		if err != nil {
			fmt.Println("ShouldPopupTeamUp teamup err", err.Error())
			return false, err
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

func ShouldPopupSpin(user User) (bool, error) {
	now := time.Now()
	TodayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// if not displayed today
	var spin_result ploutos.SpinResult
	err := DB.Model(ploutos.SpinResult{}).Where("user_id = ?", user.ID).
		Order("created_at DESC").
		First(&spin_result).Error
		if errors.Is(err, logger.ErrRecordNotFound) {
			// if spin result not found
			err = nil
			return true, nil
		}
	if spin_result.CreatedAt.Before(TodayStart) {
		return true, nil
	}
	return false, nil

}


func GetPopupList(condition int64) (resp_list []ploutos.Popups, err error) {
	err = DB.Model(ploutos.Popups{}).Where("condition = ?", condition).
		Find(&resp_list).Error
	return
}