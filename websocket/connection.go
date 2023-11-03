package websocket

import (
	"context"
	"github.com/gorilla/websocket"
	"os"
	"strings"
	"sync"
	"web-api/util"
)

type Connection struct {
	Socket     *websocket.Conn
	writeMutex sync.Mutex
}

func (conn *Connection) Send(message string) error {
	conn.writeMutex.Lock()
	defer conn.writeMutex.Unlock()
	return conn.Socket.WriteMessage(websocket.TextMessage, []byte(message))
}

func (conn *Connection) Connect() (ctx context.Context) {
	ctx, cancelFunc := context.WithCancel(context.TODO())
	util.Log().Info(`connecting to ` + os.Getenv("WS_URL"))
	c, _, err := websocket.DefaultDialer.Dial(os.Getenv("WS_URL"), nil)
	if err != nil {
		util.Log().Error("ws connection error", err)
		cancelFunc()
		return
	}
	util.Log().Info(`connected`)
	conn.Socket = c

	go func() {
		for {
			_, msg, err := conn.Socket.ReadMessage()
			if err != nil {
				util.Log().Error("ws read error", err)
				cancelFunc()
				return
			}
			message := string(msg)
			if strings.Contains(message, "socket_id") && strings.Contains(message, "welcome") {
				for _, f := range Functions {
					go f(ctx, cancelFunc)
				}
				return
			}
		}
	}()
	conn.Send(`40{"token":"` + os.Getenv("WS_TOKEN") + `"}`)
	return
}
