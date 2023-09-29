package websocket

import (
	"github.com/gorilla/websocket"
	"sync"
)

type Connection struct {
	Socket     *websocket.Conn
	writeMutex sync.Mutex
}

func (c *Connection) Send(message string) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()
	return c.Socket.WriteMessage(websocket.TextMessage, []byte(message))
}
