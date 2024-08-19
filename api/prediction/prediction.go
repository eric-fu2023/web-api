package api

import (
	"web-api/api"
	"web-api/service"

	"github.com/gin-gonic/gin"
)

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
