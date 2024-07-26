package service

import (
	"fmt"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"

	"github.com/gin-gonic/gin"
)

type PredictionService struct {
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

	hasAuth := c.GetString("_token_string") != ""

	fmt.Println(hasAuth)

	deviceInfo, err := util.GetDeviceInfo(c)
	if err != nil {
		r = serializer.DBErr(c, service, i18n.T("general_error"), err)
		return
	}

	hasPaymentToday := true // TODO : query from db 

	fmt.Printf("Getting data with device id %s\n", deviceInfo.Uuid)

	if hasAuth {
		u, _ := c.Get("user")
		user := u.(model.User)		

		if hasPaymentToday {
			return serializer.Response{
				Msg:  i18n.T("success"),
				Data: "TODO: get all data :)",
				// TODO : ^^^^^
			}, nil

		} else {
			predictions, err := model.GetUserPrediction(model.GetUserPredictionCond{DeviceId: user.LastLoginDeviceUuid, UserId: user.ID})
			if err != nil {
				r = serializer.DBErr(c, service, i18n.T("general_error"), err)
				return r, err
			}
			
			return serializer.Response{
				Msg:  i18n.T("success"),
				Data: serializer.BuildUserPredictionsList(predictions),
			}, nil
		}

	} else {
		// no log in, query with device id and user id 0
		predictions, err := model.GetUserPrediction(model.GetUserPredictionCond{DeviceId: deviceInfo.Uuid, UserId: 0})
		if err != nil {
			r = serializer.DBErr(c, service, i18n.T("general_error"), err)
			return r, err
		}
		
		return serializer.Response{
			Msg:  i18n.T("success"),
			Data: serializer.BuildUserPredictionsList(predictions),
		}, nil
	}

	// if len(predictions) == 0 {

	// 	// TODO : get random prediction
	// 	model.CreateUserPrediction(0, deviceInfo.Uuid, 99)

	// 	return serializer.Response{
	// 		Msg:  i18n.T("success"),
	// 		Data: serializer.BuildUserPredictionsList(),
	// 	}, nil
	// }

	
}
