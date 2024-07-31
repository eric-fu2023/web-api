package serializer

import (
	"time"
	"web-api/model"
)

type Prediction struct {
	PredictionId      int64     `json:"prediction_id"`
	AnalystId         int64     `json:"analyst_id"`
	PredictedMatches  []Match   `json:"predicted_matches"`
	PredictionTitle   string    `json:"prediction_title"`
	PredictionDesc    string    `json:"prediction_desc"`
	IsLocked          bool      `json:"is_locked"`
	CreatedAt         time.Time `json:"created_at"`
	ViewCount         int64     `json:"view_count"`
}

type PredictedMatch struct {
	MatchId int64 `json:"match_id"`
	FbBetId int64 `json:"fb_bet_id"`
	Status  int64 `json:"status"`
}

func BuildPredictionsList(predictions []model.Prediction) (preds []Prediction) {
	for _, p := range predictions {
		preds = append(preds, Prediction{
			PredictionId: p.ID,
			AnalystId: p.AnalystId,
			PredictionTitle: p.Title,
			PredictionDesc: p.Description,
			CreatedAt: p.CreatedAt,
			ViewCount: p.Views,
			IsLocked: false,
		})
	}
	return 
}

