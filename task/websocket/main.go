package websocket

import (
	"context"
	"fmt"
	"os"
	"time"
	"web-api/util"
	"web-api/websocket"
)

var Conn websocket.Connection
var Functions []func(conn *websocket.Connection, ctx context.Context, cancelFunc context.CancelFunc)

func Connect(retryInterval int64) {
	for {
		ctx := Conn.Connect(os.Getenv("WS_URL"), os.Getenv("WS_TOKEN"), Functions)
		<-ctx.Done()
		util.Log().Info(fmt.Sprintf(`retrying in %d seconds...`, retryInterval))
		time.Sleep(time.Duration(retryInterval) * time.Second)
	}
}
