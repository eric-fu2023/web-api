package fb

import (
	"github.com/gin-gonic/gin"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
)

type TokenService struct {
	Platform int64 `form:"platform" json:"platform" binding:"required"`
}

func (service *TokenService) Get(c *gin.Context) (res serializer.Response, err error) {
	u, _ := c.Get("user")
	user := u.(model.User)

	client := util.FBFactory.NewClient()
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
