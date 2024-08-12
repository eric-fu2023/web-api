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
	SelectionList   []FbSelectionInfo `json:"selection_list,omitempty"`
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

type OddDetail struct {
	Na       string  `json:"na"`
	Nm       string  `json:"nm"`
	Ty       int     `json:"ty"`
	Od       float64 `json:"od"`
	Bod      float64 `json:"bod"`
	Odt      int     `json:"odt"`
	Li       string  `json:"li"`
}

type OddsInfo struct {
	Op []OddDetail `json:"op"`
	ID  int    `json:"id"`
	Ss  int    `json:"ss"`
	Au  int    `json:"au"`
	Mbl int    `json:"mbl"`
	Li  string `json:"li"`
}

type MarketGroupInfo struct {
	MarketGroupType int    `json:"mty"`
	MarketGroupPeriod  int    `json:"pe"`
	MarketGroupName  string `json:"nm"`
	Mks OddsInfo `json:"mks"`
}

type LeagueInfo struct {
	Na   string `json:"na"`
	ID   int    `json:"id"`
	Or   int    `json:"or"`
	Lurl string `json:"lurl"`
	Sid  int    `json:"sid"`
	Rid  int    `json:"rid"`
	Rnm  string `json:"rnm"`
	Rlg  string `json:"rlg"`
	Hot  bool   `json:"hot"`
	Slid int    `json:"slid"`
}

type TeamInfo struct {
	Na   string `json:"na"`
	ID   int    `json:"id"`
	Lurl string `json:"lurl"`
}
type FbSelectionInfo struct {
	Mg []MarketGroupInfo `json:"mg"`
	Lg LeagueInfo `json:"lg"`
	Ts []TeamInfo `json:"ts"`
	ID int   `json:"id"`
	Bt int64 `json:"bt"`
}



func BuildPredictionsList(predictions []model.Prediction) (preds []Prediction) {
	finalList := make([]Prediction, len(predictions))
	for i, p := range predictions {
		finalList[i] = BuildPrediction(p, false)
	}
	return finalList
}

func BuildPrediction(prediction model.Prediction, omitAnalyst bool) (pred Prediction) {
	// selectionList := make([]SelectionDetail, len(prediction.PredictionSelections))
	// for j, match := range prediction.PredictionSelections {
	// 	selectionList[j] = SelectionDetail{
	// 		MatchId:           match.MatchId,
	// 		MarketGroupType:   match.FbOdds.MarketGroupType,
	// 		MarketGroupPeriod: match.FbOdds.MarketGroupPeriod,
	// 		OrderMarketlineId: match.FbOdds.RecentMarketlineID,
	// 		MatchType:         int64(match.FbMatch.MatchType),
	// 		MarketGroupName:   "让球",
	// 		LeagueName:        "欧洲杯",
	// 		MatchTime:         time.Now().UnixMilli(),
	// 		MatchName:         "法国vs比利时",
	// 	}
	// }

	selectionList := make([]FbSelectionInfo, len(prediction.PredictionSelections))
	mgList := make([]MarketGroupInfo, len(prediction.PredictionSelections))
	for j, match := range prediction.PredictionSelections {

		// get selection detail using match.selectionid
		opList := make([]OddDetail, len(match.FbOdds.RelatedOdds))

		for oddIdx, odd := range match.FbOdds.RelatedOdds{
			opList[oddIdx] = OddDetail{
				Na: odd.OddsName,
				Nm: "", // TODO 
				Ty: int(odd.SelectionType),
				Od: odd.Rate, // not sure
				Bod: odd.Rate, // not sure
				Odt: int(odd.OddsFormat),
				Li: "", // not saved..
			}
		}
		mks := OddsInfo{
			Op: opList,
		}

		mgList[j] = MarketGroupInfo{
			MarketGroupType: int(match.FbOdds.MarketGroupType),
			MarketGroupPeriod: int(match.FbOdds.MarketGroupPeriod),
			MarketGroupName: match.FbMatch.NameCn,
			Mks: mks,
		}

		selectionList[j] = FbSelectionInfo{
			Bt: match.FbMatch.StartTimeUtcTs,
			ID: int(match.FbMatch.MatchID),
			Ts: []TeamInfo{
				{
					Na: match.FbMatch.HomeTeam.NameCn,
					ID: int(match.FbMatch.HomeTeam.TeamID),
					Lurl: match.FbMatch.HomeTeam.LogoURL,
				},
				{
					Na: match.FbMatch.AwayTeam.NameCn,
					ID: int(match.FbMatch.AwayTeam.TeamID),
					Lurl: match.FbMatch.AwayTeam.LogoURL,
				},
			},
			Mg: mgList,
			// Lg: , // TODO 
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