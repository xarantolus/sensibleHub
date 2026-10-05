package analysis

import (
	"math"
	"xarantolus/sensibleHub/store/music"
)

// TimbreStats normalises timbre vectors across a library, so that no single
// coefficient with a large range dominates the similarity.
type TimbreStats struct {
	mean, std []float64
}

func NewTimbreStats(analyses []*music.Analysis) TimbreStats {
	var s TimbreStats
	var n float64
	for _, a := range analyses {
		if !a.Usable() {
			continue
		}
		if s.mean == nil {
			s.mean = make([]float64, len(a.Timbre))
			s.std = make([]float64, len(a.Timbre))
		}
		if len(a.Timbre) != len(s.mean) {
			continue
		}
		n++
		for i, v := range a.Timbre {
			s.mean[i] += v
			s.std[i] += v * v
		}
	}
	for i := range s.mean {
		s.mean[i] /= n
		s.std[i] = math.Sqrt(math.Max(s.std[i]/n-s.mean[i]*s.mean[i], 1e-9))
	}
	return s
}

func (s TimbreStats) similarity(a, b []float64) (float64, bool) {
	if len(s.mean) == 0 || len(a) != len(s.mean) || len(b) != len(s.mean) {
		return 0, false
	}
	var dot, na, nb float64
	for i := range a {
		za := (a[i] - s.mean[i]) / s.std[i]
		zb := (b[i] - s.mean[i]) / s.std[i]
		dot += za * zb
		na += za * za
		nb += zb * zb
	}
	if na == 0 || nb == 0 {
		return 0, false
	}
	return dot / math.Sqrt(na*nb), true
}

// Transition returns multiplicative factors (1 is neutral) rating how well b
// follows a: consonant keys, matching tempo, a smooth energy level and similar timbre.
// Weak or ambiguous measurements move their factor less.
func Transition(a, b *music.Analysis, stats TimbreStats) map[string]float64 {
	if !a.Usable() || !b.Usable() {
		return nil
	}
	f := make(map[string]float64, 4)

	keyConfidence := math.Min(a.KeyStrength, b.KeyStrength)
	compat := KeyCompatibility(a.Key, a.Minor, b.Key, b.Minor)
	f["key"] = 1 + 0.6*keyConfidence*(2*compat-1)

	beatConfidence := math.Min(a.BeatStrength, b.BeatStrength)
	f["tempo"] = 1 + 0.5*beatConfidence*(2*TempoSimilarity(a.BPM, b.BPM)-1)

	dEnergy := b.EnergyDB - a.EnergyDB
	f["energy"] = 0.6 + 0.8*math.Exp(-math.Pow(dEnergy/6, 2))

	if sim, ok := stats.similarity(a.Timbre, b.Timbre); ok {
		f["timbre"] = math.Exp(0.8 * sim)
	}
	return f
}
