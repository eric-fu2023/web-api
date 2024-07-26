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

	deviceInfo, err := util.GetDeviceInfo(c)
	if err != nil {
		r = serializer.DBErr(c, service, i18n.T("general_error"), err)
		return
	}

	fmt.Printf("Getting data with device id %s\n", deviceInfo.Uuid)

	strategy, err := model.GetUserPrediction(model.GetUserPredictionCond{DeviceId: deviceInfo.Uuid, UserId: 0})
	if err != nil {
		r = serializer.DBErr(c, service, i18n.T("general_error"), err)
		return
	}

	predictionIds := make([]int64, len(strategy))
	for i, pred := range strategy {
		predictionIds[i] = pred.PredictionId
	}

	if len(strategy) == 0 {

		// TODO : get random prediction
		model.CreateUserPrediction(0, deviceInfo.Uuid, 99)
		predictionIds = append(predictionIds, 99)

		return serializer.Response{
			Msg:  i18n.T("success"),
			Data: predictionIds,
		}, nil
	}

	return serializer.Response{
		Msg:  i18n.T("success"),
		Data: predictionIds,
	}, nil
}
