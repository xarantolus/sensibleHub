package api

import (
	"context"
	"net/http"
	"time"
	"xarantolus/sensibleHub/store"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
)

type SongAddedEvent struct {
	Song SongSummary `json:"song"`
}

type SongUpdatedEvent struct {
	Song SongSummary `json:"song"`
}

type SongDeletedEvent struct {
	ID string `json:"id"`
}

type ArtistSyncChangedEvent struct {
	Artist string `json:"artist"`
	Sync   bool   `json:"sync"`
}

type DownloadStartedEvent struct{}

type DownloadFinishedEvent struct {
	Error *DownloadFailure `json:"error,omitempty"`
}

const keepAlive = 25 * time.Second

func registerEvents(api huma.API, m *store.Manager) {
	sse.Register(api, huma.Operation{
		OperationID: "events", Method: http.MethodGet, Path: base + "/events",
		Summary:     "Live updates",
		Description: "Server-sent events for library and download changes. The stream ends if the client falls too far behind; reconnect and refetch.",
	}, map[string]any{
		"songAdded":         SongAddedEvent{},
		"songUpdated":       SongUpdatedEvent{},
		"songDeleted":       SongDeletedEvent{},
		"artistSyncChanged": ArtistSyncChangedEvent{},
		"downloadStarted":   DownloadStartedEvent{},
		"downloadFinished":  DownloadFinishedEvent{},
		"analysisProgress":  AnalysisStatus{},
	}, func(ctx context.Context, _ *struct{}, send sse.Sender) {
		events, cancel := m.Subscribe()
		defer cancel()

		if m.IsWorking() {
			if send.Data(DownloadStartedEvent{}) != nil {
				return
			}
		}

		ticker := time.NewTicker(keepAlive)
		defer ticker.Stop()

		for {
			var err error
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err = send.Comment("keep-alive")
			case e, ok := <-events:
				if !ok {
					return
				}
				err = send.Data(eventPayload(m, e))
			}
			if err != nil {
				return
			}
		}
	})
}

func eventPayload(m *store.Manager, e store.Event) any {
	switch e := e.(type) {
	case store.SongAdded:
		return SongAddedEvent{Song: songSummary(m, e.Song)}
	case store.SongUpdated:
		return SongUpdatedEvent{Song: songSummary(m, e.Song)}
	case store.ArtistSyncChanged:
		return ArtistSyncChangedEvent{Artist: e.Artist, Sync: e.Sync}
	case store.SongDeleted:
		return SongDeletedEvent{ID: e.ID}
	case store.DownloadStarted:
		return DownloadStartedEvent{}
	case store.DownloadFinished:
		return DownloadFinishedEvent{Error: downloadFailure(e.Err)}
	case store.AnalysisProgress:
		return AnalysisStatus{Running: e.Running, Done: e.Done, Total: e.Total}
	}
	panic("unhandled event type")
}
