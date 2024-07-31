package model

import (
	"errors"
	"time"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm/logger"
)

func ShouldPopupWinLose(user User) (bool, error) {
	now := time.Now()
	TodayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	yesterdayEnd := yesterdayStart.Add(24 * time.Hour)


	// if not displayed today
	var previous_win_lose ploutos.PopupRecord
	err := DB.Model(ploutos.PopupRecord{}).Where("user_id = ? AND type = 1", user.ID).
		Order("created_at DESC").
		First(&previous_win_lose).Error
	if err != nil && !errors.Is(err, logger.ErrRecordNotFound) {
		return false, err
	}
	if previous_win_lose.CreatedAt.Before(TodayStart) {

		// check if user has GGR yesterday
		var count int64
		settleStatus := []int64{5}
		err = DB.Model(ploutos.BetReport{}).Scopes(ByOrderListConditionsGGR(user.ID, settleStatus, yesterdayStart, yesterdayEnd)).Count(&count).Error
		if err != nil {
			return false, err
		}
		if count > 0 {
			// get user GGR
			var ggr int64
			err = DB.Model(ploutos.BetReport{}).Scopes(ByOrderListConditionsGGR(user.ID, settleStatus, yesterdayStart, yesterdayEnd)).Select("SUM(win-bet) as win").Find(&ggr).Error
			if ggr != 0 {
				// if user has win or lose
				return true, nil
			}
		}
	}
	return false, err
}

func ShouldPopupVIP(user User) (bool, error) {
	now := time.Now()
	TodayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// if not displayed today
	var previous_vip ploutos.PopupRecord
	err := DB.Model(ploutos.PopupRecord{}).Where("user_id = ?  AND type = 2", user.ID).
		Order("created_at DESC").
		First(&previous_vip).Error
	if err != nil && !errors.Is(err, logger.ErrRecordNotFound) {
		return false, err
	}
	if previous_vip.CreatedAt.Before(TodayStart) {
		// check if user has VIP lvl up yesterday
		vip, err := GetVipWithDefault(nil, user.ID)
		if err != nil {
			return false, err
		}
		currentVipRule := vip.VipRule
		// if there is a level up
		if previous_vip.VipLevel < currentVipRule.VIPLevel {
			return true, nil
		} else {
			// if there is a vip downgrade, we need to update the deleted_at for the record
			err = DB.Model(ploutos.PopupRecord{}).
				Where("user_id = ? AND vip_level > ?  AND type = 2", user.ID, currentVipRule.VIPLevel).
				Update("deleted_at", now).Error
		}
	}
	return false,err
}