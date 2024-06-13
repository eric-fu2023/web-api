package api

import (
	"github.com/gin-gonic/gin"
	"web-api/service/ninewicket"
)

func GetBalanceNine(c *gin.Context) {
	println("Hello")
	var service ninewicket.NineWicket

	if err := c.ShouldBind(&service); err == nil {
		//res, _ := service.GetGameBalance("2517")
		//res, _ := service.TransferFrom()
		res, _ := service.TransferTo()
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
