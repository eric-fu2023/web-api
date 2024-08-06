package serializer

import (
	"web-api/model"
)

type Analyst struct {
	AnalystId        int64        `json:"analyst_id"`
	AnalystName      string       `json:"analyst_name"`
	AnalystDesc      string       `json:"analyst_desc"`
	AnalystSource    Source       `json:"analyst_source"`
	AnalystImage     string       `json:"analyst_image"`
	WinningStreak    int          `json:"winning_streak"`
	Accuracy         int          `json:"accuracy"`
	NumFollowers     int          `json:"num_followers"`
	TotalPredictions int          `json:"total_predictions"`
	Predictions      []Prediction `json:"predictions"`
}

type Source struct {
	Name string `json:"source_name"`
	Icon string `json:"source_icon"`
}

func BuildAnalysts(analysts []model.Analyst) (resp []Analyst) {
	for _, a := range analysts {
		resp = append(resp, Analyst{
			AnalystId:        a.ID,
			AnalystName:      a.Name,
			AnalystSource:    Source{Name: a.TipsAnalystSource.Name, Icon: a.TipsAnalystSource.ImgIcon},
			AnalystImage:     "https://cdn.tayalive.com/aha-img/user/default_user_image/102.jpg",
			WinningStreak:    20,
			Accuracy:         0,
			AnalystDesc:      a.Desc,
			Predictions:      []Prediction{},
			NumFollowers:     len(a.Followers),
			TotalPredictions: len(a.Predictions),
		})
	}
	return
}

func BuildAnalystDetail(analyst model.Analyst) (resp Analyst) {
	predList := make([]model.Prediction, len(analyst.Predictions))
	for i, pred := range analyst.Predictions {
		predList[i] = model.Prediction{pred, []model.PredictionSelection{}}
	}

	resp = Analyst{
		AnalystId:        analyst.ID,
		AnalystName:      analyst.Name,
		AnalystSource:    Source{Name: analyst.TipsAnalystSource.Name, Icon: analyst.TipsAnalystSource.ImgIcon},
		AnalystImage:     "https://cdn.tayalive.com/aha-img/user/default_user_image/102.jpg",
		WinningStreak:    20,
		Accuracy:         0,
		AnalystDesc:      analyst.Desc,
		Predictions:      BuildPredictionsList(predList),
		NumFollowers:     len(analyst.Followers),
		TotalPredictions: len(analyst.Predictions),
	}
	return
}

func BuildFollowingList(followings []model.UserAnalystFollowing) (resp []Analyst) {
	for _, a := range followings {
		resp = append(resp, Analyst{
			AnalystId:        a.ID,
			AnalystName:      a.Analyst.Name,
			AnalystSource:    Source{Name: a.Analyst.TipsAnalystSource.Name, Icon: a.Analyst.TipsAnalystSource.ImgIcon},
			AnalystImage:     "https://cdn.tayalive.com/aha-img/user/default_user_image/102.jpg",
			WinningStreak:    20,
			Accuracy:         0,
			AnalystDesc:      a.Analyst.Desc,
			Predictions:      []Prediction{},
			NumFollowers:     len(a.Analyst.Followers),
			TotalPredictions: len(a.Analyst.Predictions),
		})
	}
	return
}
