package web

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"xarantolus/sensibleHub/store"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	connectedSockets = make(map[*websocket.Conn]chan struct{})
	csl              sync.Mutex
)

// AllSockets runs `f` on all connected websockets, disconnecting any websockets for which `f` returns an non-nil error
// It is used as the Manager's `evtFunc`
func AllSockets(f func(c *websocket.Conn) error) {
	csl.Lock()
	defer csl.Unlock()

	for c, cc := range connectedSockets {
		err := f(c)
		if err != nil {
			// If we cannot write to a socket, we disconnect it (the connection was broken anyways)
			c.Close()
			close(cc)
			delete(connectedSockets, c)
		}
	}
}

// HandleWebsocket connects/upgrades a websocket request. It runs until the websocket is disconnected.
func HandleWebsocket(w http.ResponseWriter, r *http.Request) (err error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil // Upgrader has already responded to the request
	}

	if store.M.IsWorking() {
		err = conn.WriteJSON(map[string]interface{}{
			"type": "progress-start",
		})
		if err != nil {
			return conn.Close()
		}
	}

	closeChan := make(chan struct{})

	csl.Lock()
	connectedSockets[conn] = closeChan
	csl.Unlock()

	<-closeChan
	return nil
}
