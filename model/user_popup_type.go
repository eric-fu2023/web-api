package model

import (
	"time"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)



func GetPopupTypeForMe(userId int64) ([]int8, error) {
	
    POPUPTYPEMAPPING := map[string]int8{
        "GGR":  1,
        "vip":  2,
        "spin": 3,
        "none": -1,
    }
	var response = []int8{-1}

    now := time.Now()
    // Get the start of yesterday (00:00)
    yesterdayStart := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
    // Get the end of yesterday (24:00)
    yesterdayEnd := yesterdayStart.Add(24 * time.Hour)

	// check if user has GGR yesterday
	// check if displayed before
	
	var sum struct{
		win int64
	}
	settleStatus := []int64{5}
	err := DB.Model(ploutos.BetReport{}).Scopes(ByOrderListConditionsGGR(userId, settleStatus, yesterdayStart, yesterdayEnd)).
		Select(`SUM(win-bet) as win`).Find(&sum).Error
	if err != nil {
		return []int8{}, err
	}
	if sum.win!=0{
		response[0] = POPUPTYPEMAPPING["GGR"]
		return response, nil
	}
	// check if user has VIP lvl up yesterday

	// display spin

	// if dun need to display
	return response, nil
}