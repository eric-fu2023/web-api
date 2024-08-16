package serializer

import (
	"log"
	"web-api/model"
	"web-api/util"

	fbService "blgit.rfdev.tech/taya/game-service/fb2/outcome_service"
)

type Analyst struct {
	AnalystId        int64        `json:"analyst_id"`
	AnalystName      string       `json:"analyst_name"`
	AnalystDesc      string       `json:"analyst_desc"`
	AnalystSource    Source       `json:"analyst_source"`
	AnalystImage     string       `json:"analyst_image"`
	WinningStreak    int          `json:"winning_streak"`
	Accuracy         float64      `json:"accuracy"`
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
	TotalPredictions int     `json:"total_predictions"`
	Accuracy         float64 `json:"accuracy"`
	WinningStreak    int     `json:"winning_streak"`
	RecentResult     []int   `json:"recent_result"`
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
	statuses := make([]fbService.PredictionOutcome, len(analyst.Predictions))

	for i, pred := range analyst.Predictions {
		predictions[i] = BuildPrediction(pred, true, false)
		_pred := model.GetPredictionFromPrediction(pred)
		outcome, err := fbService.ComputePredictionOutcomesByOrderReport(_pred) 
		if err != nil {
			log.Printf("error computing outcome of prediction[ID:%d]", pred.ID)
		}
		statuses[i] = outcome 
	}

	statusInBool, winCount := GetBoolOutcomes(statuses)
	nearX, winX := util.NearXWinX(statusInBool)

	winStreak := util.ConsecutiveWins(statusInBool)
	accuracy := 0.0
	if len(statusInBool) > 0 {
		accuracy = float64(winCount) / float64(len(statusInBool))
	}

	resp = Analyst{
		AnalystId:        analyst.ID,
		AnalystName:      analyst.AnalystName,
		AnalystSource:    Source{Name: analyst.PredictionAnalystSource.SourceName, Icon: analyst.PredictionAnalystSource.IconUrl},
		AnalystImage:     analyst.AvatarUrl,
		AnalystDesc:      analyst.AnalystDesc,
		Predictions:      predictions,
		NumFollowers:     len(analyst.PredictionAnalystFollowers),
		TotalPredictions: len(analyst.Predictions),
		WinningStreak:    winStreak,
		Accuracy:         accuracy,
		RecentTotal:      nearX,
		RecentWins:       winX,
	}
	return
}

func BuildFollowingList(followings []model.UserAnalystFollowing) (resp []Analyst) {
	// FIXME : preload will return default if analyst not exist.. for now, filter away id = 0
	resp = []Analyst{}
	for _, a := range followings {
		if (a.ID == 0) {continue}
		resp = append(resp, BuildAnalystDetail(a.Analyst))
	}
	return
}

func BuildAnalystAchievement(results []fbService.PredictionOutcome) (resp Achievement) {
	// total predictions
	numResults := len(results)

	// win/lose for the last 10 predictions
	var last10results []fbService.PredictionOutcome
	if numResults > 10 {
		last10results = results[numResults-10:]
	} else {
		last10results = results
	}
	recentResult := make([]int, len(last10results))
	for i, res := range last10results {
		recentResult[i] = int(res)
	}

	// set up for winning streak and accuracy
	resultInBool, winCount := GetBoolOutcomes(results)

	// winning streak
	streak := util.ConsecutiveWins(resultInBool)

	// accuracy
	accuracy := 0.0
	if len(resultInBool) != 0 {
		accuracy = float64(winCount) / float64(len(resultInBool))
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
