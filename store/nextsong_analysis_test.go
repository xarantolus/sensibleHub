package store

import (
	"testing"
	"time"
	"xarantolus/sensibleHub/store/analysis"
	"xarantolus/sensibleHub/store/music"
)

func analysedSong(id string, key int, bpm float64) music.Entry {
	e := song(id, true, time.Now())
	e.Analysis = &music.Analysis{
		Version: analysis.Version, Key: key, KeyStrength: 1,
		BPM: bpm, BeatStrength: 1, Timbre: []float64{1, 2},
	}
	return e
}

func TestNextSongsFavoursCompatibleKeyAndTempo(t *testing.T) {
	m := testManager(
		analysedSong("cur", 0, 120),
		analysedSong("fits", 0, 120),
		analysedSong("clashes", 6, 170),
	)

	counts := map[string]int{}
	for range 600 {
		counts[m.NextSongs("cur", 1, nil)[0].ID]++
	}
	if counts["fits"] < 4*counts["clashes"] {
		t.Fatalf("fitting song should be picked far more often: %v", counts)
	}
}

func TestNextSongsReportsKeyAndTempoFactors(t *testing.T) {
	m := testManager(analysedSong("cur", 0, 120), analysedSong("fits", 0, 120), analysedSong("clashes", 6, 170))

	got := map[string]Suggestion{}
	for _, s := range m.NextSongs("cur", 2, nil) {
		got[s.ID] = s
	}
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
	if got["fits"].Factors["key"] <= got["clashes"].Factors["key"] || got["fits"].Factors["tempo"] <= got["clashes"].Factors["tempo"] {
		t.Errorf("fitting song should rate higher on key and tempo: %+v vs %+v", got["fits"].Factors, got["clashes"].Factors)
	}
	if got["fits"].Score <= got["clashes"].Score {
		t.Errorf("scores: fits %v, clashes %v", got["fits"].Score, got["clashes"].Score)
	}
}

func TestNextSongsWithoutAnalysisUsesNoMusicalFactors(t *testing.T) {
	plain := song("plain", true, time.Now())
	tests := map[string]*Manager{
		"current not analysed":   testManager(song("cur", true, time.Now()), analysedSong("other", 0, 120)),
		"candidate not analysed": testManager(analysedSong("cur", 0, 120), plain),
	}
	for name, m := range tests {
		for _, s := range m.NextSongs("cur", 5, nil) {
			for _, f := range []string{"key", "tempo", "energy", "timbre"} {
				if _, ok := s.Factors[f]; ok {
					t.Errorf("%s: %s has a %q factor", name, s.ID, f)
				}
			}
			if s.Score <= 0 {
				t.Errorf("%s: %s has score %v and could never be picked", name, s.ID, s.Score)
			}
		}
	}
}

func TestNextSongsIgnoresFailedAnalysis(t *testing.T) {
	broken := analysedSong("broken", 6, 170)
	broken.Analysis.Error = "decode failed"
	m := testManager(analysedSong("cur", 0, 120), broken)

	got := m.NextSongs("cur", 1, nil)
	if len(got) != 1 || got[0].ID != "broken" {
		t.Fatalf("got %+v", got)
	}
	if _, ok := got[0].Factors["key"]; ok {
		t.Error("a failed analysis must not influence the pick")
	}
}
