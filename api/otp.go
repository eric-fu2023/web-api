package api

import (
	"web-api/service"

	"github.com/gin-gonic/gin"
)

func VerifyOtp(c *gin.Context) {
	var service service.UserOtpVerificationService
	if err := c.ShouldBind(&service); err == nil {
		res := service.Verify(c)
		c.JSON(200, res)
	} else {
		c.JSON(400, ErrorResponse(c, service, err))
	}
}
