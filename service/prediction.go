package service

import (
	"errors"
	"slices"
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util"
	"web-api/util/i18n"

	"github.com/gin-gonic/gin"
)

type PredictionListService struct {
	common.Page
	AnalystId int64 `json:"analyst_id" form:"analyst_id"`
	FbMatchId int64 `json:"fb_match_id" form:"fb_match_id"`
	SportId   int64 `json:"sports_id" form:"sports_id"`
}

func (service *PredictionListService) List(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")

	user := model.User{}

	if u != nil {
		user = u.(model.User)
	}

	deviceInfo, err := util.GetDeviceInfo(c)
	if err != nil {
		return
	}

	hasAuth := user.ID != 0

	if err != nil {
		r = serializer.DBErr(c, service, i18n.T("general_error"), err)
		return
	}

	var predictions []model.Prediction
	var userPredictions []model.UserPrediction
	if hasAuth {
		u, _ := c.Get("user")
		user := u.(model.User)

		hasPaymentToday := false
		hasPaymentToday, err = model.HasTopupToday(c, user.ID)

		if err != nil {
			r = serializer.DBErr(c, service, i18n.T("general_error"), err)
			return
		}

		whitelist := []int64{2055, 3366, 2064, 2051, 2058, 2288, 16004, 2287, 2302}
		// TEMP, to be removed
		isWhitelisted := slices.Contains(whitelist, user.ID)

		if hasPaymentToday || isWhitelisted {
			// logged in, has payment - show all
			predictions, err = model.ListPredictions(model.ListPredictionCond{Limit: service.Limit, Page: service.Page.Page, AnalystId: service.AnalystId, FbMatchId: service.FbMatchId, SportId: service.SportId})
			if err != nil {
				r = serializer.DBErr(c, service, i18n.T("general_error"), err)
				return
			}

			r.Msg = i18n.T("success")
			r.Data = serializer.BuildPredictionsList(predictions)
			return

		} else {
			// logged in, no payment - unlock 3 max
			predictions, err = model.ListPredictions(model.ListPredictionCond{Limit: service.Limit, Page: service.Page.Page, AnalystId: service.AnalystId, FbMatchId: service.FbMatchId, SportId: service.SportId})
			if err != nil {
				r = serializer.DBErr(c, service, i18n.T("general_error"), err)
				return
			}
			userPredictions, err = model.GetUserPrediction(model.GetUserPredictionCond{DeviceId: user.LastLoginDeviceUuid, UserId: user.ID})
			if err != nil {
				r = serializer.DBErr(c, service, i18n.T("general_error"), err)
				return
			}

			r.Msg = i18n.T("success")
			r.Data = serializer.BuildUserPredictionsWithLock(predictions, userPredictions)
			return
		}

	} else {
		// not logged in, unlock 1
		predictions, err = model.ListPredictions(model.ListPredictionCond{Limit: service.Limit, Page: service.Page.Page, AnalystId: service.AnalystId, FbMatchId: service.FbMatchId, SportId: service.SportId})
		if err != nil {
			r = serializer.DBErr(c, service, i18n.T("general_error"), err)
			return
		}
		userPredictions, err = model.GetUserPrediction(model.GetUserPredictionCond{DeviceId: deviceInfo.Uuid, UserId: 0})
		if err != nil {
			r = serializer.DBErr(c, service, i18n.T("general_error"), err)
			return
		}

		r.Msg = i18n.T("success")
		r.Data = serializer.BuildUserPredictionsWithLock(predictions, userPredictions[:1])
		return
	}

}

type PredictionDetailService struct {
	PredictionId int64 `json:"prediction_id" form:"prediction_id"`
}

