package web

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	connectedSockets = make(map[*websocket.Conn]chan struct{})
	csl              sync.RWMutex
)

func AllSockets(f func(c *websocket.Conn) error) {
	csl.RLock()
	defer csl.RUnlock()

	for c, cc := range connectedSockets {
		err := f(c)
		if err != nil {
			c.Close()

			csl.RUnlock()
			csl.Lock()

			close(cc)
			delete(connectedSockets, c)

			csl.Unlock()
			csl.RLock()
		}
	}
}

func HandleWebsocket(w http.ResponseWriter, r *http.Request) (err error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil // Upgrader has already responded to the request
	}

	closeChan := make(chan struct{})

	csl.Lock()
	connectedSockets[conn] = closeChan
	csl.Unlock()

	<-closeChan
	return nil
}
