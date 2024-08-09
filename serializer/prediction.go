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
	Status          int64             `json:"status"`
	AnalystDetail   *Analyst           `json:"analyst_detail,omitempty"`
	SportId         int64             `json:"sport_id"`
}

type PredictedMatch struct {
	MatchId int64 `json:"match_id"`
	FbBetId int64 `json:"fb_bet_id"`
	Status  int64 `json:"status"`
}

type SelectionDetail struct {
	MarketGroupType   int64  `json:"mty"`
	MarketGroupPeriod int64  `json:"pe"`
	OrderMarketlineId int64  `json:"id"`
	MatchType         int64  `json:"ty"`
	MatchId           int64  `json:"match_id"`
	MarketGroupName   string `json:"mgnm"`
	LeagueName        string `json:"lgna"`
	MatchTime         int64  `json:"bt"`
	MatchName         string `json:"nm"`
}

func BuildPredictionsList(predictions []model.Prediction) (preds []Prediction) {
	finalList := make([]Prediction, len(predictions))
	for i, p := range predictions {
		finalList[i] = BuildPrediction(p, false)
	}
	return finalList
}

func BuildPrediction(prediction model.Prediction, omitAnalyst bool) (pred Prediction) {
	selectionList := make([]SelectionDetail, len(prediction.PredictionSelections))
	for j, match := range prediction.PredictionSelections {
		selectionList[j] = SelectionDetail{
			MatchId:           match.MatchId,
			MarketGroupType:   match.FbOdds.MarketGroupType,
			MarketGroupPeriod: match.FbOdds.MarketGroupPeriod,
			OrderMarketlineId: match.FbOdds.RecentMarketlineID,
			MatchType:         int64(match.FbMatch.MatchType),
			MarketGroupName:   "让球",
			LeagueName:        "欧洲杯",
			MatchTime:         time.Now().UnixMilli(),
			MatchName:         "法国vs比利时",
		}
	}

	if omitAnalyst {
		pred = Prediction{
			PredictionId:    prediction.ID,
			AnalystId:       prediction.AnalystId,
			PredictionTitle: prediction.Title,
			PredictionDesc:  prediction.Description,
			CreatedAt:       prediction.CreatedAt,
			ViewCount:       prediction.Views,
			IsLocked:        false,
			SelectionList:   selectionList,
			SportId:         GetPredictionSportId(prediction),
		}
	} else {
		analyst := BuildAnalystDetail(prediction.AnalystDetail)
		pred = Prediction{
			PredictionId:    prediction.ID,
			AnalystId:       prediction.AnalystId,
			PredictionTitle: prediction.Title,
			PredictionDesc:  prediction.Description,
			CreatedAt:       prediction.CreatedAt,
			ViewCount:       prediction.Views,
			IsLocked:        false,
			SelectionList:   selectionList,
			AnalystDetail:   &analyst,
			SportId:         GetPredictionSportId(prediction),
		}
	}
	return
}

func GetPredictionSportId(p model.Prediction) int64{
	if len(p.PredictionSelections) == 0 {
		return 0
	} else {
		return p.PredictionSelections[0].FbMatch.SportsID
	}
}