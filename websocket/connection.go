package websocket

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
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

func (conn *Connection) Connect(url string, token string, functions []func(conn *Connection, ctx context.Context, cancelFunc context.CancelFunc)) (ctx context.Context) {
	ctx, cancelFunc := context.WithCancel(context.TODO())
	util.Log().Info(`connecting to ` + url)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
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
				for _, f := range functions {
					go f(conn, ctx, cancelFunc)
				}
				return
			}
		}
	}()
	conn.Send(fmt.Sprintf(`40{"token":"%s"}`, token))
	return
}
