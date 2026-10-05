package api

import (
	"net/http"
	"slices"
	"testing"
	"time"
	"xarantolus/sensibleHub/store/analysis"
	"xarantolus/sensibleHub/store/music"
)

func syncedSong(id string) music.Entry {
	e := song(id, id)
	e.SyncSettings.Should = true
	return e
}

func withAnalysis(e music.Entry, a music.Analysis) music.Entry {
	a.Version = analysis.Version
	e.Analysis = &a
	return e
}

func TestNextSongsReturnsOnlySyncedSongsOtherThanCurrent(t *testing.T) {
	a := newTestAPI(t, syncedSong("aaaa"), song("bbbb", "unsynced"), syncedSong("cccc"), syncedSong("dddd"))

	for range 20 {
		rec := a.do(http.MethodGet, "/api/v1/player/next?current=aaaa&count=10", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d\n%s", rec.Code, rec.Body.String())
		}
		got := decode[NextSongs](t, rec)
		var ids []string
		for _, s := range got.Songs {
			ids = append(ids, s.ID)
		}
		slices.Sort(ids)
		if !slices.Equal(ids, []string{"cccc", "dddd"}) {
			t.Fatalf("got %v, want the synced songs other than the current one", ids)
		}
	}
}

func TestNextSongsHonoursCount(t *testing.T) {
	a := newTestAPI(t, syncedSong("aaaa"), syncedSong("bbbb"), syncedSong("cccc"), syncedSong("dddd"))

	if got := decode[NextSongs](t, a.do(http.MethodGet, "/api/v1/player/next?count=2", "")); len(got.Songs) != 2 {
		t.Errorf("got %d songs, want 2", len(got.Songs))
	}
	if got := decode[NextSongs](t, a.do(http.MethodGet, "/api/v1/player/next", "")); len(got.Songs) != 4 {
		t.Errorf("default count should return up to 5 of the 4 songs, got %d", len(got.Songs))
	}
}

func TestNextSongsFactorsIsAlwaysAnObject(t *testing.T) {
	a := newTestAPI(t, syncedSong("aaaa"), syncedSong("bbbb"))

	rec := a.do(http.MethodGet, "/api/v1/player/next?count=1", "")
	raw := decode[struct {
		Songs []map[string]any `json:"songs"`
	}](t, rec)
	if len(raw.Songs) != 1 {
		t.Fatalf("got %v", raw.Songs)
	}
	factors, ok := raw.Songs[0]["factors"].(map[string]any)
	if !ok {
		t.Fatalf("factors should be an object, got %#v", raw.Songs[0]["factors"])
	}
	if _, ok := factors["recency"].(float64); !ok {
		t.Errorf("factors should explain the pick, got %v", factors)
	}
	if _, ok := raw.Songs[0]["score"].(float64); !ok {
		t.Errorf("score missing: %v", raw.Songs[0])
	}
}

func TestNextSongsEmptyLibraryReturnsEmptyList(t *testing.T) {
	a := newTestAPI(t)
	rec := a.do(http.MethodGet, "/api/v1/player/next", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d\n%s", rec.Code, rec.Body.String())
	}
	raw := decode[map[string]any](t, rec)
	if songs, ok := raw["songs"].([]any); !ok || len(songs) != 0 {
		t.Errorf("songs should be an empty array, got %#v", raw["songs"])
	}
}

func TestNextSongsRejectsInvalidCount(t *testing.T) {
	a := newTestAPI(t, syncedSong("aaaa"))
	for _, q := range []string{"count=0", "count=51", "count=-1"} {
		rec := a.do(http.MethodGet, "/api/v1/player/next?"+q, "")
		a.expectProblem(rec, http.StatusUnprocessableEntity, "validation")
	}
}

func TestNextSongsPrefersMusicallyFittingSongs(t *testing.T) {
	cur := withAnalysis(syncedSong("cur0"), music.Analysis{Key: 0, KeyStrength: 1, BPM: 120, BeatStrength: 1, Timbre: []float64{1}})
	fits := withAnalysis(syncedSong("fits"), music.Analysis{Key: 0, KeyStrength: 1, BPM: 120, BeatStrength: 1, Timbre: []float64{1}})
	clashes := withAnalysis(syncedSong("clsh"), music.Analysis{Key: 6, KeyStrength: 1, BPM: 170, BeatStrength: 1, Timbre: []float64{1}})
	a := newTestAPI(t, cur, fits, clashes)

	counts := map[string]int{}
	for range 200 {
		got := decode[NextSongs](t, a.do(http.MethodGet, "/api/v1/player/next?current=cur0&count=1", ""))
		counts[got.Songs[0].ID]++
	}
	if counts["fits"] < 4*counts["clsh"] {
		t.Errorf("fitting song should be picked far more often: %v", counts)
	}
}

