package store

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
	"xarantolus/sensibleHub/store/analysis"
	"xarantolus/sensibleHub/store/config"
	"xarantolus/sensibleHub/store/music"
)

func usableAnalysis(start, end float64) *music.Analysis {
	return &music.Analysis{Version: analysis.Version, Start: start, End: end, Timbre: []float64{1}}
}

func trimmedSong(duration, start, end float64) music.Entry {
	e := music.Entry{ID: "x"}
	e.MusicData.Duration = duration
	e.AudioSettings = music.AudioSettings{Start: start, End: end}
	return e
}

func TestNeedsAnalysis(t *testing.T) {
	tests := []struct {
		name string
		edit func(e *music.Entry)
		want bool
	}{
		{"no analysis", func(e *music.Entry) {}, true},
		{"older version", func(e *music.Entry) {
			a := usableAnalysis(0, 0)
			a.Version = analysis.Version - 1
			e.Analysis = a
		}, true},
		{"current", func(e *music.Entry) { e.Analysis = usableAnalysis(0, 0) }, false},
		{"failed but current", func(e *music.Entry) {
			a := usableAnalysis(0, 0)
			a.Timbre = nil
			a.Error = "broken file"
			e.Analysis = a
		}, false},
		{"start trim changed", func(e *music.Entry) {
			e.Analysis = usableAnalysis(0, 0)
			e.AudioSettings.Start = 5
		}, true},
		{"end trim changed", func(e *music.Entry) {
			e.Analysis = usableAnalysis(0, 0)
			e.AudioSettings.End = 60
		}, true},
		{"trim unchanged", func(e *music.Entry) {
			e.Analysis = usableAnalysis(5, 60)
			e.AudioSettings = music.AudioSettings{Start: 5, End: 60}
		}, false},
		{"failed with old trim is retried", func(e *music.Entry) {
			e.Analysis = &music.Analysis{Version: analysis.Version, Error: "x"}
			e.AudioSettings.Start = 3
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := trimmedSong(100, 0, 0)
			tt.edit(&e)
			if got := needsAnalysis(e); got != tt.want {
				t.Errorf("needsAnalysis = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimBoundsNormalisesUnsetAndFullLengthEnds(t *testing.T) {
	tests := []struct {
		name               string
		e                  music.Entry
		wantStart, wantEnd float64
	}{
		{"unset", trimmedSong(100, -1, -1), 0, 0},
		{"zero", trimmedSong(100, 0, 0), 0, 0},
		{"full length", trimmedSong(100, 0, 100), 0, 0},
		{"beyond length", trimmedSong(100, 0, 120), 0, 0},
		{"trimmed", trimmedSong(100, 10, 60), 10, 60},
		{"only start", trimmedSong(100, 10, 100), 10, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, e := trimBounds(tt.e)
			if s != tt.wantStart || e != tt.wantEnd {
				t.Errorf("trimBounds = (%v, %v), want (%v, %v)", s, e, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestNextToAnalyseReturnsNewestFirstAndSkipsGoneSongs(t *testing.T) {
	now := time.Now()
	m := testManager(
		song("old", true, now.Add(-2*time.Hour)),
		song("new", true, now),
		song("mid", true, now.Add(-time.Hour)),
	)
	m.queueAnalysis("old", "gone", "new", "mid")

	var order []string
	for {
		id, ok := m.nextToAnalyse()
		if !ok {
			break
		}
		order = append(order, id)
	}

	want := []string{"new", "mid", "old"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v", order, want)
		}
	}
}

func TestAnalysisStatusCountsProgressAndResetsWhenEmpty(t *testing.T) {
	now := time.Now()
	m := testManager(song("a", true, now), song("b", true, now.Add(-time.Hour)))

	if s := m.AnalysisStatus(); s != (AnalysisProgress{}) {
		t.Fatalf("initial status %+v", s)
	}

	m.queueAnalysis("a", "b", "a")
	if s := m.AnalysisStatus(); s.Total != 2 || s.Done != 0 {
		t.Fatalf("after queueing: %+v, want 0/2 (duplicates ignored)", s)
	}

	if _, ok := m.nextToAnalyse(); !ok {
		t.Fatal("expected a song")
	}
	m.finishAnalysis()
	if s := m.AnalysisStatus(); s.Done != 1 || s.Total != 2 {
		t.Fatalf("after one song: %+v, want 1/2", s)
	}

	if _, ok := m.nextToAnalyse(); !ok {
		t.Fatal("expected a second song")
	}
	m.finishAnalysis()
	if s := m.AnalysisStatus(); s.Done != 2 || s.Total != 2 {
		t.Fatalf("after both songs: %+v, want 2/2", s)
	}

	if _, ok := m.nextToAnalyse(); ok {
		t.Fatal("queue should be empty")
	}
	if s := m.AnalysisStatus(); s.Done != 0 || s.Total != 0 {
		t.Fatalf("counters should reset once the queue is empty: %+v", s)
	}
}

func TestQueueAnalysisDoesNothingWhenDisabled(t *testing.T) {
	m := testManager(song("a", true, time.Now()))
	m.cfg.Analysis.Disabled = true

	m.queueAnalysis("a")

	if s := m.AnalysisStatus(); s.Total != 0 {
		t.Errorf("status %+v, want nothing queued", s)
	}
	if _, ok := m.nextToAnalyse(); ok {
		t.Error("a disabled analysis must not queue songs")
	}
}

func TestStartAnalysisWithFFmpeg(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not installed")
	}
	t.Chdir(t.TempDir())

	var cfg config.Config
	cfg.Alternatives.FFmpeg = ffmpeg
	m, err := NewManager(cfg)
	if err != nil {
		t.Fatal(err)
	}

	e := music.Entry{ID: "tone", Added: time.Now()}
	e.FileData.Filename = "tone.wav"
	e.MusicData.Duration = 12
	if err := os.MkdirAll(filepath.Dir(e.AudioPath()), 0o755); err != nil {
		t.Fatal(err)
	}
	gen := exec.Command(ffmpeg, "-v", "error", "-f", "lavfi", "-i", "sine=frequency=440:duration=12", "-ar", "44100", e.AudioPath())
	if out, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("generating audio: %v: %s", err, out)
	}

	m.SongsLock.Lock()
	err = m.Add(&e)
	m.SongsLock.Unlock()
	if err != nil {
		t.Fatal(err)
	}

	events, cancelSub := m.Subscribe()
	defer cancelSub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.StartAnalysis(ctx)

	timeout := time.After(60 * time.Second)
	updated := false
	for !updated {
		select {
		case ev, ok := <-events:
			if !ok {
				t.Fatal("event stream closed")
			}
			if u, isUpdate := ev.(SongUpdated); isUpdate && u.Song.ID == "tone" {
				updated = true
				if !u.Song.Analysis.Usable() {
					t.Errorf("published song has no usable analysis: %+v", u.Song.Analysis)
				}
			}
		case <-timeout:
			t.Fatal("timed out waiting for the SongUpdated event")
		}
	}

	got, _ := m.GetEntry("tone")
	if !got.Analysis.Usable() {
		t.Fatalf("stored analysis is not usable: %+v", got.Analysis)
	}
	if needsAnalysis(got) {
		t.Error("analysed song still needs analysis")
	}

	deadline := time.Now().Add(30 * time.Second)
	for {
		var saved struct {
			Songs map[string]music.Entry `json:"songs"`
		}
		if b, err := os.ReadFile(managerDataFile); err == nil && json.Unmarshal(b, &saved) == nil && saved.Songs["tone"].Analysis.Usable() {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("analysis was never saved to disk")
		}
		time.Sleep(50 * time.Millisecond)
	}
}
