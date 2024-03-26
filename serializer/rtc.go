package serializer

import "web-api/model"

type RtcToken struct {
	Token  string `json:"token"`
	Expiry int64  `json:"expiry"`
}

func BuildRtcToken(a model.RtcToken) (b RtcToken) {
	b = RtcToken{
		Token:  a.Token,
		Expiry: a.Expiry,
	}
	return
}

type RtcTokenWithStreamerId struct {
	StreamerId int64  `json:"streamer_id"`
	Token      string `json:"token"`
}

type RtcTokens struct {
	List   []RtcTokenWithStreamerId `json:"list"`
	Expiry int64                    `json:"expiry"`
}

func BuildRtcTokens(a []model.RtcTokenWithStreamerId, expiry int64) (b RtcTokens) {
	b = RtcTokens{
		Expiry: expiry,
	}
	if len(a) > 0 {
		var list []RtcTokenWithStreamerId
		for _, aa := range a {
			list = append(list, RtcTokenWithStreamerId{
				StreamerId: aa.StreamerId,
				Token:      aa.Token,
			})
		}
		b.List = list
	}
	return
}
