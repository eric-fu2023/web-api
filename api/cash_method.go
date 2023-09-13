package api

import (
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func TopupMethodList(c *gin.Context) {
	var service = service.CasheMethodListService{
		TopupOnly: true,
	}
	res, _ := service.List(c)
	c.JSON(200, res)
}

func WithdrawMethodList(c *gin.Context) {
	var service = service.CasheMethodListService{
		WithdrawOnly: true,
	}
	res, _ := service.List(c)
	c.JSON(200, res)
}
