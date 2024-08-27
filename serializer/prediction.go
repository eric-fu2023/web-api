package serializer

import (
	"encoding/json"
	"slices"
	"time"
	"web-api/model"

	models "blgit.rfdev.tech/taya/ploutos-object"
)

type Prediction struct {
	PredictionId    int64           `json:"prediction_id"`
	AnalystId       int64           `json:"analyst_id"`
	PredictionTitle string          `json:"prediction_title"`
	PredictionDesc  string          `json:"prediction_desc"`
	IsLocked        bool            `json:"is_locked"`
	CreatedAt       time.Time       `json:"created_at"`
	ViewCount       int64           `json:"view_count"`
	SelectionList   []SelectionInfo `json:"selection_list,omitempty"`
	Status          int64           `json:"status"`
	AnalystDetail   *Analyst        `json:"analyst_detail,omitempty"`
	SportId         int64           `json:"sport_id"`
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
	ID  int64       `json:"id"`
	Ss  int         `json:"ss"`
	Au  int         `json:"au"`
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
type SelectionInfo struct {
	Mg []MarketGroupInfo `json:"mg"`
	Lg LeagueInfo        `json:"lg"`
	Ts []TeamInfo        `json:"ts"`
	ID int               `json:"id"`
	Bt int64             `json:"bt"`

	// for imsb
	ImsbMg []ImsbSelectionInfo `json:"odds,omitempty"`
}

type ImsbSelectionInfo struct {
	BetID      string           `json:"bet_id"`
	BetName    string           `json:"bet_name"`
	BetStatus  int              `json:"bet_status"`
	BetLocked  bool             `json:"bet_locked"`
	BetTypeID  int              `json:"bet_type_id"`
	BetMarket  int              `json:"bet_market"`
	Match      ImsbMatchDetail  `json:"match"`
	OddsDetail []ImsbOddsDetail `json:"odds"`
	Priority   int              `json:"priority"`
}

type ImsbOddsDetail struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	Value            float64 `json:"value"`
	Status           int     `json:"status"`
	Market           int     `json:"market"`
	IsLocked         bool    `json:"is_locked"`
	OddsStatus       int     `json:"odds_status"`
	BetStatus        int     `json:"bet_status"`
	PredictionStatus int     `json:"predict_status"`
	IsSelected       bool    `json:"is_selected"`
}

type ImsbMatchDetail struct {
	CricketMatchID int    `json:"cricket_match_id"`
	Title          string `json:"title"`
	ImMatchID      int    `json:"im_match_id"`
}

func BuildPredictionsList(predictions []model.Prediction, brandId model.BrandId) (preds []Prediction) {
	finalList := make([]Prediction, len(predictions))
	for i, p := range predictions {
		finalList[i] = BuildPrediction(p, false, false, brandId)
	}
	return SortPredictionList(finalList, page, limit)
}

