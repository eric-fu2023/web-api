package websocket

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"strings"
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/serializer"
	"web-api/util"
	"web-api/util/i18n"
	"web-api/websocket"
)

func Reply(conn *websocket.Connection, ctx context.Context, cancelFunc context.CancelFunc) {
	conn.Send(`42["join_admin", {"room":"administration"}]`)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, msg, err := conn.Socket.ReadMessage()
			if err != nil {
				util.Log().Error("ws read error", err)
				cancelFunc()
				return
			}
			message := string(msg)
			if message == "2" { // reply pong to ping from server
				conn.Send(`3`)
			}
			if strings.Contains(message, "socket_id") {
				switch {
				case strings.Contains(message, `"room_join"`):
					go welcomeToRoom(conn, message)
				}
			}
		}
	}
}

func welcomeToRoom(conn *websocket.Connection, message string) {
	str := strings.Replace(message, `42["room_join",`, "", 1)
	str = strings.Replace(str, `"}]`, `"}`, 1)
	var j map[string]interface{}
	if e := json.Unmarshal([]byte(str), &j); e == nil {
		if rm, exists := j["room"]; exists {
			room := rm.(string)
			lang := "zh"
			if v, ok := j["locale"]; ok {
				lang = v.(string)
				lang = lang[0:2]
			}
			i18n := i18n.I18n{}
			i18n.LoadLanguages(lang)
			m := i18n.T("chat_welcome_message")
			if os.Getenv("CHAT_WELCOME_MESSAGES") != "" {
				m = os.Getenv("CHAT_WELCOME_MESSAGES")
			}
			n := i18n.T("chat_welcome_name")
			if os.Getenv("CHAT_WELCOME_NAMES") != "" {
				n = os.Getenv("CHAT_WELCOME_NAMES")
			}
			if v, exists := j["rejoin"]; !exists || !v.(bool) {
				msg := websocket.RoomMessage{
					SocketId:  j["socket_id"].(string),
					Room:      room,
					Timestamp: time.Now().Unix(),
					Message:   m,
					UserId:    consts.ChatSystemId,
					UserType:  consts.ChatUserType["system"],
					Nickname:  n,
					Avatar:    serializer.Url(os.Getenv("CHAT_SYSTEM_PROFILE_IMG")),
					Type:      consts.WebSocketMessageType["text"],
				}
				msg.Send(conn)

				coll := model.MongoDB.Collection("room_message")
				filter := bson.M{"room": room, "deleted_at": nil}
				opts := options.Find()
				opts.SetLimit(50)
				opts.SetSort(bson.D{{"timestamp", -1}, {"_id", -1}})
				ctx := context.TODO()
				cursor, err := coll.Find(ctx, filter, opts)
				if err != nil {
					return
				}
				var ms []model.RoomMessage
				for cursor.Next(ctx) {
					var pm model.RoomMessage
					cursor.Decode(&pm)
					ms = append(ms, pm)
				}
				for i := len(ms) - 1; i >= 0; i-- {
					msg1 := websocket.RoomMessage{
						Id:        ms[i].Id,
						SocketId:  j["socket_id"].(string),
						Room:      room,
						Timestamp: ms[i].Timestamp,
						Message:   ms[i].Message,
						UserId:    ms[i].UserId,
						UserType:  ms[i].UserType,
						Nickname:  ms[i].Nickname,
						Avatar:    ms[i].Avatar,
						Type:      ms[i].Type,
						IsHistory: true,
						VipId:     ms[i].VipId,
					}
					msg1.Send(conn)
				}
			}
		}
	}
}
