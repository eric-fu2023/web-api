package websocket

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"math/rand"
	"os"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
	"web-api/websocket"
)

func Event(conn *websocket.Connection, ctx context.Context, cancelFunc context.CancelFunc) {
	c := cron.New(cron.WithSeconds())
	c.AddFunc("0 */5 * * * *", func() {
		sendEvent(conn, ctx, cancelFunc)
	})
	c.Start()
}

func sendEvent(conn *websocket.Connection, ctx context.Context, cancelFunc context.CancelFunc) {
	var streams []ploutos.LiveStream
	err := model.DB.Model(ploutos.LiveStream{}).Where(`status`, 2).Order(`id`).Find(&streams).Error
	if err != nil {
		return
	}
	if len(streams) == 0 {
		return
	}
	var events []ploutos.RoomEvent
	err = model.DB.Model(ploutos.RoomEvent{}).Where(`is_enabled`, true).Find(&events).Error
	if err != nil {
		return
	}
	if len(events) == 0 {
		return
	}
	eventLangs := map[string][]ploutos.RoomEvent{}
	for _, event := range events {
		eventLangs[event.Lang] = append(eventLangs[event.Lang], event)
	}
	for _, stream := range streams {
		for l, e := range eventLangs {
			r := rand.Intn(len(e))
			event := e[r]
			lang := "zh"
			if l != "" {
				lang = l
			}
			i18n := i18n.I18n{}
			i18n.LoadLanguages(lang)
			n := i18n.T("CHAT_WELCOME_NAME")
			msg := websocket.RoomMessage{
				Room:     fmt.Sprintf(`stream:%d.%v`, stream.ID, lang),
				Message:  event.Text,
				UserId:   consts.ChatSystemId,
				UserType: consts.ChatUserType["system"],
				Nickname: n,
				Avatar:   serializer.Url(os.Getenv("CHAT_SYSTEM_PROFILE_IMG")),
				Type:     consts.WebSocketMessageType["text"],
			}
			msg.Send(conn)
		}
	}
}
