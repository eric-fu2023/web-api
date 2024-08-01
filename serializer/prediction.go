package serializer

import (
	"time"
	"web-api/model"
)

type Prediction struct {
	PredictionId    int64             `json:"prediction_id"`
	AnalystId       int64             `json:"analyst_id"`
	PredictionTitle string            `json:"prediction_title"`
	PredictionDesc  string            `json:"prediction_desc"`
	IsLocked        bool              `json:"is_locked"`
	CreatedAt       time.Time         `json:"created_at"`
	ViewCount       int64             `json:"view_count"`
	SelectionList   []SelectionDetail `json:"selection_list,omitempty"`
}

type PredictedMatch struct {
	MatchId int64 `json:"match_id"`
	FbBetId int64 `json:"fb_bet_id"`
	Status  int64 `json:"status"`
}

type SelectionDetail struct {
	MarketGroupType   int64 `json:"mty"`
	MarketGroupPeriod int64 `json:"pe"`
	OrderMarketlineId int64 `json:"id"`
	MatchType         int64 `json:"ty"`
	MatchId           int64 `json:"match_id"`
}

func BuildPredictionsList(predictions []model.Prediction) (preds []Prediction) {
	for _, p := range predictions {
		preds = append(preds, Prediction{
			PredictionId:    p.ID,
			AnalystId:       p.AnalystId,
			PredictionTitle: p.Title,
			PredictionDesc:  p.Description,
			CreatedAt:       p.CreatedAt,
			ViewCount:       p.Views,
			IsLocked:        false,
		})
	}
	return
}

func BuildPrediction(prediction model.Prediction) (pred Prediction) {
	selectionList := make([]SelectionDetail, len(prediction.PredictionSelections))
	for j, match := range prediction.PredictionSelections {
		selectionList[j] = SelectionDetail{
			MatchId:           match.MatchId,
			MarketGroupType:   0,
			MarketGroupPeriod: 0,
			OrderMarketlineId: 0,
			MatchType:         0,
		}
	}

	pred = Prediction{
		PredictionId:    prediction.ID,
		AnalystId:       prediction.AnalystId,
		PredictionTitle: prediction.Title,
		PredictionDesc:  prediction.Description,
		CreatedAt:       prediction.CreatedAt,
		ViewCount:       prediction.Views,
		IsLocked:        false,
		SelectionList: selectionList,
	}
	return
}
