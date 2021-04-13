package web

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// AllSockets runs `f` on all connected websockets, disconnecting any websockets for which `f` returns an non-nil error
// It is used as the Manager's `evtFunc`
func (s *server) AllSockets(f func(c *websocket.Conn) error) {
	s.connectedSocketsLock.Lock()
	defer s.connectedSocketsLock.Unlock()

	for c, cc := range s.connectedSockets {
		err := f(c)
		if err != nil {
			// If we cannot write to a socket, we disconnect it (the connection was broken anyways)
			c.Close()
			close(cc)
			delete(s.connectedSockets, c)
		}
	}
}

// HandleWebsocket connects/upgrades a websocket request. It runs until the websocket is disconnected.
func (s *server) HandleWebsocket(w http.ResponseWriter, r *http.Request) (err error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil // Upgrader has already responded to the request
	}

	if s.m.IsWorking() {
		err = conn.WriteJSON(map[string]interface{}{
			"type": "progress-start",
		})
		if err != nil {
			return conn.Close()
		}
	}

	closeChan := make(chan struct{})

	s.connectedSocketsLock.Lock()
	s.connectedSockets[conn] = closeChan
	s.connectedSocketsLock.Unlock()

	<-closeChan
	return nil
}
