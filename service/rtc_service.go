package service

import (
	"fmt"
	"github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/rtctokenbuilder2"
	"github.com/gin-gonic/gin"
	"time"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
)

type RtcTokenService struct {
	Uuid string `form:"uuid" json:"uuid"`
}

func (service *RtcTokenService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var userAccount string
	var user model.User
	u, isUser := c.Get("user")
	if isUser {
		user = u.(model.User)
		userAccount = fmt.Sprintf(`user:%d`, user.ID)
	} else {
		if service.Uuid == "" {
			r = serializer.ParamErr(c, service, i18n.T("uuid_error"), err)
			return
		}
		userAccount = fmt.Sprintf(`guest:%s`, service.Uuid)
	}
	rtcToken, err := model.ShengWang.GetTokenWithoutChannel(userAccount, rtctokenbuilder2.RoleSubscriber)
	if err != nil {
		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
		return
	}
	r = serializer.Response{
		Data: serializer.BuildRtcToken(rtcToken),
	}
	return
}

type RtcTokensService struct {
	Uuid       string `form:"uuid" json:"uuid"`
	StreamerId int64  `form:"streamer_id" json:"streamer_id"`
}

func (service *RtcTokensService) Get(c *gin.Context) (r serializer.Response, err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	var userAccount string
	var user model.User
	u, isUser := c.Get("user")
	if isUser {
		user = u.(model.User)
		userAccount = fmt.Sprintf(`user:%d`, user.ID)
	} else {
		if service.Uuid == "" {
			r = serializer.ParamErr(c, service, i18n.T("uuid_error"), err)
			return
		}
		userAccount = fmt.Sprintf(`guest:%s`, service.Uuid)
	}
	var streams []model.Stream
	var tokens []model.RtcTokenWithStreamerId
	q := model.DB.Model(model.Stream{}).Scopes(model.StreamsOnlineSorted("", "", false))
	if service.StreamerId != 0 {
		q = q.Where(`users.id`, service.StreamerId)
	}
	if err = q.Find(&streams).Error; err != nil {
		return
	}
	var duration int64 = 24 * 60 * 60
	expiry := time.Now().Add(time.Duration(duration) * time.Second).Unix()
	for _, stream := range streams {
		if rtcToken, e := model.ShengWang.GetTokenWithChannelAndDuration(fmt.Sprintf(`%d`, stream.StreamerId), userAccount, rtctokenbuilder2.RoleSubscriber, duration); e == nil {
			tokens = append(tokens, model.RtcTokenWithStreamerId{
				StreamerId: stream.StreamerId,
				Token:      rtcToken.Token,
			})
		}
	}

	r = serializer.Response{
		Data: serializer.BuildRtcTokens(tokens, expiry),
	}
	return
}
