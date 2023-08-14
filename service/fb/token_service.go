package fb

import (
	"blgit.rfdev.tech/taya/game-service/fb"
	"github.com/gin-gonic/gin"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
)

type TokenService struct {
	Platform int64 `form:"platform" json:"platform" binding:"required"`
}

func (service *TokenService) Get(c *gin.Context) (res serializer.Response, err error) {
	u, _ := c.Get("user")
	user := u.(model.User)

	client := fb.FB{
		MerchantId:        "1552945083054354433",
		MerchantApiSecret: "Lc63hMKwQz0R8Y4MbB7F6mhCbzLuZoU9",
		IsSandbox:         true,
	}
	r, err := client.GetToken(user.Username, consts.PlatformIdToFbPlatformId[service.Platform], "")
	if err != nil {
		res = serializer.Err(c, service, serializer.CodeGeneralError, "", err)
		return
	}
	res = serializer.Response{
		Data: r,
	}
	return
}
