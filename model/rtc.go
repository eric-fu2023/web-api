package model

import (
	rtctokenbuilder "github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/rtctokenbuilder2"
	"os"
	"time"
)

type Rtc struct {
	AppId       string
	Certificate string
}

type RtcToken struct {
	Token  string
	Expiry int64
}

type RtcTokenWithStreamerId struct {
	StreamerId int64
	Token      string
}

func (r Rtc) GetTokenWithoutChannel(userAccount string, role rtctokenbuilder.Role) (rtcToken RtcToken, err error) {
	rtcToken, err = r.GetToken(userAccount, "*", role, 0)
	return
}

func (r Rtc) GetTokenWithChannelAndDuration(channel string, userAccount string, role rtctokenbuilder.Role, duration int64) (rtcToken RtcToken, err error) {
	rtcToken, err = r.GetToken(userAccount, channel, role, duration)
	return
}

func (r Rtc) GetToken(userAccount string, channelName string, role rtctokenbuilder.Role, duration int64) (rtcToken RtcToken, err error) {
	if duration == 0 {
		duration = 24 * 60 * 60
	}
	rtcToken.Expiry = time.Now().Add(time.Duration(duration) * time.Second).Unix()
	rtcToken.Token, err = rtctokenbuilder.BuildTokenWithUserAccount(
		os.Getenv("SHENGWANG_APP_ID"),
		os.Getenv("SHENGWANG_CERTIFICATE"),
		channelName,
		userAccount,
		role,
		uint32(duration),
		uint32(duration))
	return
}
