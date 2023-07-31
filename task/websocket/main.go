package websocket

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"os"
	"strings"
	"sync"
	"time"
	"web-api/conf/consts"
)

type Connection struct {
	Socket     *websocket.Conn
	writeMutex sync.Mutex
	Ready      chan bool
	Ended      chan bool
}

var Websocket Connection
var PauseBotsUntil map[string]time.Time

func (c *Connection) Send(message string) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()
	return c.Socket.WriteMessage(websocket.TextMessage, []byte(message))
}

type RoomMessage struct {
	Room      string `json:"room"`
	Message   string `json:"message"`
	UserId    int64  `json:"user_id"`
	UserType  int64  `json:"user_type"`
	Nickname  string `json:"nickname"`
	Timestamp int64  `json:"timestamp"`
	Type      int64  `json:"type"`
}

type RoomTargetedMessage struct {
	RoomMessage
	SocketId string `json:"socket_id"`
}

type PrivateMessage struct {
	Room      string `json:"room,omitempty"`
	Message   string `json:"message,omitempty"`
	UId       int64  `json:"uid,omitempty"`
	UserType  int64  `json:"user_type,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	Timestamp int64  `json:"timestamp"`
	Type      int64  `json:"type"`
	MatchId   int64  `json:"match_id,omitempty"`
}

type Sticker struct {
	RoomMessage
	Code string `json:"code"`
}

func (a RoomMessage) Send() (err error) {
	var msg []byte
	if msg, err = json.Marshal(a); err == nil {
		if err = Websocket.Send(`42["room", ` + string(msg) + `]`); err != nil {
			fmt.Println(err)
		}
	}
	return
}

func (a PrivateMessage) Send() (err error) {
	var msg []byte
	if msg, err = json.Marshal(a); err == nil {
		if err = Websocket.Send(`42["private", ` + string(msg) + `]`); err != nil {
			fmt.Println(err)
		}
	}
	return
}

func (a RoomTargetedMessage) Send() (err error) {
	var msg []byte
	if msg, err = json.Marshal(a); err == nil {
		if err = Websocket.Send(`42["room_socket", ` + string(msg) + `]`); err != nil {
			fmt.Println(err)
		}
	}
	return
}

func (a Sticker) Send() (err error) {
	var msg []byte
	if msg, err = json.Marshal(a); err == nil {
		if err = Websocket.Send(`42["room", ` + string(msg) + `]`); err != nil {
			fmt.Println(err)
		}
	}
	return
}

func SetupWebsocket() {
	Websocket.Ended = make(chan bool, 1)
	Websocket.Ready = make(chan bool, 1)

	fmt.Println(`connecting to ` + os.Getenv("WS_URL"))
	c, _, err := websocket.DefaultDialer.Dial(os.Getenv("WS_URL"), nil)
	if err != nil {
		fmt.Print("websocket err 1| ")
		fmt.Println(err)
		fmt.Println(`retrying in 15 seconds...`)
		time.Sleep(15 * time.Second)
		close(Websocket.Ended)
		return
	}
	Websocket.Socket = c

	// sending token and receiving welcome message
	go func() {
		for {
			_, msg, err := Websocket.Socket.ReadMessage()
			if err != nil {
				fmt.Print("websocket err 2| ")
				fmt.Println(err)
				close(Websocket.Ended)
				return
			}
			message := string(msg)
			if strings.Contains(message, "socket_id") && strings.Contains(message, "welcome") {
				close(Websocket.Ready)
				return
			}
		}
	}()
	Websocket.Send(`40{"token":"` + os.Getenv("WS_TOKEN") + `"}`)

	go func() {
		<-Websocket.Ready
		Websocket.Send(`42["join_admin", {"room":"administration"}]`)

		for {
			_, msg, err := Websocket.Socket.ReadMessage()
			if err != nil {
				fmt.Print("websocket err 3| ")
				fmt.Println(err)
				close(Websocket.Ended)
				return
			}
			//log.Println("↓ %s", msg)
			message := string(msg)
			if message == "2" { // reply pong to ping from server
				Websocket.Send(`3`)
			}
			if strings.Contains(message, "socket_id") {
				switch {
				case strings.Contains(message, `"room_join"`):
					go welcomeToRoom(message)
				}
			}
		}
	}()
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
					msg := RoomTargetedMessage{
						SocketId: j["socket_id"].(string),
					}
					msg.RoomMessage = RoomMessage{
						Room:      room,
						Message:   m,
						UserId:    consts.ChatSystemId,
						UserType:  consts.ChatUserType["system"],
						Nickname:  consts.ChatSystem["names"][0],
						Timestamp: time.Now().Unix(),
						Type:      consts.WebSocketMessageType["text"],
					}
					msg.Send()
				}
			}
		}
	}
}
