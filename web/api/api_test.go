package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"xarantolus/sensibleHub/store"
	"xarantolus/sensibleHub/store/config"
	"xarantolus/sensibleHub/store/music"

	"github.com/gorilla/mux"
)

type testAPI struct {
	t       *testing.T
	handler http.Handler
	manager *store.Manager
}

func newTestAPI(t *testing.T, songs ...music.Entry) *testAPI {
	t.Helper()
	t.Chdir(t.TempDir())

	m, err := store.NewManager(config.Config{})
	if err != nil {
		t.Fatal(err)
	}

	m.SongsLock.Lock()
	for i := range songs {
		if err := m.Add(&songs[i]); err != nil {
			t.Fatal(err)
		}
	}
	m.SongsLock.Unlock()

	r := mux.NewRouter()
	New(r, m)
	return &testAPI{t: t, handler: r, manager: m}
}

func (a *testAPI) do(method, path, body string) *httptest.ResponseRecorder {
	a.t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	a.handler.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("response is not valid JSON for %T: %v\n%s", v, err, rec.Body.String())
	}
	return v
}

type problemBody struct {
	Status   int    `json:"status"`
	Code     string `json:"code"`
	Resource string `json:"resource"`
	Errors   []struct {
		Message  string `json:"message"`
		Location string `json:"location"`
	} `json:"errors"`
}

func (a *testAPI) expectProblem(rec *httptest.ResponseRecorder, status int, code string) problemBody {
	a.t.Helper()
	if rec.Code != status {
		a.t.Fatalf("status = %d, want %d\n%s", rec.Code, status, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/problem+json") {
		a.t.Errorf("content type = %q, want application/problem+json", ct)
	}
	p := decode[problemBody](a.t, rec)
	if p.Code != code {
		a.t.Errorf("code = %q, want %q", p.Code, code)
	}
	return p
}

func editJSON(title string, start, end float64) string {
	b, err := json.Marshal(map[string]any{
		"title": title, "artist": "", "album": "", "start": start, "end": end, "sync": false,
	})
	if err != nil {
		panic(err)
	}
	return string(b)
}

func song(id, title string) music.Entry {
	return music.Entry{
		ID:            id,
		Added:         time.Now(),
		LastEdit:      time.Now(),
		MusicData:     music.MusicData{Title: title, Duration: 100},
		AudioSettings: music.AudioSettings{Start: 0, End: 100},
	}
}

func TestListSongs(t *testing.T) {
	unset := song("aaaa", "Unset Bounds")
	unset.AudioSettings = music.AudioSettings{Start: -1, End: -1}
	trimmed := song("bbbb", "Trimmed")
	trimmed.AudioSettings = music.AudioSettings{Start: 10, End: 60}
	covered := song("cccc", "Covered")
	covered.PictureData = music.PictureData{Filename: "cover.jpg", Size: 300, DominantColorHEX: "#aabbcc"}

	a := newTestAPI(t, unset, trimmed, covered)

	rec := a.do(http.MethodGet, "/api/v1/songs", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d\n%s", rec.Code, rec.Body.String())
	}
	got := map[string]SongSummary{}
	for _, s := range decode[[]SongSummary](t, rec) {
		got[s.ID] = s
	}
	if len(got) != 3 {
		t.Fatalf("got %d songs, want 3", len(got))
	}

	if p := got["aaaa"].Playback; p.Start != 0 || p.End != 100 {
		t.Errorf("unset bounds should play the whole file, got %+v", p)
	}
	if p := got["bbbb"].Playback; p.Start != 10 || p.End != 60 {
		t.Errorf("stored bounds should be kept, got %+v", p)
	}
	if got["aaaa"].Cover != nil || got["bbbb"].Cover != nil {
		t.Error("songs without a cover file must not have a cover")
	}
	if c := got["cccc"].Cover; c == nil || c.Size != 300 || c.Color != "#aabbcc" {
		t.Errorf("cover = %+v, want size 300 color #aabbcc", c)
	}
	if strings.Contains(rec.Body.String(), `"cover":null`) {
		t.Error("absent cover must be omitted, not null")
	}
}

func TestListSongsEmptyIsArray(t *testing.T) {
	a := newTestAPI(t)
	rec := a.do(http.MethodGet, "/api/v1/songs", "")
	if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
		t.Errorf("body = %s, want []", got)
	}
}

