package store

import "github.com/gorilla/websocket"

type EventType string

const (
	ETIdle EventType = "Idle"

	ETStartDownload    EventType = "StartDownload"
	ETFinishedDownload EventType = "FinishedDownload"
)

// ManagerEvent is an event originating form a Manager instance
type ManagerEvent struct {
	Type EventType

	// Progress is a number ∈ [0, 1]
	Progress float64

	// Err is a potential error. It might be non-nil if `Type` is `ETFinishedDownload`
	Err error
}

// SetEventFunc sets the function that will be called with Manager events.
// Used for websocket notifications
func (m *Manager) SetEventFunc(f func(f func(c *websocket.Conn) error)) {
	m.evtFunc = f
}

func (m *Manager) event(evtType string, evtData interface{}) {
	go m.evtFunc(func(c *websocket.Conn) error {
		return c.WriteJSON(map[string]interface{}{
			"type": evtType,
			"data": evtData,
		})
	})
}
