package store

import (
	"sync"
	"xarantolus/sensibleHub/store/music"
)

type Event interface{ isEvent() }

type SongAdded struct{ Song music.Entry }
type SongUpdated struct{ Song music.Entry }
type SongDeleted struct{ ID string }
type DownloadStarted struct{}
type DownloadFinished struct{ Err *DownloadError }

func (SongAdded) isEvent()        {}
func (SongUpdated) isEvent()      {}
func (SongDeleted) isEvent()      {}
func (DownloadStarted) isEvent()  {}
func (DownloadFinished) isEvent() {}

type broker struct {
	mu   sync.Mutex
	subs map[chan Event]struct{}
}

// Subscribe returns a channel receiving all future events. A subscriber that
// falls behind by more than the buffer gets its channel closed instead of blocking
// the store, so it can resubscribe and refetch rather than silently miss events.
func (m *Manager) Subscribe() (events <-chan Event, cancel func()) {
	c := make(chan Event, 64)

	m.events.mu.Lock()
	if m.events.subs == nil {
		m.events.subs = make(map[chan Event]struct{})
	}
	m.events.subs[c] = struct{}{}
	m.events.mu.Unlock()

	return c, func() {
		m.events.mu.Lock()
		defer m.events.mu.Unlock()
		if _, ok := m.events.subs[c]; ok {
			delete(m.events.subs, c)
			close(c)
		}
	}
}

func (m *Manager) publish(e Event) {
	m.events.mu.Lock()
	defer m.events.mu.Unlock()

	for c := range m.events.subs {
		select {
		case c <- e:
		default:
			delete(m.events.subs, c)
			close(c)
		}
	}
}
