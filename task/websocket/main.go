package websocket

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"os"
	"strings"
	"time"
	modelWebsocket "web-api/model/websocket"
	"web-api/util"
)

var Conn modelWebsocket.Connection
var Functions []func(ctx context.Context, cancelFunc context.CancelFunc)

func Connect(retryInterval int64) {
	for {
		ctx := connect()
		<-ctx.Done()
		util.Log().Info(fmt.Sprintf(`retrying in %d seconds...`, retryInterval))
		time.Sleep(time.Duration(retryInterval) * time.Second)
	}
}

func connect() (ctx context.Context) {
	ctx, cancelFunc := context.WithCancel(context.TODO())
	util.Log().Info(`connecting to ` + os.Getenv("WS_URL"))
	c, _, err := websocket.DefaultDialer.Dial(os.Getenv("WS_URL"), nil)
	if err != nil {
		util.Log().Error("ws connection error", err)
		cancelFunc()
		return
	}
	util.Log().Info(`connected`)
	Conn.Socket = c

	go func() {
		for {
			_, msg, err := Conn.Socket.ReadMessage()
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
	Conn.Send(`40{"token":"` + os.Getenv("WS_TOKEN") + `"}`)
	return
}
