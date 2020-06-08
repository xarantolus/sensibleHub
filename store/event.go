package store

import "github.com/gorilla/websocket"

// SetEventFunc sets the function that will be called with Manager events.
// Used for websocket notifications
func (m *Manager) SetEventFunc(f func(f func(c *websocket.Conn) error)) {
	m.evtFunc = f
}

func (m *Manager) event(evtType string, evtData interface{}) {
	if m.evtFunc == nil {
		return
	}

	go m.evtFunc(func(c *websocket.Conn) error {
		return c.WriteJSON(map[string]interface{}{
			"type": evtType,
			"data": evtData,
		})
	})
}
