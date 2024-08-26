package serializer

import (
	"web-api/model"
)

type UserPrediction struct {
	PredictionId int64 `json:"prediction_id"`
}

func BuildUserPredictionsWithLock(preds []model.Prediction, userPreds []model.UserPrediction, page, limit, brandId int) []Prediction {
	// Create a map to quickly check if a PredictionId is in userPreds
	userPredMap := make(map[uint]bool, len(userPreds))
	for _, up := range userPreds {
		userPredMap[uint(up.PredictionId)] = true
	}

	// Build the list of predictions
	ls := make([]Prediction, len(preds))
	for i, pred := range preds {
		_, exist := userPredMap[uint(pred.ID)] // If PredictionId exists in userPredMap, locked will be true

		ls[i] = BuildPrediction(pred, false, !exist, brandId)
	}

	return SortPredictionList(ls, page, limit)
}
