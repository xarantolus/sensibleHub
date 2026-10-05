package api

import (
	"context"
	"net/http"
	"xarantolus/sensibleHub/store"

	"github.com/danielgtaylor/huma/v2"
)

type NextSong struct {
	ID      string             `json:"id"`
	Score   float64            `json:"score"`
	Factors map[string]float64 `json:"factors" doc:"Why the song was picked; 1 is neutral. For debugging only."`
}

type NextSongs struct {
	Songs []NextSong `json:"songs"`
}

func registerPlayer(api huma.API, m *store.Manager) {
	huma.Register(api, huma.Operation{
		OperationID: "getAnalysisStatus", Method: http.MethodGet, Path: base + "/analysis",
		Summary: "Progress of the background audio analysis",
	}, func(ctx context.Context, _ *struct{}) (*body[AnalysisStatus], error) {
		s := m.AnalysisStatus()
		return &body[AnalysisStatus]{AnalysisStatus{Running: s.Running, Done: s.Done, Total: s.Total}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "nextSongs", Method: http.MethodGet, Path: base + "/player/next",
		Summary:     "Songs to play next",
		Description: "Picks synced songs to continue playback after the queue runs out. Results are random; call again for more.",
	}, func(ctx context.Context, in *struct {
		Current string   `query:"current" doc:"The song playing now, if any"`
		Count   int      `query:"count" minimum:"1" maximum:"50" default:"5"`
		Exclude []string `query:"exclude" maxItems:"500" doc:"Recently played songs to avoid"`
	}) (*body[NextSongs], error) {
		picks := m.NextSongs(in.Current, in.Count, in.Exclude)
		out := NextSongs{Songs: make([]NextSong, len(picks))}
		for i, p := range picks {
			f := p.Factors
			if f == nil {
				f = map[string]float64{}
			}
			out.Songs[i] = NextSong{ID: p.ID, Score: p.Score, Factors: f}
		}
		return &body[NextSongs]{out}, nil
	})
}
