package api

import (
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func TopupMethodList(c *gin.Context) {
	var service = service.CasheMethodListService{
		TopupOnly: true,
	}
	res, e := service.List(c)
	c.JSON(200, res)
	if e != nil {
		c.Abort()
	}
}

func WithdrawMethodList(c *gin.Context) {
	var service = service.CasheMethodListService{
		WithdrawOnly: true,
	}
	res, e := service.List(c)
	c.JSON(200, res)
	if e != nil {
		c.Abort()
	}
}
