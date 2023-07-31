package fb_api

import (
	"blgit.rfdev.tech/taya/game-service/fb"
	"github.com/gin-gonic/gin"
	"web-api/api"
	"web-api/model"
	"web-api/serializer"
)

func GetToken(c *gin.Context) {
	u, _ := c.Get("user")
	user := u.(model.User)

	client := fb.FB{
		MerchantId:        "1552945083054354433",
		MerchantApiSecret: "Lc63hMKwQz0R8Y4MbB7F6mhCbzLuZoU9",
		IsTesting:         true,
	}
	res, err := client.GetToken(user.Username, "pc", "")
	if err != nil {
		c.JSON(500, api.ErrorResponse(err))
	} else {
		c.JSON(200, serializer.Response{
			Data: res,
		})
	}
}
