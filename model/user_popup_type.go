package model

import (
	"errors"
	"fmt"
	"time"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm/logger"
)

func GetPopupTypeForMe(user User) ([]int8, error) {

	POPUPTYPEMAPPING := map[string]int8{
		"GGR":  1,
		"vip":  2,
		"spin": 3,
		"none": -1,
	}
	var response = []int8{-1}

	// winlose popup check
	// check if displayed before
	var previous_win_lose ploutos.PopupRecord
	err := DB.Model(ploutos.PopupRecord{}).Where("user_id = ? AND type = 1", user.ID).
		Order("created_at DESC").
		First(&previous_win_lose).Error
	if err != nil && !errors.Is(err, logger.ErrRecordNotFound) {
		return []int8{}, err
	}

	now := time.Now()
	TodayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	yesterdayEnd := yesterdayStart.Add(24 * time.Hour)

	// if not displayed today
	if previous_win_lose.CreatedAt.Before(TodayStart) {
		// check if user has GGR yesterday
		var count int64
		settleStatus := []int64{5}
		err = DB.Model(ploutos.BetReport{}).Scopes(ByOrderListConditionsGGR(user.ID, settleStatus, yesterdayStart, yesterdayEnd)).Count(&count).Error
		if err != nil {
			return []int8{}, err
		}
		if count > 0 {
			var ggr int64
			err = DB.Model(ploutos.BetReport{}).Scopes(ByOrderListConditionsGGR(user.ID, settleStatus, yesterdayStart, yesterdayEnd)).Select("SUM(win-bet) as win").Find(&ggr).Error
			if ggr != 0 {
				response[0] = POPUPTYPEMAPPING["GGR"]
				return response, nil
			}
		}
	}
	// check if user has VIP lvl up yesterday
	var previous_vip ploutos.PopupRecord
	err = DB.Model(ploutos.PopupRecord{}).Where("user_id = ?  AND type = 2", user.ID).
		Order("created_at DESC").
		First(&previous_vip).Error
	if err != nil && !errors.Is(err, logger.ErrRecordNotFound) {
		return []int8{}, err
	}
	vip, err := GetVipWithDefault(nil, user.ID)
	if err != nil {
		return []int8{}, err
	}
	currentVipRule := vip.VipRule
	// if there is a level up
	if previous_vip.VipLevel < currentVipRule.VIPLevel {
		response[0] = POPUPTYPEMAPPING["vip"]
		return response, nil
	} else {
		// if there is a vip downgrade, we need to update the deleted_at fro the record
		err = DB.Model(ploutos.PopupRecord{}).
			Where("user_id = ? AND vip_level > ?  AND type = 2", user.ID, currentVipRule.VIPLevel).
			Update("deleted_at", now).Error
	}
	// display spin
	var previous_spin_result ploutos.SpinResult
	err = DB.Model(ploutos.SpinResult{}).Where("user_id = ?", user.ID).
		Order("created_at DESC").
		First(&previous_spin_result).Error
	fmt.Print(previous_spin_result.CreatedAt)
	fmt.Print(TodayStart)
	if previous_spin_result.CreatedAt.Before(TodayStart) {
		response[0] = POPUPTYPEMAPPING["spin"]
		return response, nil
	}
	// if dun need to display
	return response, nil
}
