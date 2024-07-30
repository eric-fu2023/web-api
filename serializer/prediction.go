package serializer

type Prediction struct {
	PredictionId     int64   `json:"prediction_id"`
	AnalystId        int64   `json:"analyst_id"`
	PredictedMatches []Match `json:"predicted_matches"`
	PredictionTitle  string  `json:"prediction_title"`
	PredictionDesc   string  `json:"prediction_desc"`
	IsLocked         bool    `json:"is_locked"`
}

type PredictedMatch struct {
	MatchId int64 `json:"match_id"`
	FbBetId int64 `json:"fb_bet_id"`
	Status  int64 `json:"status"`
}

// func BuildPredictionList(predictions []models.Prediction) (res []Prediction, err error) {

// 	for _, prediction := range predictions {

// 		p := Prediction{
// 			AnalystId:       prediction.AnalystId,
// 			PredictionTitle: prediction.PredictionTitle,
// 			PredictionDesc:  prediction.PredictionDesc,
// 		}

// 		res = append(res, p)
// 	}

// 	return
// }
