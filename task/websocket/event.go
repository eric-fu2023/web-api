package websocket

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util/i18n"
	"web-api/websocket"
)

func Event(conn *websocket.Connection, ctx context.Context, cancelFunc context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(5 * time.Minute)
			var streams []ploutos.LiveStream
			err := model.DB.Model(ploutos.LiveStream{}).Where(`status`, 2).Order(`id`).Find(&streams).Error
			if err != nil {
				continue
			}
			if len(streams) == 0 {
				continue
			}
			var events []ploutos.RoomEvent
			err = model.DB.Model(ploutos.RoomEvent{}).Where(`is_enabled`, true).Find(&events).Error
			if err != nil {
				continue
			}
			if len(events) == 0 {
				continue
			}
			eventLangs := map[string][]ploutos.RoomEvent{}
			for _, event := range events {
				eventLangs[event.Lang] = append(eventLangs[event.Lang], event)
			}
			for _, stream := range streams {
				for l, evnts := range eventLangs {
					r := rand.Intn(len(evnts))
					event := evnts[r]
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
						UserType: consts.ChatUserType["admin"],
						Nickname: n,
						Avatar:   serializer.Url(os.Getenv("CHAT_SYSTEM_PROFILE_IMG")),
						Type:     consts.WebSocketMessageType["text"],
					}
					if e := msg.Send(conn); e != nil {
						cancelFunc()
						return
					}
				}
			}
		}
	}
}