func (service *PredictionDetailService) GetDetail(c *gin.Context) (r serializer.Response, err error) {

	data, err := model.GetPrediction(service.PredictionId)

	// if (service.PredictionId == 8) {
	// 	var jsonData map[string]interface{}
	// 	err = json.Unmarshal([]byte(dummyJson), &jsonData)
	// 	if err != nil {
	// 		return
	// 	}
	// 	r.Data = jsonData
	// 	return
	// }

	if err != nil {
		r = serializer.DBErr(c, service, "", err)
		return
	}

	if err = model.IncreasePredictionViewCountBy1(data); err != nil {
		r = serializer.DBErr(c, service, "", err)
		return
	}

	r.Data = serializer.BuildPrediction(data, false, false)

	return
}

type AddUserPredictionService struct {
	UserId       int64 `json:"user_id" form:"user_id"`
	PredictionId int64 `json:"prediction_id" form:"prediction_id"`
}

func (service *AddUserPredictionService) Add(c *gin.Context) (r serializer.Response, err error) {
	u, _ := c.Get("user")

	user := model.User{}

	if u != nil {
		user = u.(model.User)
	}

	deviceInfo, err := util.GetDeviceInfo(c)
	if err != nil {
		return
	}

	if user.ID != 0 {
		var count int64
		count, err = model.GetUserPredictionCount(user.LastLoginDeviceUuid)

		if err != nil {
			r = serializer.DBErr(c, service, "", err)
			return
		}

		if count >= 3 {
			r = serializer.GeneralErr(c, errors.New("exceed limit"))
			return
		}

		err = model.CreateUserPrediction(user.ID, user.LastLoginDeviceUuid, service.PredictionId)

		return
	} else {
		var count int64
		count, err = model.GetUserPredictionCount(deviceInfo.Uuid)

		if err != nil {
			r = serializer.DBErr(c, service, "", err)
			return
		}

		if count >= 1 {
			r = serializer.GeneralErr(c, errors.New("exceed limit"))
			return
		}

		err = model.CreateUserPrediction(user.ID, deviceInfo.Uuid, service.PredictionId)

		if err != nil {
			r = serializer.GeneralErr(c, err)
			return
		}

		return
	}
}

