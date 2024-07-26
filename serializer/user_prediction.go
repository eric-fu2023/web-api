package serializer

import "web-api/model"

type UserPrediction struct {
	PredictionId int64 `json:"prediction_id"`
}

func BuildUserPredictionsList(predictions []model.UserPrediction, newPredictions []int64) []UserPrediction {
	var resp []UserPrediction
	for _, pred := range predictions {
		resp = append(resp, UserPrediction{
			PredictionId: pred.PredictionId,
		})
	}
	for _, predId := range newPredictions {
		resp = append(resp, UserPrediction{
			PredictionId: predId,
		})
	}

	return resp
}