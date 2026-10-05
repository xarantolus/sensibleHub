package api

import (
	"context"
	"net/http"
	"xarantolus/sensibleHub/store"

	"github.com/danielgtaylor/huma/v2"
)

// SongIDPath is exported because Huma ignores embedded structs of unexported types.
type SongIDPath struct {
	ID string `path:"id" doc:"Song ID"`
}

type songEditBody struct {
	Title  string  `json:"title" minLength:"1"`
	Artist string  `json:"artist"`
	Album  string  `json:"album"`
	Year   *int    `json:"year,omitempty" minimum:"0" maximum:"9999" doc:"Omit to clear the year"`
	Start  float64 `json:"start" minimum:"0"`
	End    float64 `json:"end" minimum:"0"`
	Sync   bool    `json:"sync"`
}

type coverUpload struct {
	Cover huma.FormFile `form:"cover" contentType:"image/png,image/jpeg,image/webp,image/gif" required:"true"`
}

func registerSongs(api huma.API, m *store.Manager) {
	huma.Register(api, huma.Operation{
		OperationID: "listSongs", Method: http.MethodGet, Path: base + "/songs",
		Summary: "All songs", Description: "Every song in the collection, sorted by title. Other endpoints refer to songs by ID only.",
	}, func(ctx context.Context, _ *struct{}) (*body[[]SongSummary], error) {
		entries := m.AllEntries()
		out := make([]SongSummary, len(entries))
		for i, e := range entries {
			out[i] = songSummary(e)
		}
		return &body[[]SongSummary]{out}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "getRandomSong", Method: http.MethodGet, Path: base + "/songs/random",
		Summary: "A random song",
	}, func(ctx context.Context, _ *struct{}) (*body[SongSummary], error) {
		e, ok := m.RandomSong()
		if !ok {
			return nil, toProblem("random song", &store.NotFoundError{Kind: "song", Key: "random"})
		}
		return &body[SongSummary]{songSummary(e)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "getSong", Method: http.MethodGet, Path: base + "/songs/{id}",
		Summary: "Song details",
	}, func(ctx context.Context, in *SongIDPath) (*body[SongDetail], error) {
		e, ok := m.GetEntry(in.ID)
		if !ok {
			return nil, toProblem("get song", &store.NotFoundError{Kind: "song", Key: in.ID})
		}
		return &body[SongDetail]{songDetail(e, m.GetRelatedSongs(e))}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "updateSong", Method: http.MethodPut, Path: base + "/songs/{id}",
		Summary: "Edit a song", Description: "Replaces all editable fields of a song.",
	}, func(ctx context.Context, in *struct {
		SongIDPath
		Body songEditBody
	}) (*body[SongSummary], error) {
		e, err := m.UpdateEntry(in.ID, store.SongEdit{
			Title:  in.Body.Title,
			Artist: in.Body.Artist,
			Album:  in.Body.Album,
			Year:   in.Body.Year,
			Start:  in.Body.Start,
			End:    in.Body.End,
			Sync:   in.Body.Sync,
		})
		if err != nil {
			return nil, toProblem("update song", err)
		}
		return &body[SongSummary]{songSummary(e)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "deleteSong", Method: http.MethodDelete, Path: base + "/songs/{id}",
		Summary: "Delete a song and its files", DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, in *SongIDPath) (*struct{}, error) {
		if err := m.DeleteEntry(in.ID); err != nil {
			return nil, toProblem("delete song", err)
		}
		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "setSongCover", Method: http.MethodPut, Path: base + "/songs/{id}/cover",
		Summary: "Replace a song's cover", Description: "The image is cropped to a square.",
	}, func(ctx context.Context, in *struct {
		SongIDPath
		RawBody huma.MultipartFormFiles[coverUpload]
	}) (*body[SongSummary], error) {
		f := in.RawBody.Data().Cover
		defer f.Close()

		e, err := m.SetCover(in.ID, f, f.Filename)
		if err != nil {
			return nil, toProblem("set cover", err)
		}
		return &body[SongSummary]{songSummary(e)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "deleteSongCover", Method: http.MethodDelete, Path: base + "/songs/{id}/cover",
		Summary: "Remove a song's cover",
	}, func(ctx context.Context, in *SongIDPath) (*body[SongSummary], error) {
		if err := m.DeleteCoverImage(in.ID); err != nil {
			return nil, toProblem("delete cover", err)
		}
		e, ok := m.GetEntry(in.ID)
		if !ok {
			return nil, toProblem("delete cover", &store.NotFoundError{Kind: "song", Key: in.ID})
		}
		return &body[SongSummary]{songSummary(e)}, nil
	})
}
