package service

import (
	"errors"
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

		if hasPaymentToday {
			// logged in, has payment - show all
			predictions, err = model.ListPredictions(model.ListPredictionCond{Limit: service.Limit, Page: service.Page.Page, AnalystId: service.AnalystId, FbMatchId: service.FbMatchId})
			if err != nil {
				r = serializer.DBErr(c, service, i18n.T("general_error"), err)
				return
			}

			r.Msg = i18n.T("success")
			r.Data = serializer.BuildPredictionsList(predictions)
			return

		} else {
			// logged in, no payment - unlock 3 max
			predictions, err = model.ListPredictions(model.ListPredictionCond{Limit: service.Limit, Page: service.Page.Page, AnalystId: service.AnalystId, FbMatchId: service.FbMatchId})
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
		predictions, err = model.ListPredictions(model.ListPredictionCond{Limit: service.Limit, Page: service.Page.Page, AnalystId: service.AnalystId, FbMatchId: service.FbMatchId})
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
	PredictionId int64 	`json:"prediction_id" form:"prediction_id"`
}

func (service *PredictionDetailService) GetDetail(c *gin.Context) (r serializer.Response, err error) {
	data, err := model.GetPrediction(service.PredictionId)

	if err != nil {
		r = serializer.DBErr(c, service, "", err)
		return 
	}

	r.Data = serializer.BuildPrediction(data)

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

		return
	}
}
