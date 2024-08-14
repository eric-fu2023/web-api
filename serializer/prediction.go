package serializer

import (
	"log"
	"slices"
	"time"
	"web-api/model"

	fbService "blgit.rfdev.tech/taya/game-service/fb2/service"
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
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
	AnalystDetail   *Analyst          `json:"analyst_detail,omitempty"`
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
	Selected bool    `json:"selected"`
}

type OddsInfo struct {
	Op     []OddDetail `json:"op"`
	ID     int         `json:"id"`
	Ss     int         `json:"ss"`
	Au     int         `json:"au"`
	Mbl    int         `json:"mbl"`
	Li     string      `json:"li"`
	Status uint8       `json:"status"`
}

type MarketGroupInfo struct {
	MarketGroupType   int        `json:"mty"`
	MarketGroupPeriod int        `json:"pe"`
	MarketGroupName   string     `json:"nm"`
	Mks               []OddsInfo `json:"mks"`
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
	Lg LeagueInfo        `json:"lg"`
	Ts []TeamInfo        `json:"ts"`
	ID int               `json:"id"`
	Bt int64             `json:"bt"`
}

func BuildPredictionsList(predictions []model.Prediction) (preds []Prediction) {
	finalList := make([]Prediction, len(predictions))
	for i, p := range predictions {
		finalList[i] = BuildPrediction(p, false, false)
	}
	return finalList
}

func BuildPrediction(prediction model.Prediction, omitAnalyst bool, isLocked bool) (pred Prediction) {
	selectionList := []FbSelectionInfo{}
	for _, selection := range prediction.PredictionSelections {
		selectionIdx := slices.IndexFunc(selectionList, func(s FbSelectionInfo) bool {
			return s.ID == int(selection.FbMatch.MatchID)
		})

		var mgList []MarketGroupInfo

		if selectionIdx == -1 {
			mgList = []MarketGroupInfo{}
		} else {
			mgList = selectionList[selectionIdx].Mg
		}

		opList := make([]OddDetail, len(selection.FbOdds.RelatedOdds))

		for oddIdx, odd := range selection.FbOdds.RelatedOdds {
			opList[oddIdx] = OddDetail{
				Na:       odd.OddsNameCN,
				Nm:       odd.ShortNameCN,
				Ty:       int(odd.SelectionType),
				Od:       odd.Rate, // not sure
				Bod:      odd.Rate, // not sure
				Odt:      int(odd.OddsFormat),
				Li:       odd.OldNameCN, 
				Selected: odd.ID == selection.FbOdds.ID,
			}
		}
		selectionStatus := GetSelectionStatus(selection)
		mks := []OddsInfo{
			{Op: opList, Status: uint8(selectionStatus)},
		}

		mgList = append(mgList, MarketGroupInfo{
			MarketGroupType:   int(selection.FbOdds.MarketGroupType),
			MarketGroupPeriod: int(selection.FbOdds.MarketGroupPeriod),
			MarketGroupName:   selection.FbMatch.NameCn,
			Mks:               mks,
		})

		if selectionIdx == -1 {
			selectionList = append(selectionList, FbSelectionInfo{
				Bt: selection.FbMatch.StartTimeUtcTs,
				ID: int(selection.FbMatch.MatchID),
				Ts: []TeamInfo{
					{
						Na:   selection.FbMatch.HomeTeam.NameCn,
						ID:   int(selection.FbMatch.HomeTeam.TeamID),
						Lurl: selection.FbMatch.HomeTeam.LogoURL,
					},
					{
						Na:   selection.FbMatch.AwayTeam.NameCn,
						ID:   int(selection.FbMatch.AwayTeam.TeamID),
						Lurl: selection.FbMatch.AwayTeam.LogoURL,
					},
				},
				Mg: mgList,
				Lg: LeagueInfo{
					Na: selection.FbMatch.LeagueInfo.LeagueNameCN,
					ID: int(selection.FbMatch.LeagueInfo.LeagueId),
					Or: int(selection.FbMatch.LeagueInfo.LeagueLevel),
					Lurl: selection.FbMatch.LeagueInfo.LeagueUrl,
					Sid: int(selection.FbMatch.LeagueInfo.SportId),
					Rid: int(selection.FbMatch.LeagueInfo.RegionId),
					Rnm: selection.FbMatch.LeagueInfo.RegionNameCN,
					Rlg: selection.FbMatch.LeagueInfo.RegionLogoUrl,
					Hot: selection.FbMatch.LeagueInfo.IsPopular,
					Slid: int(selection.FbMatch.LeagueInfo.LeagueGroupId),

				}, 
			})
		} else {
			selectionList[selectionIdx].Mg = mgList
		}
	}

	predictionStatus := GetPredictionStatus(prediction)

	if omitAnalyst {
		pred = Prediction{
			PredictionId:    prediction.ID,
			AnalystId:       prediction.AnalystId,
			PredictionTitle: prediction.Title,
			PredictionDesc:  prediction.Description,
			CreatedAt:       prediction.CreatedAt,
			ViewCount:       prediction.Views,
			IsLocked:        isLocked,
			SelectionList:   selectionList,
			SportId:         GetPredictionSportId(prediction),
			Status:          int64(predictionStatus),
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
			IsLocked:        isLocked,
			SelectionList:   selectionList,
			AnalystDetail:   &analyst,
			SportId:         GetPredictionSportId(prediction),
			Status:          int64(predictionStatus),
		}
	}
	return
}

func GetPredictionSportId(p model.Prediction) int64 {
	if len(p.PredictionSelections) == 0 {
		return 0
	} else {
		return int64(p.PredictionSelections[0].FbMatch.SportsID)
	}
}

func GetSelectionStatus(selection model.PredictionSelection) (status fbService.SelectionOutCome) {
	reports := []ploutos.FbBetReport{}

	for _, request := range selection.FbOdds.FbOddsOrderRequestList {
		reports = append(reports, request.FbBetReport)
	}

	status, err := fbService.ComputeOutcomeByOrderReport(reports)
	if err != nil {
		status = fbService.SelectionOutcomeUnknown
		log.Printf("error getting selection status id %d. %s\n", selection.ID, err.Error())
	}
	return
}

func GetPredictionStatus(prediction model.Prediction) (status fbService.SelectionOutCome) {
	selectionStatuses := []fbService.SelectionOutCome{}

	for _, selection := range prediction.PredictionSelections {
		selectionOutcome := GetSelectionStatus(selection)
		selectionStatuses = append(selectionStatuses, selectionOutcome)
	}

	if len(selectionStatuses) == 0 || slices.Contains(selectionStatuses, fbService.SelectionOutcomeUnknown) { // if has any unsettled, whole pred is unsettled
		status = fbService.SelectionOutcomeUnknown
	} else if slices.Contains(selectionStatuses, fbService.SelectionOutcomeBlack) { // if has any black, whole pred is black
		status = fbService.SelectionOutcomeBlack
	} else {
		status = fbService.SelectionOutcomeRed
	}
	return
}
