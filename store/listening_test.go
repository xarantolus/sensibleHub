package store

import (
	"math"
	"testing"
	"time"
	"xarantolus/sensibleHub/store/analysis"
	"xarantolus/sensibleHub/store/music"
)

func TestSkipFactor(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	if f := skipFactor(nil, now); f != 1 {
		t.Errorf("never played: %v, want neutral 1", f)
	}

	often := &music.Listening{Skips: 6, Updated: now, LastSkipped: now.Add(-48 * time.Hour)}
	loved := &music.Listening{Plays: 6, Updated: now, LastPlayed: now}
	if fo, fl := skipFactor(often, now), skipFactor(loved, now); !(fo < 0.3 && fl == 1) {
		t.Errorf("often skipped %v should be far below a song always played through %v", fo, fl)
	}

	just := &music.Listening{Skips: 1, Updated: now, LastSkipped: now.Add(-time.Hour)}
	dayAgo := &music.Listening{Skips: 1, Updated: now, LastSkipped: now.Add(-25 * time.Hour)}
	if fj, fd := skipFactor(just, now), skipFactor(dayAgo, now); !(fj < fd) {
		t.Errorf("a skip an hour ago (%v) should weigh more than one a day ago (%v)", fj, fd)
	}

	if f := skipFactor(&music.Listening{Skips: 1000, Updated: now, LastSkipped: now}, now); f <= 0 {
		t.Errorf("even a hated song must stay possible, got %v", f)
	}
}

func TestOldSkipsFade(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	fresh := &music.Listening{Skips: 4, Updated: now, LastSkipped: now.Add(-48 * time.Hour)}
	old := &music.Listening{Skips: 4, Updated: now.Add(-120 * 24 * time.Hour), LastSkipped: now.Add(-120 * 24 * time.Hour)}
	// Four half-lives leave a quarter skip: 1 / (1 + 2·0.25).
	if fo, ff := skipFactor(old, now), skipFactor(fresh, now); math.Abs(fo-1/1.5) > 0.01 || ff > 0.2 {
		t.Errorf("four skips should fade to a quarter over four months: old %v, fresh %v", fo, ff)
	}
}

func TestRecordPlays(t *testing.T) {
	now := time.Now()
	m := testManager(song("a", true, now), song("b", true, now))
	t.Chdir(t.TempDir())

	n, err := m.RecordPlays([]PlayEvent{
		{SongID: "a", At: now, Skipped: true},
		{SongID: "a", At: now.Add(time.Second), Listened: 200},
		{SongID: "gone", At: now, Skipped: true},
		{SongID: "b", At: now.Add(time.Hour)},
	})
	if err != nil || n != 3 {
		t.Fatalf("stored %d, %v; want 3 events (unknown song ignored)", n, err)
	}

	a := m.Songs["a"].Listening
	if a == nil || a.Skips < 0.99 || a.Plays < 0.99 || a.LastSkipped.IsZero() || a.LastPlayed.IsZero() {
		t.Fatalf("a: %+v", a)
	}
	if b := m.Songs["b"].Listening; b == nil || !b.LastPlayed.Before(now.Add(2*time.Minute)) {
		t.Fatalf("an event from the future must be clamped to now, got %+v", b)
	}
}

func TestNextSongsAvoidsSkippedSongs(t *testing.T) {
	now := time.Now()
	skipped := song("skipped", true, now)
	skipped.Listening = &music.Listening{Skips: 5, Updated: now, LastSkipped: now}
	m := testManager(skipped, song("liked", true, now))

	counts := map[string]int{}
	for range 400 {
		counts[m.NextSongs("", 1, nil)[0].ID]++
	}
	if counts["liked"] < 8*counts["skipped"] {
		t.Fatalf("a song skipped five times just now should rarely come up: %v", counts)
	}
}

func TestNextSongsChainsFromEachPick(t *testing.T) {
	now := time.Now()
	withKey := func(id string, tonic int) music.Entry {
		e := song(id, true, now)
		e.Analysis = &music.Analysis{Version: analysis.Version, Key: tonic, KeyStrength: 1, BPM: 120, BeatStrength: 1, Timbre: []float64{1, 2}}
		return e
	}
	// C → G → D: each step is a fifth. F♯ clashes with everything else.
	m := testManager(withKey("c", 0), withKey("g", 7), withKey("d", 2), withKey("fs", 6))

	chained := 0
	for range 300 {
		picks := m.NextSongs("c", 2, nil)
		if len(picks) == 2 && picks[0].ID == "g" && picks[1].ID == "d" {
			chained++
		}
	}
	if chained < 60 {
		t.Fatalf("the second pick should follow the first (g→d), seen %d/300 times", chained)
	}
}
