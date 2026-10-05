package api

import (
	"context"
	"net/http"
	"xarantolus/sensibleHub/store"

	"github.com/danielgtaylor/huma/v2"
)

func registerDownloads(api huma.API, m *store.Manager) {
	status := func() DownloadStatus {
		url, _ := m.IsDownloading()
		return DownloadStatus{
			Running:   m.IsWorking(),
			URL:       url,
			Queued:    m.QueueLength(),
			LastError: downloadFailure(m.LastError()),
		}
	}

	huma.Register(api, huma.Operation{
		OperationID: "getDownloads", Method: http.MethodGet, Path: base + "/downloads",
		Summary: "Download status",
	}, func(ctx context.Context, _ *struct{}) (*body[DownloadStatus], error) {
		return &body[DownloadStatus]{status()}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "enqueueDownload", Method: http.MethodPost, Path: base + "/downloads",
		Summary: "Queue a download", DefaultStatus: http.StatusAccepted,
		Description: "Accepts a link to any site yt-dlp supports, or a search term that is looked up on YouTube Music.",
	}, func(ctx context.Context, in *struct {
		Body struct {
			Query string `json:"query" minLength:"1" maxLength:"2048" doc:"URL or search term"`
		}
	}) (*body[DownloadStatus], error) {
		if err := m.Enqueue(in.Body.Query); err != nil {
			return nil, toProblem("enqueue", err)
		}
		return &body[DownloadStatus]{status()}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "abortDownload", Method: http.MethodDelete, Path: base + "/downloads/current",
		Summary: "Abort the running download", DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, _ *struct{}) (*struct{}, error) {
		if err := m.AbortDownload(); err != nil {
			return nil, toProblem("abort", err)
		}
		return nil, nil
	})
}
