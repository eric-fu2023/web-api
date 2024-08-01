package api

import (
	"web-api/api"
	"web-api/service"

	"github.com/gin-gonic/gin"
)

// func ListStrategy(c *gin.Context) {
// 	/*
// 		get user from context
// 		get deviceId from context

// 		if deviceId = nil
// 			return error

// 		paid = getUserPaidToday

// 		if (not logged in )
// 			query with device id
// 		else if login && not paid
// 			query with userId and deviceId
// 		else
// 			query all

// 	*/
// }

func ListPredictions(c *gin.Context) {
	var service service.PredictionListService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.List(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func AddUserPrediction(c *gin.Context) {
	var service service.AddUserPredictionService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.Add(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}

func GetPredictionDetail(c *gin.Context) {
	var service service.PredictionDetailService
	if err := c.ShouldBind(&service); err == nil {
		res, e := service.GetDetail(c)
		c.JSON(200, res)
		if e != nil {
			c.Abort()
		}
	} else {
		c.JSON(400, api.ErrorResponse(c, service, err))
	}
}