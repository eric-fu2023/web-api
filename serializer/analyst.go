package serializer

import (
	"web-api/model"
	"web-api/util"

	fbService "blgit.rfdev.tech/taya/game-service/fb2/service"
)

type Analyst struct {
	AnalystId        int64        `json:"analyst_id"`
	AnalystName      string       `json:"analyst_name"`
	AnalystDesc      string       `json:"analyst_desc"`
	AnalystSource    Source       `json:"analyst_source"`
	AnalystImage     string       `json:"analyst_image"`
	WinningStreak    int          `json:"winning_streak"`
	Accuracy         int          `json:"accuracy"`
	NumFollowers     int          `json:"num_followers"`
	TotalPredictions int          `json:"total_predictions"`
	Predictions      []Prediction `json:"predictions"`
	RecentTotal      int          `json:"recent_total"`
	RecentWins       int          `json:"recent_wins"`
}

type Source struct {
	Name string `json:"source_name"`
	Icon string `json:"source_icon"`
}

type Achievement struct {
	TotalPredictions int   `json:"total_predictions"`
	Accuracy         int   `json:"accuracy"`
	WinningStreak    int   `json:"winning_streak"`
	RecentResult     []int `json:"recent_result"`
}

func BuildAnalystsList(analysts []model.Analyst) (resp []Analyst) {
	resp = []Analyst{}
	for _, a := range analysts {
		resp = append(resp, BuildAnalystDetail(a))
	}
	return
}

func BuildAnalystDetail(analyst model.Analyst) (resp Analyst) {

	predictions := make([]Prediction, len(analyst.Predictions))

	for i, pred := range analyst.Predictions {
		predictions[i] = BuildPrediction(pred, true, false)
	}

	resp = Analyst{
		AnalystId:        analyst.ID,
		AnalystName:      analyst.Name,
		AnalystSource:    Source{Name: analyst.PredictionSource.SourceName, Icon: analyst.PredictionSource.IconUrl},
		AnalystImage:     "https://cdn.tayalive.com/aha-img/user/default_user_image/102.jpg",
		AnalystDesc:      analyst.Desc,
		Predictions:      predictions,
		NumFollowers:     len(analyst.Followers),
		TotalPredictions: len(analyst.Predictions),
		WinningStreak:    20, // TODO
		Accuracy:         99, // TODO
		RecentTotal:      0,  // TODO
		RecentWins:       0,  // TODO
	}
	return
}

func BuildFollowingList(followings []model.UserAnalystFollowing) (resp []Analyst) {
	for _, a := range followings {
		resp = append(resp, BuildAnalystDetail(a.Analyst))
	}
	return
}

func BuildAnalystAchievement(results []fbService.SelectionOutCome) (resp Achievement) {
	numResults := len(results)
	var last10results []fbService.SelectionOutCome
	if numResults > 10 {
		last10results = results[numResults-10:]
	} else {
		last10results = results
	}

	resultInBool := []bool{}
	winCount := 0
	for _, result := range results {
		if result == fbService.SelectionOutcomeRed {
			resultInBool = append(resultInBool, true)
			winCount++
		} else if result == fbService.SelectionOutcomeBlack {
			resultInBool = append(resultInBool, false)
		} else {
			continue // dont consider unsetteled/unknown statuses
		}
	}

	streak := util.ConsecutiveWins(resultInBool)

	accuracy := 0
	if len(resultInBool) != 0 {
		accuracy = winCount / len(resultInBool)
	}

	recentResult := make([]int, len(last10results))
	for i, res := range last10results {
		recentResult[i] = int(res)
	}

	resp = Achievement{
		TotalPredictions: len(results),
		Accuracy:         accuracy,
		WinningStreak:    streak,
		RecentResult:     recentResult,
	}
	return
}

func BuildFollowingAnalystIdsList(followings []model.UserAnalystFollowing) (resp []int64) {
	for _, a := range followings {
		resp = append(resp, a.AnalystId)
	}
	return
}
