package serializer

import "web-api/model"

type UserPrediction struct {
	PredictionId int64 `json:"prediction_id"`
}

func BuildUserPredictionsList(predictions []model.UserPrediction, newPredictions []int64, maxElems int) []UserPrediction {
	var resp []UserPrediction
	for _, pred := range predictions {
		if len(resp) >= maxElems {break}
		resp = append(resp, UserPrediction{
			PredictionId: pred.PredictionId,
		})
	}
	for _, predId := range newPredictions {
		if len(resp) >= maxElems {break}
		resp = append(resp, UserPrediction{
			PredictionId: predId,
		})
	}

	return resp
}

func BuildPredictions(userPreds []model.UserPrediction) []Prediction {
	preds := []Prediction{}

	for _, userPred := range userPreds {
		preds = append(preds, Prediction{PredictionId: userPred.PredictionId, AnalystId: userPred.Prediction.AnalystId, PredictionTitle: userPred.Prediction.PredictionTitle, PredictionDesc: userPred.Prediction.PredictionDesc, IsLocked: userPred.IsLocked})
	}
	return preds
}