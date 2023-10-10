package websocket

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
	"web-api/conf/consts"
	"web-api/model"
	"web-api/model/websocket"
	"web-api/util"
)

func Reply(ctx context.Context, cancelFunc context.CancelFunc) {
	Conn.Send(`42["join_admin", {"room":"administration"}]`)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, msg, err := Conn.Socket.ReadMessage()
			if err != nil {
				util.Log().Error("ws read error", err)
				cancelFunc()
				return
			}
			message := string(msg)
			if message == "2" { // reply pong to ping from server
				Conn.Send(`3`)
			}
			if strings.Contains(message, "socket_id") {
				switch {
				case strings.Contains(message, `"room_join"`):
					go welcomeToRoom(message)
				}
			}
		}
	}
}

func welcomeToRoom(message string) {
	str := strings.Replace(message, `42["room_join",`, "", 1)
	str = strings.Replace(str, `"}]`, `"}`, 1)
	var j map[string]interface{}
	if e := json.Unmarshal([]byte(str), &j); e == nil {
		if rm, exists := j["room"]; exists {
			room := rm.(string)
			if v, exists := j["rejoin"]; !exists || !v.(bool) {
				for _, m := range consts.ChatSystem["messages"] {
					msg := websocket.RoomMessage{
						SocketId:  j["socket_id"].(string),
						Room:      room,
						Timestamp: time.Now().Unix(),
						Message:   m,
						UserId:    consts.ChatSystemId,
						UserType:  consts.ChatUserType["system"],
						Nickname:  consts.ChatSystem["names"][0],
						Type:      consts.WebSocketMessageType["text"],
					}
					msg.Send(&Conn)
				}

				coll := model.MongoDB.Collection("room_message")
				filter := bson.M{"room": room}
				opts := options.Find()
				opts.SetLimit(20)
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
					}
					msg1.Send(&Conn)
				}
			}
		}
	}
}
