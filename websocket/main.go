package websocket

import (
	"context"
	"fmt"
	"time"
	"web-api/util"
)

var Conn Connection
var Functions []func(ctx context.Context, cancelFunc context.CancelFunc)

func Connect(retryInterval int64) {
	for {
		Conn = Connection{}
		ctx := Conn.Connect()
		<-ctx.Done()
		util.Log().Info(fmt.Sprintf(`retrying in %d seconds...`, retryInterval))
		time.Sleep(time.Duration(retryInterval) * time.Second)
	}
}
