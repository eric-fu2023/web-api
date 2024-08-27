package serializer

import (
	"web-api/model"

	fbService "blgit.rfdev.tech/taya/game-service/fb2/outcome_service"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
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

	IsShowStreak     bool `json:"is_show_streak"`
	IsShowAccuracy   bool `json:"is_show_accuracy"`
	IsShowRecentWins bool `json:"is_show_recent_wins"`
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

	IsShowTotal      bool `json:"is_show_total"`
	IsShowAccuracy   bool `json:"is_show_accuracy"`
	IsShowStreak     bool `json:"is_show_streak"`
	IsShowRecentResult bool `json:"is_show_recent_result"`
}

func BuildAnalystsList(analysts []model.Analyst, brandId int) (resp []Analyst) {
	resp = []Analyst{}
	for _, a := range analysts {
		// only display analysts with published PredictionArticles
		if len(a.Predictions) > 0 {
			resp = append(resp, BuildAnalystDetail(a, brandId))
		}
	}
	return
}

func BuildAnalystDetail(analyst model.Analyst, brandId int) (resp Analyst) {
	predictions := make([]Prediction, len(analyst.Predictions))

	for i, pred := range analyst.Predictions {
		predictions[i] = BuildPrediction(pred, true, false, brandId)
	}

	summary := ploutos.PredictionAnalystSummary{}
	for _, s := range analyst.Summaries{
		if s.FbSportId == 0 { // overall results
			summary = s
			break
		}
	}

	resp = Analyst{
		AnalystId:        analyst.ID,
		AnalystName:      analyst.AnalystName,
		AnalystSource:    Source{Name: analyst.PredictionAnalystSource.SourceName, Icon: Url(analyst.PredictionAnalystSource.IconUrl)},
		AnalystImage:     Url(analyst.AvatarUrl),
		AnalystDesc:      analyst.AnalystDesc,
		Predictions:      predictions,
		NumFollowers:     len(analyst.PredictionAnalystFollowers),
		TotalPredictions: len(analyst.Predictions),
		WinningStreak:    summary.RecentStreak,
		Accuracy:         summary.Accuracy,
		RecentTotal:      summary.RecentTotal,
		RecentWins:       summary.RecentWin,

		IsShowStreak:     summary.RecentStreak >= 3,
		IsShowAccuracy:   summary.Accuracy > 10,
		IsShowRecentWins: (float64(summary.RecentWin)/float64(summary.RecentTotal) * 100) > 50.0 && summary.RecentTotal >= 3,
	}
	return
}

func BuildFollowingList(followings []model.UserAnalystFollowing, brandId int) (resp []Analyst) {
	resp = []Analyst{}
	for _, a := range followings {
		resp = append(resp, BuildAnalystDetail(a.Analyst, brandId))
	}
	return
}

func BuildAnalystAchievement(analyst model.Analyst, sportId int) (resp Achievement) {
	summary := ploutos.PredictionAnalystSummary{}

	for _, s := range analyst.Summaries{
		if s.FbSportId == sportId {
			summary = s 
			break
		}
	}

	recentResults := []int{}
	for _, res := range summary.RecentResults {
		if res == int32(ploutos.PredictionResultUnknown) {
			continue
		} else {
			recentResults = append(recentResults, int(res))
		}
	}
	if len(recentResults) > 10 { // truncate 
		recentResults = recentResults[:10]
	}

	resp = Achievement{
		TotalPredictions: summary.TotalArticles,
		Accuracy: summary.Accuracy,
		WinningStreak: summary.HighestStreak,
		RecentResult: recentResults,

		IsShowTotal: true,
		IsShowAccuracy: summary.Accuracy > 10,
		IsShowStreak: summary.HighestStreak >= 3,
		IsShowRecentResult: len(recentResults) > 0,
	}
	return
}

func BuildFollowingAnalystIdsList(followings []model.UserAnalystFollowing) (resp []int64) {
	for _, a := range followings {
		resp = append(resp, a.AnalystId)
	}
	return
}

func GetBoolOutcomes(results []fbService.PredictionOutcome) (resultInBool []bool, winCount int) {
	resultInBool = []bool{}
	winCount = 0
	for _, result := range results {
		if result == fbService.PredictionOutcomeOutcomeRed {
			resultInBool = append(resultInBool, true)
			winCount++
		} else if result == fbService.PredictionOutcomeOutcomeBlack {
			resultInBool = append(resultInBool, false)
		} else {
			continue // dont consider unsetteled/unknown statuses
		}
	}
	return
}
