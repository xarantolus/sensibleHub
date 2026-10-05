package web

import (
	"net/http"
	"xarantolus/sensibleHub/store"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// HandleWebsocket serves the event feed of the legacy frontend until the connection breaks.
func (s *server) HandleWebsocket(w http.ResponseWriter, r *http.Request) (err error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil // Upgrader has already responded to the request
	}
	defer conn.Close()

	events, cancel := s.m.Subscribe()
	defer cancel()

	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				cancel()
				return
			}
		}
	}()

	if s.m.IsWorking() {
		if err := conn.WriteJSON(map[string]any{"type": "progress-start"}); err != nil {
			return nil
		}
	}

	for e := range events {
		if err := conn.WriteJSON(legacyEvent(e)); err != nil {
			return nil
		}
	}
	return nil
}

func legacyEvent(e store.Event) map[string]any {
	switch e := e.(type) {
	case store.SongAdded:
		return map[string]any{"type": "song-add", "data": map[string]any{"id": e.Song.ID, "song": e.Song}}
	case store.SongUpdated:
		return map[string]any{"type": "song-edit", "data": map[string]any{"id": e.Song.ID, "song": e.Song}}
	case store.SongDeleted:
		return map[string]any{"type": "song-delete", "data": map[string]any{"id": e.ID}}
	case store.DownloadStarted:
		return map[string]any{"type": "progress-start"}
	case store.DownloadFinished:
		data := map[string]string{}
		if e.Err != nil {
			data["error"] = e.Err.Error()
		}
		return map[string]any{"type": "progress-end", "data": data}
	}
	return map[string]any{"type": "unknown"}
}
