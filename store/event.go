package store

import "github.com/gorilla/websocket"

// EventType is an event type for the Manager
type EventType string

// Event types that are possible
const (
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
