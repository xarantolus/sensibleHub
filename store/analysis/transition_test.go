package analysis

import (
	"math"
	"testing"
	"xarantolus/sensibleHub/store/music"
)

func tonal(key int, minor bool, bpm, strength float64) *music.Analysis {
	return &music.Analysis{
		Version: Version, Key: key, Minor: minor, KeyStrength: strength,
		BPM: bpm, BeatStrength: strength, Timbre: []float64{1, 2},
	}
}

func TestTransitionPrefersSameKeyAndTempo(t *testing.T) {
	cur := tonal(0, false, 120, 1)
	same := tonal(0, false, 120, 1)
	clashingKey := tonal(6, false, 120, 1)
	clashingTempo := tonal(0, false, 170, 1)

	fSame := Transition(cur, same, TimbreStats{})
	fKey := Transition(cur, clashingKey, TimbreStats{})
	fTempo := Transition(cur, clashingTempo, TimbreStats{})

	if fSame["key"] <= 1 || fKey["key"] >= 1 {
		t.Errorf("key factors: same %v, clashing %v; want boost and penalty", fSame["key"], fKey["key"])
	}
	if fSame["tempo"] <= 1 || fTempo["tempo"] >= 1 {
		t.Errorf("tempo factors: same %v, clashing %v; want boost and penalty", fSame["tempo"], fTempo["tempo"])
	}
	if fSame["tempo"] != fKey["tempo"] || fSame["key"] != fTempo["key"] {
		t.Error("key and tempo should be rated independently")
	}
}

func TestTransitionTreatsHalfTimeAsMatchingTempo(t *testing.T) {
	cur := tonal(0, false, 90, 1)
	if f := Transition(cur, tonal(0, false, 180, 1), TimbreStats{}); f["tempo"] <= 1 {
		t.Errorf("double time should still boost, got %v", f["tempo"])
	}
}

func TestTransitionWithoutUsableAnalysisHasNoFactors(t *testing.T) {
	good := tonal(0, false, 120, 1)
	failed := tonal(0, false, 120, 1)
	failed.Error = "broken"
	noTimbre := tonal(0, false, 120, 1)
	noTimbre.Timbre = nil

	for name, other := range map[string]*music.Analysis{"nil": nil, "failed": failed, "no timbre": noTimbre} {
		if f := Transition(good, other, TimbreStats{}); len(f) != 0 {
			t.Errorf("%s as target: got factors %v", name, f)
		}
		if f := Transition(other, good, TimbreStats{}); len(f) != 0 {
			t.Errorf("%s as source: got factors %v", name, f)
		}
	}
}

func TestTransitionWeakMeasurementsMoveFactorsLess(t *testing.T) {
	strongMatch := Transition(tonal(0, false, 120, 1), tonal(0, false, 120, 1), TimbreStats{})
	weakMatch := Transition(tonal(0, false, 120, 0.2), tonal(0, false, 120, 0.2), TimbreStats{})
	strongClash := Transition(tonal(0, false, 120, 1), tonal(6, false, 170, 1), TimbreStats{})
	weakClash := Transition(tonal(0, false, 120, 0.2), tonal(6, false, 170, 0.2), TimbreStats{})

	for _, name := range []string{"key", "tempo"} {
		if math.Abs(weakMatch[name]-1) >= math.Abs(strongMatch[name]-1) {
			t.Errorf("%s: weak match %v should be closer to 1 than strong %v", name, weakMatch[name], strongMatch[name])
		}
		if math.Abs(weakClash[name]-1) >= math.Abs(strongClash[name]-1) {
			t.Errorf("%s: weak clash %v should be closer to 1 than strong %v", name, weakClash[name], strongClash[name])
		}
	}
}

func TestTransitionUsesWeakerOfTheTwoKeyStrengths(t *testing.T) {
	both := Transition(tonal(0, false, 120, 1), tonal(0, false, 120, 1), TimbreStats{})
	oneWeak := Transition(tonal(0, false, 120, 1), tonal(0, false, 120, 0.1), TimbreStats{})
	if oneWeak["key"] >= both["key"] {
		t.Errorf("one weak side should reduce the boost: %v vs %v", oneWeak["key"], both["key"])
	}
}

func TestTransitionPrefersSmoothEnergy(t *testing.T) {
	quiet := tonal(0, false, 120, 1)
	quiet.EnergyDB = -30
	similar := tonal(0, false, 120, 1)
	similar.EnergyDB = -29
	loud := tonal(0, false, 120, 1)
	loud.EnergyDB = -10

	if a, b := Transition(quiet, similar, TimbreStats{})["energy"], Transition(quiet, loud, TimbreStats{})["energy"]; a <= b {
		t.Errorf("similar energy factor %v should beat a jump %v", a, b)
	}
}

func libraryAnalyses() []*music.Analysis {
	var out []*music.Analysis
	for _, tb := range [][]float64{{0, 10}, {10, 0}, {5, 5}, {2, 8}} {
		a := tonal(0, false, 120, 1)
		a.Timbre = tb
		out = append(out, a)
	}
	return out
}

func TestTimbreFactorRewardsSimilarSounds(t *testing.T) {
	lib := libraryAnalyses()
	stats := NewTimbreStats(lib)

	cur, close, far := tonal(0, false, 120, 1), tonal(0, false, 120, 1), tonal(0, false, 120, 1)
	cur.Timbre = []float64{0, 10}
	close.Timbre = []float64{1, 9}
	far.Timbre = []float64{10, 0}

	fClose := Transition(cur, close, stats)["timbre"]
	fFar := Transition(cur, far, stats)["timbre"]
	if fClose == 0 || fFar == 0 {
		t.Fatalf("timbre factor missing: %v %v", fClose, fFar)
	}
	if fClose <= 1 || fFar >= 1 || fClose <= fFar {
		t.Errorf("similar %v should boost, dissimilar %v should penalise", fClose, fFar)
	}
}

func TestTimbreFactorMissingWithoutStatsOrMatchingLength(t *testing.T) {
	a, b := tonal(0, false, 120, 1), tonal(0, false, 120, 1)
	if _, ok := Transition(a, b, TimbreStats{})["timbre"]; ok {
		t.Error("empty stats should give no timbre factor")
	}

	stats := NewTimbreStats(libraryAnalyses())
	b.Timbre = []float64{1, 2, 3}
	if _, ok := Transition(a, b, stats)["timbre"]; ok {
		t.Error("mismatched timbre length should give no timbre factor")
	}
}

func TestNewTimbreStatsIgnoresUnusableAnalyses(t *testing.T) {
	failed := tonal(0, false, 120, 1)
	failed.Error = "broken"
	failed.Timbre = []float64{1000, -1000}

	with := NewTimbreStats(append(libraryAnalyses(), failed, nil))
	without := NewTimbreStats(libraryAnalyses())

	a, b := tonal(0, false, 120, 1), tonal(0, false, 120, 1)
	a.Timbre = []float64{0, 10}
	b.Timbre = []float64{10, 0}
	if got, want := Transition(a, b, with)["timbre"], Transition(a, b, without)["timbre"]; math.Abs(got-want) > 1e-12 {
		t.Errorf("unusable analyses changed the stats: %v vs %v", got, want)
	}
}

func TestNewTimbreStatsOfNothingGivesNoTimbreFactor(t *testing.T) {
	stats := NewTimbreStats(nil)
	if _, ok := Transition(tonal(0, false, 120, 1), tonal(0, false, 120, 1), stats)["timbre"]; ok {
		t.Error("expected no timbre factor")
	}
}