func TestGetSongArraysAreNeverNull(t *testing.T) {
	a := newTestAPI(t, song("aaaa", "Lonely"))

	rec := a.do(http.MethodGet, "/api/v1/songs/aaaa", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d\n%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"related":[]`) {
		t.Errorf("related should be an empty array: %s", rec.Body.String())
	}

	rec = a.do(http.MethodGet, "/api/v1/listings/unsynced", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d\n%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "null") {
		t.Errorf("listing contains null: %s", rec.Body.String())
	}

	rec = a.do(http.MethodGet, "/api/v1/search?q=nothingmatchesthis", "")
	if !strings.Contains(rec.Body.String(), `"songIds":[]`) {
		t.Errorf("search without matches should return an empty array: %s", rec.Body.String())
	}
}

func TestGetUnknownSong(t *testing.T) {
	a := newTestAPI(t)
	p := a.expectProblem(a.do(http.MethodGet, "/api/v1/songs/nope", ""), http.StatusNotFound, "not_found")
	if p.Resource != "song" {
		t.Errorf("resource = %q, want song", p.Resource)
	}
}

func TestUnknownListingIsRejected(t *testing.T) {
	a := newTestAPI(t)
	a.expectProblem(a.do(http.MethodGet, "/api/v1/listings/bogus", ""), http.StatusUnprocessableEntity, "validation")
}

func TestUpdateSong(t *testing.T) {
	year := 1999
	orig := song("aaaa", "Old")
	orig.MusicData.Year = &year
	orig.MusicData.Artist = "Someone"
	a := newTestAPI(t, orig)

	rec := a.do(http.MethodPut, "/api/v1/songs/aaaa",
		`{"title":"  New  ","artist":"Other","album":"LP","start":5,"end":50,"sync":true}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d\n%s", rec.Code, rec.Body.String())
	}

	got := decode[SongSummary](t, a.do(http.MethodGet, "/api/v1/songs/aaaa", ""))
	if got.Title != "New" || got.Artist != "Other" || got.Album != "LP" {
		t.Errorf("tags not updated: %+v", got)
	}
	if got.Year != nil {
		t.Errorf("year should be cleared when omitted, got %d", *got.Year)
	}
	if got.Playback.Start != 5 || got.Playback.End != 50 || !got.Sync {
		t.Errorf("playback/sync not updated: %+v", got)
	}
}

func TestUpdateSongRejectsInvalidInput(t *testing.T) {
	a := newTestAPI(t, song("aaaa", "Keep"))

	t.Run("start not before end", func(t *testing.T) {
		rec := a.do(http.MethodPut, "/api/v1/songs/aaaa", editJSON("X", 50, 50))
		p := a.expectProblem(rec, http.StatusUnprocessableEntity, "validation")
		found := false
		for _, e := range p.Errors {
			if e.Location == "body.end" {
				found = true
			}
		}
		if !found {
			t.Errorf("want an error located at body.end, got %+v", p.Errors)
		}
	})

	t.Run("end beyond duration", func(t *testing.T) {
		rec := a.do(http.MethodPut, "/api/v1/songs/aaaa", editJSON("X", 0, 101))
		a.expectProblem(rec, http.StatusUnprocessableEntity, "validation")
	})

	t.Run("empty title", func(t *testing.T) {
		rec := a.do(http.MethodPut, "/api/v1/songs/aaaa", editJSON("", 0, 10))
		a.expectProblem(rec, http.StatusUnprocessableEntity, "validation")
	})

	t.Run("blank title", func(t *testing.T) {
		rec := a.do(http.MethodPut, "/api/v1/songs/aaaa", editJSON("   ", 0, 10))
		a.expectProblem(rec, http.StatusUnprocessableEntity, "validation")
	})

	got := decode[SongSummary](t, a.do(http.MethodGet, "/api/v1/songs/aaaa", ""))
	if got.Title != "Keep" {
		t.Errorf("rejected edits must not change the song, title is %q", got.Title)
	}
}

func TestUpdateUnknownSong(t *testing.T) {
	a := newTestAPI(t)
	rec := a.do(http.MethodPut, "/api/v1/songs/nope", editJSON("X", 0, 10))
	a.expectProblem(rec, http.StatusNotFound, "not_found")
}

func TestAbortWithoutDownload(t *testing.T) {
	a := newTestAPI(t)
	a.expectProblem(a.do(http.MethodDelete, "/api/v1/downloads/current", ""), http.StatusConflict, "not_downloading")
}

func TestDeleteSong(t *testing.T) {
	a := newTestAPI(t, song("aaaa", "Gone"))

	if rec := a.do(http.MethodDelete, "/api/v1/songs/aaaa", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204\n%s", rec.Code, rec.Body.String())
	}
	a.expectProblem(a.do(http.MethodGet, "/api/v1/songs/aaaa", ""), http.StatusNotFound, "not_found")
	a.expectProblem(a.do(http.MethodDelete, "/api/v1/songs/aaaa", ""), http.StatusNotFound, "not_found")
}

func nextEvent(t *testing.T, events <-chan store.Event) store.Event {
	t.Helper()
	select {
	case e, ok := <-events:
		if !ok {
			t.Fatal("event channel closed")
		}
		return e
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for an event")
		return nil
	}
}

func TestMutationsPublishEvents(t *testing.T) {
	a := newTestAPI(t, song("aaaa", "Watched"))
	events, cancel := a.manager.Subscribe()
	defer cancel()

	a.do(http.MethodPut, "/api/v1/songs/aaaa", editJSON("Renamed", 0, 10))
	updated, ok := nextEvent(t, events).(store.SongUpdated)
	if !ok || updated.Song.ID != "aaaa" || updated.Song.MusicData.Title != "Renamed" {
		t.Errorf("want SongUpdated for aaaa with the new title, got %+v", updated)
	}

	a.do(http.MethodDelete, "/api/v1/songs/aaaa", "")
	deleted, ok := nextEvent(t, events).(store.SongDeleted)
	if !ok || deleted.ID != "aaaa" {
		t.Errorf("want SongDeleted for aaaa, got %+v", deleted)
	}
}