func BuildFbPrediction(prediction model.Prediction, omitAnalyst bool, isLocked bool) (pred Prediction) {
	selectionList := []SelectionInfo{}
	// get all odds id that the analyst had selected
	allSelectedOddsId := make([]int64, len(prediction.PredictionSelections))
	// the unknown/black/red status of of the entire PredictionArticle
	// predictionStatus := fbService.PredictionOutcomeOutcomeUnknown // first is unknown
	for i, selection := range prediction.PredictionSelections {
		allSelectedOddsId[i] = selection.FbOdds.ID
	}

	for _, selection := range prediction.PredictionSelections {
		marketGroupKey := model.GenerateMarketGroupKeyFromSelection(selection)

		selectionIdx := slices.IndexFunc(selectionList, func(s SelectionInfo) bool {
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
			// get prediction bet result first
			// if it's red/black already, use directly
			// otherwise compute from taya_bet_report

			/* no need to compute, use DB as source of truth */
			// betResult := 0
			// if selection.BetResult == models.BetResultUnknown {
			// 	if oddStatus, err := fbService.ComputeOutcomeByOrderReportI(model.GetOrderByOddFromSelection(selection, odd.ID)); err != nil {
			// 		log.Printf("error computing outcome for Odds [ID:%d]: %s\n", odd.ID, err)
			// 	} else {
			// 		betResult = int(oddStatus)
			// 	}
			// } else {
			// 	betResult = int(selection.BetResult)
			// }

			betResult := models.BetResultUnknown
			for _, sel := range prediction.PredictionSelections {
				if odd.ID == sel.FbOdds.ID {
					betResult = sel.BetResult
				}
			}
			opList[oddIdx] = OddDetail{
				Na:       odd.OddsNameCN,
				Nm:       odd.ShortNameCN, // odd.ShortNameCN,
				Ty:       int(odd.SelectionType),
				Od:       -999, //odd.Rate, // not sure
				Bod:      -999, //odd.Rate, // not sure
				Odt:      int(odd.OddsFormat),
				Li:       odd.OldNameCN,
				Selected: slices.Contains(allSelectedOddsId, odd.ID),
				Status:   int(betResult),
			}
		}

		mks := []OddsInfo{
			{
				Op: opList,
				ID: selection.FbOdds.RecentMarketlineID,
				Ss: 1,
				Au: 1,
			},
		}

		mgListIdx := slices.IndexFunc(mgList, func(s MarketGroupInfo) bool {
			return s.InternalIdentifier == marketGroupKey
		})

		if mgListIdx == -1 {
			// // market group doesn't exist. add into list for the first time.
			// marketGroup := model.GetMarketGroupOrdersByKeyFromPrediction(prediction, marketGroupKey)
			// marketgroupStatus, err := fbService.ComputeMarketGroupOutcomesByOrderReport(marketGroup)
			// // handle PredictionArticle status
			// predictionStatus = fbService.PredictionOutcomeOutcomeRed // default as red first.
			// if marketgroupStatus == fbService.MarketGroupOutComeOutcomeUnknown {
			// 	// if any marketgroupStatus is unknown, entire PredictionArticle is unknown
			// 	predictionStatus = fbService.PredictionOutcomeOutcomeUnknown
			// } else if marketgroupStatus == fbService.MarketGroupOutComeOutcomeBlack && predictionStatus != fbService.PredictionOutcomeOutcomeUnknown {
			// 	// if any marketgroupStatus is black, and PredictionArticle is not unknown, PredictionArticle is black
			// 	predictionStatus = fbService.PredictionOutcomeOutcomeBlack
			// }
			// if err != nil {
			// 	log.Printf("error computing marketgroup status: %s\n", err)
			// }

			mgStatus := models.BetResultUnknown
			for _, op := range opList {
				if op.Status == int(models.BetResultWin) {
					// if any op win, whole mg win
					mgStatus = models.BetResultWin
					break
				}
				// all lose
				mgStatus = models.BetResultLose
			}

			mgList = append(mgList, MarketGroupInfo{
				MarketGroupType:    int(selection.FbOdds.MarketGroupType),
				MarketGroupPeriod:  int(selection.FbOdds.MarketGroupPeriod),
				MarketGroupName:    CustomizeOddsName(selection.FbOdds.MarketGroupInfo.FullNameCn),
				Mks:                mks,
				InternalIdentifier: marketGroupKey,
				Status:             int(mgStatus),
			})
		}

		if selectionIdx == -1 {
			selectionList = append(selectionList, SelectionInfo{
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
			IsLocked:        prediction.PredictionResult == models.PredictionResultUnknown && isLocked,
			SelectionList:   selectionList,
			SportId:         int64(prediction.FbSportId),
			Status:          int64(prediction.PredictionResult),
		}
	} else {
		analyst := BuildAnalystDetail(prediction.AnalystDetail, model.BrandIdAha)
		pred = Prediction{
			PredictionId:    prediction.ID,
			AnalystId:       prediction.AnalystId,
			PredictionTitle: prediction.Title,
			PredictionDesc:  prediction.Content,
			CreatedAt:       prediction.CreatedAt,
			ViewCount:       int64(prediction.Views),
			IsLocked:        prediction.PredictionResult == models.PredictionResultUnknown && isLocked,
			SelectionList:   selectionList,
			AnalystDetail:   &analyst,
			SportId:         int64(prediction.FbSportId),
			Status:          int64(prediction.PredictionResult),
		}
	}
	return
}

func SortPredictionList(predictions []Prediction, page,limit int) []Prediction {
	// sort by status.. unsettled then settled
	// then in each grp, sort by （命中率 50%，近X中X 50%）
	unsettled := []Prediction{}
	settled := []Prediction{}

	filteredPredictions := []Prediction{}
	_y, _m, _d := time.Now().AddDate(0, 0, -7).Date()
	weekAgo := time.Date(_y, _m, _d, 0, 0, 0, 0, time.Now().Location())
	for _, pred := range predictions {
		if pred.CreatedAt.After(weekAgo) {
			filteredPredictions = append(filteredPredictions, pred)
		}
	}

	for _, pred := range filteredPredictions {
		if pred.Status == 0 {
			unsettled = append(unsettled, pred)
		} else {
			settled = append(settled, pred)
		}
	}

	// slices.SortFunc(unsettled, func(a, b Prediction) int {
	// 	wa, wb := weightage(a), weightage(b)
	// 	if wa < wb {
	// 		return 1
	// 	} else if wa > wb {
	// 		return -1
	// 	}
	// 	return 0
	// })
	// slices.SortFunc(settled, func(a, b Prediction) int {
	// 	wa, wb := weightage(a), weightage(b)
	// 	if wa < wb {
	// 		return 1
	// 	} else if wa > wb {
	// 		return -1
	// 	}
	// 	return 0
	// })
	// FIXME : need to fix weightage
	slices.SortFunc(settled, func(a, b Prediction) int {
		if weightage(a) < weightage(b) {
			return 1
		} else if weightage(a) > weightage(b) {
			return -1
		} else {
			return 0
		}
	})

	slices.SortFunc(unsettled, func(a, b Prediction) int {
		if weightage(a) < weightage(b) {
			return 1
		} else if weightage(a) > weightage(b) {
			return -1
		} else {
			return 0
		}
	})

	finalList := append(unsettled, settled...)

	start := limit * (page - 1)
	if start > len(finalList) {
		return []Prediction{}
	}
	end := limit * page
	if end > len(finalList) {
		end = len(finalList)
	}

	return finalList[start:end]
}

func weightage(prediction Prediction) float64 {
	if prediction.AnalystDetail == nil {
		return 0.00
	}
	accuracyWeight := float64(prediction.AnalystDetail.Accuracy) * 0.5
	nearXweight := ((float64(prediction.AnalystDetail.RecentWins) / float64(prediction.AnalystDetail.RecentTotal)) * 100 * 0.5)
	return accuracyWeight + nearXweight
}

func CustomizeOddsName(oddsName string) string {
	customizedOddsNames := map[string]string{"独赢": "胜平负"}
	if customizedOddsName, exists := customizedOddsNames[oddsName]; exists {
		return customizedOddsName
	} else {
		return oddsName
	}
}

func BuildMockImsbPrediction(prediction model.Prediction, omitAnalyst bool, isLocked bool) (pred Prediction) {
	data := Prediction{}

	err := json.Unmarshal([]byte(mockData), &data)

	if err != nil {
		return
	}
	return data

}

func BuildMockSelectionListImsb() (ls []SelectionInfo){
	err := json.Unmarshal([]byte(mockSelectionList), &ls)

	if err != nil {
		return
	}
	return ls
}

var mockData = `
{
   "prediction_id": 46,
   "analyst_id": 27,
   "prediction_title": "美女方案",
   "prediction_desc": "<p>啊的发疯的发生地方</p>",
   "is_locked": false,
   "created_at": "2024-08-22T02:24:28.628004Z",
   "view_count": 16,
   "selection_list": [
      {
         "odds": [
            {
               "bet_id": "f041ea4b-f758-4bee-98ed-a105693585f1",
               "bet_name": "Match - Head to Head (inc. Super Over) - Live",
               "bet_status": 1,
               "bet_locked": false,
               "bet_type_id": 4,
               "bet_market": 3,
               "match": {
                  "cricket_match_id": 0,
                  "title": "Kolkata Knight Riders SRL VS Punjab Kings SRL",
                  "im_match_id": 59249875
               },
               "odds": [
                  {
                     "id": 3574733796,
                     "name": "Home",
                     "value": 0,
                     "status": 1,
                     "market": 3,
                     "is_locked": false,
                     "odds_status": 1,
                     "bet_status": 1,
					 "predict_status":0,
					 "is_selected":false
                  },
                  {
                     "id": 3574733797,
                     "name": "Away",
                     "value": 0,
                     "status": 1,
                     "market": 3,
                     "is_locked": false,
                     "odds_status": 1,
                     "bet_status": 1,
					 "predict_status":2,
					 "is_selected":true
                  }
               ],
               "priority": 2
            },
            {
               "bet_id": "d66c7792-8288-435b-b5d2-227829653987",
               "bet_name": "Innings 1 - Total Runs for Over 3 - Punjab Kings SRL - Live",
               "bet_status": 1,
               "bet_locked": false,
               "bet_type_id": 99,
               "bet_market": 3,
               "match": {
                  "cricket_match_id": 0,
                  "title": "Kolkata Knight Riders SRL VS Punjab Kings SRL",
                  "im_match_id": 59249875
               },
               "odds": [
                  {
                     "id": 3576873744,
                     "name": "Over 8.5",
                     "value": 0,
                     "status": 1,
                     "market": 3,
                     "is_locked": false,
                     "odds_status": 1,
                     "bet_status": 1,
					 "predict_status":2,
					 "is_selected":true
                  },
                  {
                     "id": 3576873747,
                     "name": "Under 8.5",
                     "value": 0,
                     "status": 1,
                     "market": 3,
                     "is_locked": false,
                     "odds_status": 1,
                     "bet_status": 1,
					 "predict_status":1,
					 "is_selected":true
                  }
               ],
               "priority": 26
            }
         ],
         "lg": {
            "na": "爱尔兰甲级联赛",
            "id": 10937,
            "or": 1586,
            "lurl": "https://static.fastbs55.com/data/b7156de3cd38c823863d7e35df595908.png",
            "sid": 1,
            "rid": 70,
            "rnm": "爱尔兰",
            "rlg": "https://static.fastbs55.com/data/28b56ea98e9c4b353e574757cb33d90f.png",
            "hot": false,
            "slid": 109370000
         },
         "ts": [
            {
               "na": "科布漫步者",
               "id": 55479,
               "lurl": "https://static.fastbs55.com/data/9970a75e1ee535a3c140cc0fcc3ab44c.png"
            },
            {
               "na": "科克城",
               "id": 55424,
               "lurl": "https://static.fastbs55.com/data/53f5e1f95af79ac48ac756212df43a51.png"
            }
         ],
         "id": 59249875,
         "bt": 1724438700000
      }
   ],
   "status": 0,
   "analyst_detail": {
      "analyst_id": 27,
      "analyst_name": "美女球探",
      "analyst_desc": "{球探名字}是一位经验丰富的足球球探，专注于挖掘和评估年轻足球天才。他/她在足球界拥有多年经验，曾为多家顶级俱乐部和国家队提供过出色的球员推荐服务。{球探名字}以敏锐的眼光和深厚的战术知识著称，能够在早期识别出球员的潜力，并帮助他们迈向职业生涯的巅峰。  凭借广泛的人脉和深刻的洞察力，{球探名字}在全球范围内建立了广泛的球员观察网络，擅长发现隐藏的足球人才，帮助俱乐部和球队建立起具有长期竞争力的阵容。他/她对年轻球员的全面评估和独到见解，深受足球界的高度认可。",
      "analyst_source": {
         "source_name": "九品芝麻官之白面包青天",
         "source_icon": "https://static.tayalive.com/aha-img/prediction_source_icon/prediction_source_icon-20240816090644-gAOnKB.jpg"
      },
      "analyst_image": "https://static.tayalive.com/aha-img/prediction_analyst_avatar/prediction_analyst_avatar-20240816090157-qHr8c3.jpg",
      "winning_streak": 0,
      "accuracy": 0,
      "num_followers": 0,
      "total_predictions": 2,
      "predictions": [
         {
            "prediction_id": 46,
            "analyst_id": 27,
            "prediction_title": "美女方案",
            "prediction_desc": "<p>啊的发疯的发生地方</p>",
            "is_locked": false,
            "created_at": "2024-08-22T02:24:38.917257Z",
            "view_count": 16,
            "status": 0,
            "sport_id": 1
         },
         {
            "prediction_id": 21,
            "analyst_id": 27,
            "prediction_title": "asda",
            "prediction_desc": "<p>asdsada</p>",
            "is_locked": false,
            "created_at": "2024-08-16T08:41:17.766199Z",
            "view_count": 32,
            "status": 2,
            "sport_id": 1
         }
      ],
      "recent_total": 0,
      "recent_wins": 0,
      "is_show_streak": false,
      "is_show_accuracy": false,
      "is_show_recent_wins": false
   },
   "sport_id": 1
}
`
var mockSelectionList = `
[
      {
         "odds": [
            {
               "bet_id": "f041ea4b-f758-4bee-98ed-a105693585f1",
               "bet_name": "Match - Head to Head (inc. Super Over) - Live",
               "bet_status": 1,
               "bet_locked": false,
               "bet_type_id": 4,
               "bet_market": 3,
               "match": {
                  "cricket_match_id": 0,
                  "title": "Kolkata Knight Riders SRL VS Punjab Kings SRL",
                  "im_match_id": 59249875
               },
               "odds": [
                  {
                     "id": 3574733796,
                     "name": "Home",
                     "value": 0,
                     "status": 1,
                     "market": 3,
                     "is_locked": false,
                     "odds_status": 1,
                     "bet_status": 1,
					 "predict_status":0,
					 "is_selected":false
                  },
                  {
                     "id": 3574733797,
                     "name": "Away",
                     "value": 0,
                     "status": 1,
                     "market": 3,
                     "is_locked": false,
                     "odds_status": 1,
                     "bet_status": 1,
					 "predict_status":2,
					 "is_selected":true
                  }
               ],
               "priority": 2
            },
            {
               "bet_id": "d66c7792-8288-435b-b5d2-227829653987",
               "bet_name": "Innings 1 - Total Runs for Over 3 - Punjab Kings SRL - Live",
               "bet_status": 1,
               "bet_locked": false,
               "bet_type_id": 99,
               "bet_market": 3,
               "match": {
                  "cricket_match_id": 0,
                  "title": "Kolkata Knight Riders SRL VS Punjab Kings SRL",
                  "im_match_id": 59249875
               },
               "odds": [
                  {
                     "id": 3576873744,
                     "name": "Over 8.5",
                     "value": 0,
                     "status": 1,
                     "market": 3,
                     "is_locked": false,
                     "odds_status": 1,
                     "bet_status": 1,
					 "predict_status":2,
					 "is_selected":true
                  },
                  {
                     "id": 3576873747,
                     "name": "Under 8.5",
                     "value": 0,
                     "status": 1,
                     "market": 3,
                     "is_locked": false,
                     "odds_status": 1,
                     "bet_status": 1,
					 "predict_status":1,
					 "is_selected":true
                  }
               ],
               "priority": 26
            }
         ],
         "lg": {
            "na": "爱尔兰甲级联赛",
            "id": 10937,
            "or": 1586,
            "lurl": "https://static.fastbs55.com/data/b7156de3cd38c823863d7e35df595908.png",
            "sid": 1,
            "rid": 70,
            "rnm": "爱尔兰",
            "rlg": "https://static.fastbs55.com/data/28b56ea98e9c4b353e574757cb33d90f.png",
            "hot": false,
            "slid": 109370000
         },
         "ts": [
            {
               "na": "科布漫步者",
               "id": 55479,
               "lurl": "https://static.fastbs55.com/data/9970a75e1ee535a3c140cc0fcc3ab44c.png"
            },
            {
               "na": "科克城",
               "id": 55424,
               "lurl": "https://static.fastbs55.com/data/53f5e1f95af79ac48ac756212df43a51.png"
            }
         ],
         "id": 59249875,
         "bt": 1724438700000
      }
   ]
	  `