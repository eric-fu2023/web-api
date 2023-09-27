package ws

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"os"
	"strings"
	"sync"
	"time"
	"web-api/util"
)

type Connection struct {
	Socket     *websocket.Conn
	writeMutex sync.Mutex
	Ready      chan bool
	Closed     chan bool
}

func (c *Connection) Start() {
	close(c.Ready)
}

func (c *Connection) Close() {
	close(c.Ready)
	close(c.Closed)
}

func (c *Connection) Send(message string) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()
	return c.Socket.WriteMessage(websocket.TextMessage, []byte(message))
}

var Conn Connection

func InitWebsocket() {
	Conn.Ready = make(chan bool, 1)
	Conn.Closed = make(chan bool, 1)
	util.Log().Info(`connecting to ` + os.Getenv("WS_URL"))
	c, _, err := websocket.DefaultDialer.Dial(os.Getenv("WS_URL"), nil)
	if err != nil {
		util.Log().Panic("ws connection error", err)
		util.Log().Info(`retrying in 15 seconds...`)
		time.Sleep(15 * time.Second)
		return
	}
	Conn.Socket = c

	// sending token and receiving welcome message
	go func() {
		for {
			_, msg, err := Conn.Socket.ReadMessage()
			if err != nil {
				util.Log().Panic("ws read error", err)
				Conn.Close()
				return
			}
			message := string(msg)
			if strings.Contains(message, "socket_id") && strings.Contains(message, "welcome") {
				Conn.Start()
				return
			}
		}
	}()
	Conn.Send(`40{"token":"` + os.Getenv("WS_TOKEN") + `"}`)
}

type RoomMessage struct {
	SocketId  string `json:"socket_id"`
	Room      string `json:"room"`
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
	UserId    int64  `json:"user_id"`
	UserType  int64  `json:"user_type"`
	Nickname  string `json:"nickname"`
	Type      int64  `json:"type"`
	IsHistory bool   `json:"is_history"`
}

func (a RoomMessage) Send() (err error) {
	event := "room"
	if a.SocketId != "" {
		event = "room_socket"
	}
	if msg, err := json.Marshal(a); err == nil {
		if err = Conn.Send(fmt.Sprintf(`42["%s", %s]`, event, string(msg))); err != nil {
			util.Log().Error("ws send error", err)
		}
	}
	return
}
