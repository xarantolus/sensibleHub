package api

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"xarantolus/sensibleHub/store"

	"github.com/danielgtaylor/huma/v2"
)

type Home struct {
	SongIDs []string `json:"songIds"`
	Today   bool     `json:"today" doc:"Whether all songs were added today"`
}

type SearchResult struct {
	SongIDs []string `json:"songIds" doc:"Best match first"`
}

func registerLibrary(api huma.API, m *store.Manager) {
	listings := map[string]func() []store.Group{
		"title":      m.GroupByTitle,
		"artist":     m.GroupByArtist,
		"year":       m.GroupByYear,
		"incomplete": m.Incomplete,
		"unsynced":   m.Unsynced,
		"edited":     m.RecentlyEdited,
		"added":      m.SortedByAddDate,
	}

	huma.Register(api, huma.Operation{
		OperationID: "getListing", Method: http.MethodGet, Path: base + "/listings/{kind}",
		Summary: "Songs grouped by a criterion",
	}, func(ctx context.Context, in *struct {
		Kind string `path:"kind" enum:"title,artist,year,incomplete,unsynced,edited,added"`
	}) (*body[[]Group], error) {
		groups := listings[in.Kind]()
		out := make([]Group, len(groups))
		for i, g := range groups {
			out[i] = Group{Title: g.Title, Description: g.Description, Link: escapePath(g.Link), SongIDs: ids(g.Songs)}
		}
		return &body[[]Group]{out}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "getHome", Method: http.MethodGet, Path: base + "/home",
		Summary: "Newest songs",
	}, func(ctx context.Context, _ *struct{}) (*body[Home], error) {
		entries, today := m.Newest()
		return &body[Home]{Home{SongIDs: ids(entries), Today: today}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "search", Method: http.MethodGet, Path: base + "/search",
		Summary: "Search songs by title, artist, album and year",
	}, func(ctx context.Context, in *struct {
		Query string `query:"q" required:"true" minLength:"1"`
		Limit int    `query:"limit" minimum:"0" doc:"0 means no limit"`
	}) (*body[SearchResult], error) {
		res := m.Search(in.Query)
		if in.Limit > 0 && len(res) > in.Limit {
			res = res[:in.Limit]
		}
		return &body[SearchResult]{SearchResult{SongIDs: ids(res)}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "getAlbum", Method: http.MethodGet, Path: base + "/albums/{artist}/{album}",
		Summary: "An album of an artist",
	}, func(ctx context.Context, in *AlbumPath) (*body[Album], error) {
		a, ok := m.GetAlbum(in.Artist, in.Album)
		if !ok {
			return nil, toProblem("get album", &store.NotFoundError{Kind: "album", Key: in.Artist + "/" + in.Album})
		}
		return &body[Album]{album(a)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "setAlbumCover", Method: http.MethodPut, Path: base + "/albums/{artist}/{album}/cover",
		Summary: "Set the cover of every song in an album", DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, in *struct {
		AlbumPath
		RawBody huma.MultipartFormFiles[coverUpload]
	}) (*struct{}, error) {
		f := in.RawBody.Data().Cover
		if err := m.EditAlbumCover(in.Artist, in.Album, f.Filename, f); err != nil {
			return nil, toProblem("set album cover", err)
		}
		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "getArtist", Method: http.MethodGet, Path: base + "/artists/{artist}",
		Summary: "An artist's albums and features",
	}, func(ctx context.Context, in *struct {
		Artist string `path:"artist"`
	}) (*body[Artist], error) {
		info, ok := m.Artist(in.Artist)
		if !ok {
			return nil, toProblem("get artist", &store.NotFoundError{Kind: "artist", Key: in.Artist})
		}
		a := Artist{
			Name:      info.Name,
			PlayTime:  info.PlayTime,
			YearStart: info.YearStart,
			YearEnd:   info.YearEnd,
			Albums:    make([]Album, len(info.Albums)),
			Featured:  ids(info.Featured),
		}
		for i, al := range info.Albums {
			a.Albums[i] = album(al)
		}
		return &body[Artist]{a}, nil
	})
}

func escapePath(p string) string {
	segments := strings.Split(p, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return strings.Join(segments, "/")
}

type AlbumPath struct {
	Artist string `path:"artist"`
	Album  string `path:"album"`
}
