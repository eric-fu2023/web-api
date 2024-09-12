package websocket

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/util"
	"web-api/util/i18n"
	"web-api/websocket"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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

			parts := strings.Split(room, ":")
			streamId := 0
			if len(parts) > 1 {
				streamId, _ = strconv.Atoi(parts[1])
			}
			streamDetail, err := model.GetStreamDetail(int64(streamId))
			if err != nil {
				return
			}

			if v, exists := j["rejoin"]; !exists || !v.(bool) {
				streamerWelcomeMsg := websocket.RoomMessage{
					SocketId:  j["socket_id"].(string),
					Room:      room,
					Timestamp: time.Now().Unix(),
					Message:   streamDetail.WelcomeMessage,
					UserId:    streamDetail.StreamerId,
					UserType:  consts.ChatUserType["streamer"],
					Nickname:  streamDetail.Streamer.Nickname,
					Avatar:    streamDetail.Streamer.Avatar,
					Type:      consts.WebSocketMessageType["text"],
				}
				streamerWelcomeMsg.Send(conn)

				if vv, ex := j["nickname"]; ex {
					msg2 := websocket.RoomMessage{
						Room:      room,
						Timestamp: time.Now().Unix(),
						Nickname:  vv.(string),
						Type:      consts.WebSocketMessageType["join"],
					}
					msg2.Send(conn)
				}

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
