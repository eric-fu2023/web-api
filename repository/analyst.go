package repository

import (
	"context"
	"errors"
	"web-api/serializer"
)

type MockAnalystRepository struct {
	Analysts []serializer.Analyst
	Err      error
}

func (repo MockAnalystRepository) GetList(ctx context.Context) (r serializer.Response, err error) {

	if len(repo.Analysts) == 0 {
		err = errors.New("")
	}

	// for i, _ := range repo.Analysts {
	// 	repo.Analysts[i].Predictions = NewMockPredictionRepo().Predictions
	// }

	r.Data = repo.Analysts
	return
}

func (repo MockAnalystRepository) GetListPagination(ctx context.Context, page int, limit int) (r serializer.Response, err error) {

	if len(repo.Analysts) == 0 {
		err = errors.New("")
	}

	for i, _ := range repo.Analysts {
		repo.Analysts[i].Predictions = NewMockPredictionRepo().Predictions
	}

	start := limit * (page - 1)
	end := start + limit

	// Ensure start and end are within bounds
	if start >= len(repo.Analysts) {
		return
	}

	if end > len(repo.Analysts) {
		end = len(repo.Analysts)
	}

	repo.Analysts = repo.Analysts[start:end]

	r.Data = repo.Analysts
	return
}

func (repo MockAnalystRepository) GetDetail(ctx context.Context, id int64) (r serializer.Response, err error) {
	// for _, analyst := range repo.Analysts {
	// 	if analyst.AnalystId == id {
	// 		analyst.Predictions = NewMockPredictionRepo().Predictions
	// 		r.Data = analyst
	// 		return 
	// 	}
	// }
	err = errors.New("Invalid id")
	r.Error = err.Error()
	return 
}

func NewMockAnalystRepo() (repo MockAnalystRepository) {
	return MockAnalystRepository{
		Analysts: []serializer.Analyst{
			// {AnalystId: 1, AnalystName: "神预言", AnalystImage: "https://cdn.tayalive.com/aha-img/user/default_user_image/101.jpg", AnalystSource: "足球天下", WinningStreak: 6, Accuracy: 78, AnalystDesc: "特别牛逼"},
			// {AnalystId: 2, AnalystName: "大炮", AnalystImage: "https://cdn.tayalive.com/aha-img/user/default_user_image/102.jpg", AnalystSource: "足球天下", WinningStreak: 6, Accuracy: 59, AnalystDesc: "特别牛逼，牛皮坏了"},
			// {AnalystId: 3, AnalystName: "你们", AnalystImage: "https://cdn.tayalive.com/aha-img/user/default_user_image/102.jpg", AnalystSource: "足球天下", WinningStreak: 6, Accuracy: 81, AnalystDesc: "特别牛逼"},
			// {AnalystId: 4, AnalystName: "我是君杰", AnalystImage: "https://cdn.tayalive.com/aha-img/user/default_user_image/102.jpg", AnalystSource: "足球天下", WinningStreak: 6, Accuracy: 83, AnalystDesc: "特别牛逼"},
			// {AnalystId: 5, AnalystName: "篱竹大宝贝大师傅牛皮死了", AnalystImage: "https://cdn.tayalive.com/aha-img/user/default_user_image/102.jpg", AnalystSource: "足球天下", WinningStreak: 6, Accuracy: 92, AnalystDesc: "特别牛逼"},
		},
		Err: nil,
	}
}
