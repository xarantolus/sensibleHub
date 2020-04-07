package store

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
