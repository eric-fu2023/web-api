package serializer

import (
	"web-api/model"
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
	for _, a := range analysts {
		resp = append(resp, BuildAnalystDetail(a))
	}
	return
}

func BuildAnalystDetail(analyst model.Analyst) (resp Analyst) {

	predictions := make([]Prediction, len(analyst.Predictions))

	for i, pred := range analyst.Predictions {
		predictions[i] = BuildPrediction(pred, true)
	}

	resp = Analyst{
		AnalystId:        analyst.ID,
		AnalystName:      analyst.Name,
		AnalystSource:    Source{Name: analyst.PredictionSource.SourceName, Icon: analyst.PredictionSource.IconUrl},
		AnalystImage:     "https://cdn.tayalive.com/aha-img/user/default_user_image/102.jpg",
		WinningStreak:    20,
		Accuracy:         99,
		AnalystDesc:      analyst.Desc,
		Predictions:      predictions,
		NumFollowers:     len(analyst.Followers),
		TotalPredictions: len(analyst.Predictions),
	}
	return
}

func BuildFollowingList(followings []model.UserAnalystFollowing) (resp []Analyst) {
	for _, a := range followings {
		resp = append(resp, BuildAnalystDetail(a.Analyst))
	}
	return
}

func BuildAnalystAchievement() (resp Achievement) {
	resp = Achievement{
		TotalPredictions: 323,
		Accuracy:         78,
		WinningStreak:    23,
		RecentResult:     []int{1, 1, 1, 1, 1, 0, 1, 1, 1, 1},
	}
	// TODO : ^^^ add logic
	return
}

func BuildFollowingAnalystIdsList(followings []model.UserAnalystFollowing) (resp []int64) {
	for _, a := range followings {
		resp = append(resp, a.AnalystId)
	}
	return
}
