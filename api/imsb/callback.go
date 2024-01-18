package imsb_api

import (
	"blgit.rfdev.tech/taya/game-service/imsb/callback"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
)

func GetBalance(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	fmt.Println(string(body))
	c.JSON(200, callback.BaseResponse{
		StatusCode: 100,
		StatusDesc: "success",
	})
	//var service imsb.TokenService
	//if err := c.ShouldBind(&service); err == nil {
	//	res, _ := service.Get(c)
	//	c.JSON(200, res)
	//} else {
	//	c.JSON(400, api.ErrorResponse(c, service, err))
	//}
}