func TestAnalysisStatusShape(t *testing.T) {
	a := newTestAPI(t, song("aaaa", "A"), song("bbbb", "B"))

	rec := a.do(http.MethodGet, "/api/v1/analysis", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d\n%s", rec.Code, rec.Body.String())
	}
	raw := decode[map[string]any](t, rec)
	for _, k := range []string{"running", "done", "total"} {
		if _, ok := raw[k]; !ok {
			t.Errorf("field %q missing from %v", k, raw)
		}
	}
	got := decode[AnalysisStatus](t, rec)
	if got.Running || got.Done != 0 || got.Total != 2 {
		t.Errorf("got %+v, want idle with 2 queued songs", got)
	}
}

func TestSongDetailAnalysisStatus(t *testing.T) {
	stale := withAnalysis(song("stal", "Stale"), music.Analysis{Timbre: []float64{1}})
	stale.Analysis.Version = analysis.Version - 1
	failed := withAnalysis(song("fail", "Failed"), music.Analysis{Error: "decode failed", AnalyzedAt: time.Now()})
	done := withAnalysis(song("done", "Done"), music.Analysis{
		Key: 9, Minor: true, KeyStrength: 0.8, Camelot: "8A", BPM: 128, BeatStrength: 0.7,
		LoudnessLUFS: -14.5, Timbre: []float64{1}, AnalyzedAt: time.Now(),
	})
	a := newTestAPI(t, song("pend", "Pending"), stale, failed, done)

	get := func(id string) SongDetail {
		rec := a.do(http.MethodGet, "/api/v1/songs/"+id, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d\n%s", id, rec.Code, rec.Body.String())
		}
		return decode[SongDetail](t, rec)
	}

	if got := get("pend"); got.Analysis.Status != "pending" || got.Loudness != nil {
		t.Errorf("unanalysed: %+v loudness %v", got.Analysis, got.Loudness)
	}
	if got := get("stal"); got.Analysis.Status != "pending" || got.Loudness == nil {
		t.Errorf("outdated analysis should be pending: %+v", got.Analysis)
	}

	got := get("fail")
	if got.Analysis.Status != "failed" || got.Analysis.Error != "decode failed" {
		t.Errorf("failed: %+v", got.Analysis)
	}
	if got.Loudness != nil {
		t.Errorf("failed analysis must not expose loudness, got %v", *got.Loudness)
	}

	got = get("done")
	if got.Analysis.Status != "done" || got.Analysis.Camelot != "8A" || got.Analysis.BPM != 128 || got.Analysis.Key == "" {
		t.Errorf("done: %+v", got.Analysis)
	}
	if got.Loudness == nil || *got.Loudness != -14.5 {
		t.Errorf("loudness = %v, want -14.5", got.Loudness)
	}
}

func TestSongSummaryLoudnessOnlyForUsableAnalyses(t *testing.T) {
	failed := withAnalysis(song("fail", "Failed"), music.Analysis{Error: "x", LoudnessLUFS: -9})
	done := withAnalysis(song("done", "Done"), music.Analysis{LoudnessLUFS: -12, Timbre: []float64{1}})
	a := newTestAPI(t, song("pend", "Pending"), failed, done)

	rec := a.do(http.MethodGet, "/api/v1/songs", "")
	got := map[string]SongSummary{}
	for _, s := range decode[[]SongSummary](t, rec) {
		got[s.ID] = s
	}
	if got["pend"].Loudness != nil || got["fail"].Loudness != nil {
		t.Errorf("loudness leaked for unusable analyses: %v %v", got["pend"].Loudness, got["fail"].Loudness)
	}
	if l := got["done"].Loudness; l == nil || *l != -12 {
		t.Errorf("loudness = %v, want -12", l)
	}
}
