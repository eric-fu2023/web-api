package api

import (
	"web-api/serializer"
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func WthdrawAccountsList(c *gin.Context) {
	var service service.ListWithdrawAccountsService
	if err := c.ShouldBind(&service); err == nil {
		if res, err := service.List(c); err == nil {
			c.JSON(200, res)
		} else {
			c.JSON(200, serializer.EnsureErr(c, err, res))
		}
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}

func WthdrawAccountsAdd(c *gin.Context) {
	Do[service.AddWithdrawAccountService](c)
}

func WthdrawAccountsRemove(c *gin.Context) {
	Do[service.DeleteWithdrawAccountService](c)
}
