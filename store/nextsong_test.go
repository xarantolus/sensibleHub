package store

import (
	"sync"
	"testing"
	"time"
	"xarantolus/sensibleHub/store/music"
)

func song(id string, sync bool, added time.Time) music.Entry {
	e := music.Entry{ID: id, Added: added}
	e.SyncSettings.Should = sync
	return e
}

func testManager(songs ...music.Entry) *Manager {
	m := &Manager{Songs: map[string]music.Entry{}, SongsLock: new(sync.RWMutex)}
	for _, s := range songs {
		m.Songs[s.ID] = s
	}
	return m
}

func TestNextSongsOnlySuggestsSyncedSongsOtherThanCurrent(t *testing.T) {
	now := time.Now()
	m := testManager(song("a", true, now), song("b", false, now), song("c", true, now), song("d", true, now))

	for range 50 {
		for _, s := range m.NextSongs("a", 10, nil) {
			if s.ID == "a" || s.ID == "b" {
				t.Fatalf("suggested %q", s.ID)
			}
		}
	}
	if got := len(m.NextSongs("a", 10, nil)); got != 2 {
		t.Fatalf("got %d suggestions, want the 2 other synced songs", got)
	}
}

func TestNextSongsAvoidsExcludedUnlessNothingElseIsLeft(t *testing.T) {
	now := time.Now()
	m := testManager(song("a", true, now), song("b", true, now), song("c", true, now))

	for range 50 {
		got := m.NextSongs("", 1, []string{"a", "b"})
		if len(got) != 1 || got[0].ID != "c" {
			t.Fatalf("got %+v, want only c", got)
		}
	}

	if got := m.NextSongs("", 3, []string{"a", "b", "c"}); len(got) != 3 {
		t.Fatalf("with everything excluded got %d songs, want the whole pool", len(got))
	}
}

func TestNextSongsWithoutSyncedSongsIsEmpty(t *testing.T) {
	m := testManager(song("a", false, time.Now()))
	if got := m.NextSongs("", 5, nil); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestRecencyFavoursNewestAndFades(t *testing.T) {
	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var songs []music.Entry
	for i := range 80 {
		songs = append(songs, song(string(rune('A'+i)), true, base.Add(time.Duration(i)*time.Hour)))
	}
	w := recencyWeights(songs)

	newest, tenth, old := w[songs[79].ID], w[songs[69].ID], w[songs[0].ID]
	if newest != recentBoostMax {
		t.Errorf("newest weight %v, want %v", newest, recentBoostMax)
	}
	if !(newest > tenth && tenth > 1) {
		t.Errorf("weights should fade: newest %v, 10th %v", newest, tenth)
	}
	if old != 1 {
		t.Errorf("old song weight %v, want 1", old)
	}
}

func TestWeightedSamplePrefersHeavierItems(t *testing.T) {
	items := []Suggestion{{ID: "light", Score: 1}, {ID: "heavy", Score: 9}}
	counts := map[string]int{}
	seq := 0.0
	rnd := func() float64 { seq += 0.0137; return seq - float64(int(seq)) }
	for range 2000 {
		counts[weightedSample(items, 1, rnd)[0].ID]++
	}
	if ratio := float64(counts["heavy"]) / float64(counts["light"]); ratio < 6 || ratio > 12 {
		t.Fatalf("heavy/light ratio %.1f, want about 9 (%v)", ratio, counts)
	}
}

func TestWeightedSampleNeverRepeats(t *testing.T) {
	items := []Suggestion{{ID: "a", Score: 1}, {ID: "b", Score: 100}, {ID: "c", Score: 0}}
	got := weightedSample(items, 5, func() float64 { return 0.5 })
	seen := map[string]bool{}
	for _, s := range got {
		if seen[s.ID] {
			t.Fatalf("repeated %q in %+v", s.ID, got)
		}
		seen[s.ID] = true
	}
	if len(got) != 3 {
		t.Fatalf("got %d, want all 3", len(got))
	}
}
