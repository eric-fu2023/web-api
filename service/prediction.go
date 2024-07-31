package service

import (
	"web-api/model"
	"web-api/serializer"
	"web-api/service/common"
	"web-api/util/i18n"

	"github.com/gin-gonic/gin"
)

type PredictionService struct {
	common.Page
	AnalystId int64 `json:"analyst_id" form:"analyst_id"`
}

// func (service *StrategyService) List(c *gin.Context) (serializer.Response, error) {

// 	return serializer.Response{}, nil
// }

func (service *PredictionService) List(c *gin.Context) (r serializer.Response, err error) {
	/*
		not logged in
		logged in
		logged in paid
	*/

	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")

	user := model.User{}

	if u != nil {
		user = u.(model.User)
	}

	hasAuth := user.ID != 0

	// deviceInfo, err := util.GetDeviceInfo(c)
	if err != nil {
		r = serializer.DBErr(c, service, i18n.T("general_error"), err)
		return
	}

	if hasAuth {
		u, _ := c.Get("user")
		user := u.(model.User)

		hasPaymentToday := true
		_, err = model.HasTopupToday(c, user.ID)

		if err != nil {
			return serializer.DBErr(c, service, i18n.T("general_error"), err), err
		}

		if hasPaymentToday {

			// predictionRepo := repo.NewMockPredictionRepo()
			// r, err = predictionRepo.GetList(c)
			// if err != nil {
			// 	r = serializer.DBErr(c, service, i18n.T("general_error"), err)
			// 	return
			// }

			// return

			predictions, err := model.MockGetUserPrediction(service.Limit, service.Page.Page, -1, service.AnalystId)
			if err != nil {
				r = serializer.DBErr(c, service, i18n.T("general_error"), err)
				return r, err
			}

			return serializer.Response{
				Msg:  i18n.T("success"),
				Data: serializer.BuildPredictions(predictions),
			}, nil

		} else {
			predictions, err := model.MockGetUserPrediction(service.Limit, service.Page.Page, 3, service.AnalystId)
			if err != nil {
				r = serializer.DBErr(c, service, i18n.T("general_error"), err)
				return r, err
			}

			return serializer.Response{
				Msg:  i18n.T("success"),
				Data: serializer.BuildPredictions(predictions),
			}, nil
		}

	} else {
		// no log in, query with device id and user id 0
		predictions, err := model.MockGetUserPrediction(service.Limit, service.Page.Page, 1, service.AnalystId)
		if err != nil {
			r = serializer.DBErr(c, service, i18n.T("general_error"), err)
			return r, err
		}

		return serializer.Response{
			Msg:  i18n.T("success"),
			Data: serializer.BuildPredictions(predictions),
		}, nil
	}

}

func mockGetRandomPredictions(length int64) []int64 {
	if length == 1 {
		return []int64{99}
	} else if length == 2 {
		return []int64{150, 151}
	} else if length == 3 {
		return []int64{881, 882, 883}
	} else {
		return []int64{}
	}
}
