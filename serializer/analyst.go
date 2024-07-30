package serializer

import (
	"web-api/model"
)

type Analyst struct {
	AnalystId        int64        `json:"analyst_id"`
	AnalystName      string       `json:"analyst_name"`
	AnalystSource    string       `json:"analyst_source"`
	AnalystImage     string       `json:"analyst_image"`
	WinningStreak    int          `json:"winning_streak"`
	Accuracy         int          `json:"accuracy"`
	AnalystDesc      string       `json:"analyst_desc"`
	Predictions      []Prediction `json:"predictions"`
	NumFollowers     int          `json:"num_followers"`
	TotalPredictions int          `json:"total_predictions"`
}

func BuildAnalysts(analysts []model.Analyst) (resp []Analyst) {
	for _, a := range analysts {
		resp = append(resp, Analyst{
			AnalystId: a.ID,
			AnalystName: a.Name,
			AnalystSource: a.Source.Name,
			AnalystImage: "https://cdn.tayalive.com/aha-img/user/default_user_image/102.jpg",
			WinningStreak: 20,
			Accuracy: 0,
			AnalystDesc: a.Desc,
			Predictions: []Prediction{},
			NumFollowers: 0,
			TotalPredictions: 0,
		})
	}
	return 
}

func BuildAnalystDetail(analyst model.Analyst) (resp Analyst) {
	resp = Analyst{
		AnalystId: analyst.ID,
		AnalystName: analyst.Name,
		AnalystSource: analyst.Source.Name,
		AnalystImage: "https://cdn.tayalive.com/aha-img/user/default_user_image/102.jpg",
		WinningStreak: 20,
		Accuracy: 0,
		AnalystDesc: analyst.Desc,
		Predictions: []Prediction{},
		NumFollowers: 0,
		TotalPredictions: 0,
	}
	return 
}

// func BuildAnalystList(analysts []models.Analyst) (res []Analyst) {

// 	for _, analyst := range analysts {

// 		res = append(res, BuildAnalyst(analyst))
// 	}

// 	return
// }

// func BuildAnalyst(analyst models.Analyst) (a Analyst) {

// 	a = Analyst{
// 		AnalystId:     analyst.AnalystId,
// 		AnalystName:   analyst.AnalystName,
// 		AnalystSource: analyst.AnalystSource,
// 	}

// 	return
// }
