package serializer

import (
	"log"
	"slices"
	"time"
	"web-api/model"

	fbService "blgit.rfdev.tech/taya/game-service/fb2/outcome_service"
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
	Status   int     `json:"status"`
}

type OddsInfo struct {
	Op  []OddDetail `json:"op"`
	ID  int         `json:"-"`
	Ss  int         `json:"-"`
	Au  int         `json:"-"`
	Mbl int         `json:"-"`
	Li  string      `json:"-"`
}

type MarketGroupInfo struct {
	MarketGroupType    int        `json:"mty"`
	MarketGroupPeriod  int        `json:"pe"`
	MarketGroupName    string     `json:"nm"`
	Mks                []OddsInfo `json:"mks"`
	InternalIdentifier string     `json:"-"`
	Status             int        `json:"status"`
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
	// return finalList
	return SortPredictionList(finalList)
}

func BuildPrediction(prediction model.Prediction, omitAnalyst bool, isLocked bool) (pred Prediction) {
	selectionList := []FbSelectionInfo{}
	// get all odds id that the analyst had selected
	allSelectedOddsId := make([]int64, len(prediction.PredictionSelections))
	// the unknown/black/red status of of the entire PredictionArticle
	predictionStatus := fbService.PredictionOutcomeOutcomeUnknown // first is unknown
	for i, selection := range prediction.PredictionSelections {
		allSelectedOddsId[i] = selection.FbOdds.ID
	}

	for _, selection := range prediction.PredictionSelections {
		marketGroupKey := model.GenerateMarketGroupKeyFromSelection(selection)

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
		// for all odds related to the selection
		for oddIdx, odd := range selection.FbOdds.RelatedOdds {
			orders := model.GetOrderByOddFromSelection(selection, odd.ID)
			oddStatus, err := fbService.ComputeOutcomeByOrderReportI(orders)

			if err != nil {
				log.Printf("error computing outcome for Odds [ID:%d]\n", odd.ID)
			}

			opList[oddIdx] = OddDetail{
				Na:       odd.OddsNameCN,
				Nm:       odd.ShortNameCN,
				Ty:       int(odd.SelectionType),
				Od:       odd.Rate, // not sure
				Bod:      odd.Rate, // not sure
				Odt:      int(odd.OddsFormat),
				Li:       odd.OldNameCN,
				Selected: slices.Contains(allSelectedOddsId, odd.ID),
				Status:   int(oddStatus),
			}
		}

		mks := []OddsInfo{
			{Op: opList},
		}

		mgListIdx := slices.IndexFunc(mgList, func(s MarketGroupInfo) bool {
			return s.InternalIdentifier == marketGroupKey
		})

		if mgListIdx == -1 {
			// market group doesn't exist. add into list for the first time.
			marketGroup := model.GetMarketGroupOrdersByKeyFromPrediction(prediction, marketGroupKey)
			marketgroupStatus, err := fbService.ComputeMarketGroupOutcomesByOrderReport(marketGroup)
			// handle PredictionArticle status
			predictionStatus = fbService.PredictionOutcomeOutcomeRed // default as red first.
			if marketgroupStatus == fbService.MarketGroupOutComeOutcomeUnknown {
				// if any marketgroupStatus is unknown, entire PredictionArticle is unknown
				predictionStatus = fbService.PredictionOutcomeOutcomeUnknown
			} else if marketgroupStatus == fbService.MarketGroupOutComeOutcomeBlack && predictionStatus != fbService.PredictionOutcomeOutcomeUnknown {
				// if any marketgroupStatus is black, and PredictionArticle is not unknown, PredictionArticle is black
				predictionStatus = fbService.PredictionOutcomeOutcomeBlack
			}
			if err != nil {
				log.Printf("error computing marketgroup status: %s\n", err)
			}

			mgList = append(mgList, MarketGroupInfo{
				MarketGroupType:    int(selection.FbOdds.MarketGroupType),
				MarketGroupPeriod:  int(selection.FbOdds.MarketGroupPeriod),
				MarketGroupName:    selection.FbOdds.MarketGroupInfo.FullNameCn,
				Mks:                mks,
				InternalIdentifier: marketGroupKey,
				Status:             int(marketgroupStatus),
			})
		}

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
					Na:   selection.FbMatch.LeagueInfo.LeagueNameCN,
					ID:   int(selection.FbMatch.LeagueInfo.LeagueId),
					Or:   int(selection.FbMatch.LeagueInfo.LeagueLevel),
					Lurl: selection.FbMatch.LeagueInfo.LeagueUrl,
					Sid:  int(selection.FbMatch.LeagueInfo.SportId),
					Rid:  int(selection.FbMatch.LeagueInfo.RegionId),
					Rnm:  selection.FbMatch.LeagueInfo.RegionNameCN,
					Rlg:  selection.FbMatch.LeagueInfo.RegionLogoUrl,
					Hot:  selection.FbMatch.LeagueInfo.IsPopular,
					Slid: int(selection.FbMatch.LeagueInfo.LeagueGroupId),
				},
			})
		} else {
			selectionList[selectionIdx].Mg = mgList
		}
	}

	// predictionStatus, err := fbService.ComputePredictionOutcomesByOrderReport(fbService.Prediction{MarketGroups: allMarketGroups})

	// if err != nil {
	// 	log.Printf("error computing prediction outcome: %s\n", err)
	// }

	if omitAnalyst {
		pred = Prediction{
			PredictionId:    prediction.ID,
			AnalystId:       prediction.AnalystId,
			PredictionTitle: prediction.Title,
			PredictionDesc:  prediction.Content,
			CreatedAt:       prediction.PublishedAt,
			ViewCount:       int64(prediction.Views),
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
			PredictionDesc:  prediction.Content,
			CreatedAt:       prediction.CreatedAt,
			ViewCount:       int64(prediction.Views),
			IsLocked:        isLocked,
			SelectionList:   selectionList,
			AnalystDetail:   &analyst,
			SportId:         GetPredictionSportId(prediction),
			Status:          int64(predictionStatus),
		}
	}
	return
}

func SortPredictionList(predictions []Prediction) []Prediction{
	// sort by status.. unsettled then settled 
	// then in each grp, sort by （命中率 50%，近X中X 50%）
	sorted := slices.Clone(predictions)

	slices.SortFunc(sorted, func(a, b Prediction) int {
		if n:= a.Status - b.Status; n != 0 {
			return int(n)
		} 
		return int(weightage(b) - weightage(a))
	})
	return sorted
}

func weightage(prediction Prediction) float64 {
	if prediction.AnalystDetail != nil {
		return float64(prediction.AnalystDetail.Accuracy) * 0.5 + (float64(prediction.AnalystDetail.RecentWins)/float64(prediction.AnalystDetail.RecentTotal)*100 * 0.5)
	}
	return 0.0 
}

func GetPredictionSportId(p model.Prediction) int64 {
	if len(p.PredictionSelections) == 0 {
		return 0
	} else {
		return int64(p.PredictionSelections[0].FbMatch.SportsID)
	}
}
