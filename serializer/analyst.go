package serializer

import (
	"web-api/model"
	"web-api/util"

	fbService "blgit.rfdev.tech/taya/game-service/fb2/outcome_service"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"github.com/eapache/go-resiliency/breaker"
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

func BuildAnalystsList(analysts []model.Analyst) (resp []Analyst) {
	resp = []Analyst{}
	for _, a := range analysts {
		// only display analysts with published PredictionArticles
		if len(a.Predictions) > 0 {
			resp = append(resp, BuildAnalystDetail(a))
		}
	}
	return
}

func BuildAnalystDetail(analyst model.Analyst) (resp Analyst) {
	predictions := make([]Prediction, len(analyst.Predictions))

	for i, pred := range analyst.Predictions {
		predictions[i] = BuildPrediction(pred, true, false)
	}

	statuses := model.GetOutcomesFromPredictions(model.GetPredictionsFromAnalyst(analyst, 0))

	statusInBool, _ := GetBoolOutcomes(statuses) // this function removes unknown statuses
	nearX, winX := util.NearXWinX(statusInBool)

	winStreak := util.RecentConsecutiveWins(statusInBool)

	// accuracty based on latest 10
	accuracy := 0
	if len(statusInBool) > 0 {
		accuracy = util.Accuracy(statusInBool)
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
		WinningStreak:    winStreak,
		Accuracy:         accuracy,
		RecentTotal:      nearX,
		RecentWins:       winX,

		IsShowStreak:     winStreak >= 3,
		IsShowAccuracy:   accuracy > 10.0,
		IsShowRecentWins: (float64(winX)/float64(nearX) * 100) > 50.0,
	}
	return
}

func BuildFollowingList(followings []model.UserAnalystFollowing) (resp []Analyst) {
	resp = []Analyst{}
	for _, a := range followings {
		resp = append(resp, BuildAnalystDetail(a.Analyst))
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
