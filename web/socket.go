package web

import (
	"net/http"
	"sync"
	"xarantolus/sensibleHub/store"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	connectedSockets = make(map[*websocket.Conn]chan struct{})
	csl              sync.Mutex
)

func AllSockets(f func(c *websocket.Conn) error) {
	csl.Lock()
	defer csl.Unlock()

	for c, cc := range connectedSockets {
		err := f(c)
		if err != nil {
			c.Close()
			close(cc)
			delete(connectedSockets, c)
		}
	}
}

func HandleWebsocket(w http.ResponseWriter, r *http.Request) (err error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil // Upgrader has already responded to the request
	}

	closeChan := make(chan struct{})

	if store.M.IsWorking() {
		conn.WriteJSON(map[string]interface{}{
			"type": "progress-start",
		})
	}

	csl.Lock()
	connectedSockets[conn] = closeChan
	csl.Unlock()

	<-closeChan
	return nil
}
