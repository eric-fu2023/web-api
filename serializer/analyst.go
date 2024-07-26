package serializer

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

// func BuildAnalystList(analysts []models.Analyst) (res []Analyst) {

// for _, analyst := range analysts {

// 	a := Analyst{
// 		AnalystId:     analyst.AnalystId,
// 		AnalystName:   analyst.AnalystName,
// 		AnalystSource: analyst.AnalystSource,
// 	}

// 	res = append(res, a)
// }

// 	return
// }