var dummyJson = `{
   "prediction_id": 8,
   "analyst_id": 8,
   "prediction_title": "Title 8",
   "prediction_desc": "Description for prediction 8",
   "is_locked": false,
   "created_at": "2024-07-31T15:26:14.665498Z",
   "view_count": 800,
   "status": 0,
   "analyst_detail": {
     "analyst_id": 8,
     "analyst_name": "专家4",
     "analyst_desc": "并崇业20年，经验丰富",
     "analyst_source": {
       "source_name": "",
       "source_icon": "/aha-img/cash_method_icon/14/14-cash_method_icon-20240502091221-CXdgbq.png"
     },
     "analyst_image": "https://cdn.tayalive.com/aha-img/user/default_user_image/102.jpg",
     "winning_streak": 20,
     "accuracy": 99,
     "num_followers": 0,
     "total_predictions": 0,
     "predictions": [],
     "recent_total": 0,
     "recent_wins": 0
   },
   "selection_list": [
      {
         "mg": [
            {
               "mty": 1000,
               "pe": 1001,
               "nm": "让球",
               "mks": [
                  {
                     "op": [
                        {
                           "na": "毕尔巴鄂竞技",
                           "nm": "-1",
                           "ty": 1,
                           "od": 1.82,
                           "bod": 1.82,
                           "odt": 1,
                           "li": "-1",
                           "selected": true
                        },
                        {
                           "na": "赫塔菲",
                           "nm": "+1",
                           "ty": 2,
                           "od": 2.11,
                           "bod": 2.11,
                           "odt": 1,
                           "li": "+1",
                           "selected": false
                        }
                     ],
                     "id": 6442750,
                     "ss": 1,
                     "au": 1,
                     "mbl": 1,
                     "li": "-1"
                  }
               ]
            }
         ],
         "lg": {
            "na": "AC西班牙甲级联赛",
            "id": 10815,
            "or": 3,
            "lurl": "https://newsports-static-image.s3.ap-northeast-1.amazonaws.com/data/9aa1019b6854acf872ccf842634cf37f.png",
            "sid": 1,
            "rid": 74,
            "rnm": "西班牙",
            "rlg": "https://newsports-static-image.s3.ap-northeast-1.amazonaws.com/data/4b5a8fc852245e041e6d1aff973f63cd.png",
            "hot": true,
            "slid": 108150000
         },
         "ts": [
            {
               "na": "毕尔巴鄂竞技",
               "id": 53044,
               "lurl": "https://newsports-static-image.s3.ap-northeast-1.amazonaws.com/data/9e5e634734aad3a739c014b96dc03537.png"
            },
            {
               "na": "赫塔菲",
               "id": 54729,
               "lurl": "https://newsports-static-image.s3.ap-northeast-1.amazonaws.com/data/8cc5e3a0f86d863bb0c5ad3dc54d7344.png"
            }
         ],
         "id": 1007696,
         "bt": 1723741200000
      },
      {
         "mg": [
            {
               "mty": 1005,
               "pe": 1001,
               "nm": "独赢",
               "mks": [
                  {
                     "op": [
                        {
                           "na": "毕尔巴鄂竞技",
                           "nm": "主",
                           "ty": 1,
                           "od": 1.49,
                           "bod": 1.49,
                           "odt": 1,
                           "selected": true
                        },
                        {
                           "na": "和",
                           "nm": "和",
                           "ty": 3,
                           "od": 4.23,
                           "bod": 4.23,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "赫塔菲",
                           "nm": "客",
                           "ty": 2,
                           "od": 6.95,
                           "bod": 6.95,
                           "odt": 1,
                           "selected": false
                        }
                     ],
                     "id": 6438796,
                     "ss": 1,
                     "au": 1
                  }
               ]
            }
         ],
         "lg": {
            "na": "AC西班牙甲级联赛",
            "id": 10815,
            "or": 3,
            "lurl": "https://newsports-static-image.s3.ap-northeast-1.amazonaws.com/data/9aa1019b6854acf872ccf842634cf37f.png",
            "sid": 1,
            "rid": 74,
            "rnm": "西班牙",
            "rlg": "https://newsports-static-image.s3.ap-northeast-1.amazonaws.com/data/4b5a8fc852245e041e6d1aff973f63cd.png",
            "hot": true,
            "slid": 108150000
         },
         "ts": [
            {
               "na": "毕尔巴鄂竞技",
               "id": 53044,
               "lurl": "https://newsports-static-image.s3.ap-northeast-1.amazonaws.com/data/9e5e634734aad3a739c014b96dc03537.png"
            },
            {
               "na": "赫塔菲",
               "id": 54729,
               "lurl": "https://newsports-static-image.s3.ap-northeast-1.amazonaws.com/data/8cc5e3a0f86d863bb0c5ad3dc54d7344.png"
            }
         ],
         "id": 1007696,
         "bt": 1723741200000
      },
      {
         "mg": [
            {
               "mty": 1099,
               "pe": 1001,
               "nm": "波胆",
               "mks": [
                  {
                     "op": [
                        {
                           "na": "1-0",
                           "nm": "1-0",
                           "ty": 110,
                           "od": 6.8,
                           "bod": 6.8,
                           "odt": 1,
                           "selected": true
                        },
                        {
                           "na": "2-0",
                           "nm": "2-0",
                           "ty": 120,
                           "od": 7.2,
                           "bod": 7.2,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "2-1",
                           "nm": "2-1",
                           "ty": 121,
                           "od": 11,
                           "bod": 11,
                           "odt": 1,
                           "selected": true
                        },
                        {
                           "na": "3-0",
                           "nm": "3-0",
                           "ty": 130,
                           "od": 11,
                           "bod": 11,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "3-1",
                           "nm": "3-1",
                           "ty": 131,
                           "od": 17,
                           "bod": 17,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "3-2",
                           "nm": "3-2",
                           "ty": 132,
                           "od": 50,
                           "bod": 50,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "4-0",
                           "nm": "4-0",
                           "ty": 140,
                           "od": 25,
                           "bod": 25,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "4-1",
                           "nm": "4-1",
                           "ty": 141,
                           "od": 35,
                           "bod": 35,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "4-2",
                           "nm": "4-2",
                           "ty": 142,
                           "od": 101,
                           "bod": 101,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "4-3",
                           "nm": "4-3",
                           "ty": 143,
                           "od": 101,
                           "bod": 101,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "0-0",
                           "nm": "0-0",
                           "ty": 100,
                           "od": 12,
                           "bod": 12,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "1-1",
                           "nm": "1-1",
                           "ty": 111,
                           "od": 9.8,
                           "bod": 9.8,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "2-2",
                           "nm": "2-2",
                           "ty": 122,
                           "od": 30,
                           "bod": 30,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "3-3",
                           "nm": "3-3",
                           "ty": 133,
                           "od": 101,
                           "bod": 101,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "4-4",
                           "nm": "4-4",
                           "ty": 144,
                           "od": 101,
                           "bod": 101,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "0-1",
                           "nm": "0-1",
                           "ty": 101,
                           "od": 19,
                           "bod": 19,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "0-2",
                           "nm": "0-2",
                           "ty": 102,
                           "od": 60,
                           "bod": 60,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "1-2",
                           "nm": "1-2",
                           "ty": 112,
                           "od": 30,
                           "bod": 30,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "0-3",
                           "nm": "0-3",
                           "ty": 103,
                           "od": 101,
                           "bod": 101,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "1-3",
                           "nm": "1-3",
                           "ty": 113,
                           "od": 101,
                           "bod": 101,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "2-3",
                           "nm": "2-3",
                           "ty": 123,
                           "od": 101,
                           "bod": 101,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "0-4",
                           "nm": "0-4",
                           "ty": 104,
                           "od": 101,
                           "bod": 101,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "1-4",
                           "nm": "1-4",
                           "ty": 114,
                           "od": 101,
                           "bod": 101,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "2-4",
                           "nm": "2-4",
                           "ty": 124,
                           "od": 101,
                           "bod": 101,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "3-4",
                           "nm": "3-4",
                           "ty": 134,
                           "od": 101,
                           "bod": 101,
                           "odt": 1,
                           "selected": false
                        },
                        {
                           "na": "其他",
                           "nm": "其他",
                           "ty": 244,
                           "od": 24,
                           "bod": 24,
                           "odt": 1,
                           "selected": false
                        }
                     ],
                     "id": 6442644,
                     "ss": 1,
                     "au": 1
                  }
               ]
            }
         ],
         "lg": {
            "na": "AC西班牙甲级联赛",
            "id": 10815,
            "or": 3,
            "lurl": "https://newsports-static-image.s3.ap-northeast-1.amazonaws.com/data/9aa1019b6854acf872ccf842634cf37f.png",
            "sid": 1,
            "rid": 74,
            "rnm": "西班牙",
            "rlg": "https://newsports-static-image.s3.ap-northeast-1.amazonaws.com/data/4b5a8fc852245e041e6d1aff973f63cd.png",
            "hot": true,
            "slid": 108150000
         },
         "ts": [
            {
               "na": "毕尔巴鄂竞技",
               "id": 53044,
               "lurl": "https://newsports-static-image.s3.ap-northeast-1.amazonaws.com/data/9e5e634734aad3a739c014b96dc03537.png"
            },
            {
               "na": "赫塔菲",
               "id": 54729,
               "lurl": "https://newsports-static-image.s3.ap-northeast-1.amazonaws.com/data/8cc5e3a0f86d863bb0c5ad3dc54d7344.png"
            }
         ],
         "id": 1007696,
         "bt": 1723741200000
      }
   ],
   "sport_id": 0
 }
 `
