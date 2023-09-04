package fb

import (
	"github.com/gin-gonic/gin"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/service"
	"web-api/util"
	"web-api/util/i18n"
)

type TokenService struct {
	service.Platform
}

func (service *TokenService) Get(c *gin.Context) (res serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	u, _ := c.Get("user")
	user := u.(model.User)

	if user.Username == "" {
		res = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("finish_setup"), nil)
		return
	}

	client := util.FBFactory.NewClient()
	r, err := client.GetToken(user.Username, consts.PlatformIdToFbPlatformId[service.Platform.Platform], "")
	if err != nil {
		res = serializer.Err(c, service, serializer.CodeGeneralError, "", err)
		return
	}
	res = serializer.Response{
		Data: r,
	}
	return
}
