package serializer

import (
	"time"
	"web-api/model"
)

type UserPrediction struct {
	PredictionId int64 `json:"prediction_id"`
}

func BuildUserPredictionsWithLock(preds []model.Prediction, userPreds []model.UserPrediction) []Prediction {
	// Create a map to quickly check if a PredictionId is in userPreds
	userPredMap := make(map[uint]bool, len(userPreds))
	for _, up := range userPreds {
		userPredMap[uint(up.PredictionId)] = true
	}

	// Build the list of predictions
	ls := make([]Prediction, len(preds))
	for i, pred := range preds {
		_, locked := userPredMap[uint(pred.ID)] // If PredictionId exists in userPredMap, locked will be true

		selectionList := make([]SelectionDetail, len(pred.PredictionSelections))
		for j, match := range pred.PredictionSelections {
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

		ls[i] = Prediction{
			PredictionId:    pred.ID,
			AnalystId:       pred.AnalystId,
			PredictionTitle: pred.Title,
			PredictionDesc:  pred.Description,
			CreatedAt:       pred.CreatedAt,
			ViewCount:       pred.Views,
			IsLocked:        !locked, // If it's not in userPredMap, it's locked
			SelectionList:   selectionList,
			AnalystDetail:   BuildAnalystDetail(pred.AnalystDetail),
		}
	}

	return ls
}
