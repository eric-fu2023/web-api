package test

import (
	"fmt"
	"math"
	"testing"
	"time"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

const (
	TEAMUP_TARGET_1 = 1200
	TEAMUP_TARGET_2 = 5000
	TEAMUP_TARGET_3 = 10000
	TEAMUP_TARGET_4 = 100000

	TEAMUP_TARGET_11 = 1771
)

type TeamupEntryCustomRes []struct {
	ContributedAmount    float64   `json:"contributed_amount"`
	ContributedTime      time.Time `json:"contributed_time"`
	Nickname             string    `json:"nickname"`
	Avatar               string    `json:"avatar"`
	Progress             int64     `json:"progress"`
	AdjustedFiatProgress float64   `json:"adjusted_fiat_progress"`
}

// var customRes1 = TeamupEntryCustomRes{
// 	{Progress: 81},
// 	{Progress: 381},
// 	{Progress: 79},
// 	{Progress: 3164},
// 	{Progress: 21},
// 	{Progress: 6621},
// }

var customRes1 = TeamupEntryCustomRes{
	{Progress: 0},
	{Progress: 0},
	{Progress: 0},
	{Progress: 0},
	{Progress: 79},
	{Progress: 9920},
}

func TestPercentageCalculation(t *testing.T) {

	teamupEntries := customRes1
	partialTotalProgress := 0.00
	teamup := ploutos.Teamup{
		TotalTeamUpTarget: TEAMUP_TARGET_4,
		// Status:            int(ploutos.TeamupStatusPending),
		// Status: int(ploutos.TeamupStatusFail),
		Status: int(ploutos.TeamupStatusSuccess),
	}

	teamupEntries = mapFormatAdjustedFiatProgress(teamupEntries, func(entries TeamupEntryCustomRes) TeamupEntryCustomRes {
		for i := len(teamupEntries) - 1; i >= 0; i-- {

			floorFiatProgress := math.Floor((float64(teamupEntries[i].Progress)/10000)*float64(teamup.TotalTeamUpTarget)/100*100) / 100
			teamupEntries[i].AdjustedFiatProgress = floorFiatProgress

			if int(floorFiatProgress) == 0 {
				teamupEntries[i].AdjustedFiatProgress = 1
			}

			teamupEntries[i].AdjustedFiatProgress = float64(int(teamupEntries[i].AdjustedFiatProgress))

			// fmt.Println(teamupEntries[i].AdjustedFiatProgress)

			if i != 0 {
				partialTotalProgress += teamupEntries[i].AdjustedFiatProgress
			} else {

				if teamup.Status == int(ploutos.TeamupStatusPending) || teamup.Status == int(ploutos.TeamupStatusFail) && partialTotalProgress >= float64(int(teamup.TotalTeamUpTarget/100)-1) {
					teamupEntries[i].AdjustedFiatProgress = 0
				}
			}

			// fmt.Println(fmt.Sprint(partialTotalProgress) + " - " + fmt.Sprint(teamupEntries[i].AdjustedFiatProgress))

			if int(partialTotalProgress) >= int(teamup.TotalTeamUpTarget/100) {
				prev := partialTotalProgress - teamupEntries[i].AdjustedFiatProgress
				teamupEntries[i].AdjustedFiatProgress = float64(teamup.TotalTeamUpTarget/100-1) - float64(prev)
				partialTotalProgress = float64(int(teamup.TotalTeamUpTarget/100) - 1)
			}

		}

		return teamupEntries

	})

	for _, e := range teamupEntries {
		fmt.Println(e.AdjustedFiatProgress)
	}
}

func mapFormatAdjustedFiatProgress[T any](t T, f func(T) T) T {
	return f(t)
}
