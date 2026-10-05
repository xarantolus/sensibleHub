package analysis

import "math"

const (
	framesPerSecond = float64(SampleRate) / hopSize

	minBPM       = 50
	maxBPM       = 220
	preferredBPM = 120
)

// onsetEnvelope removes the slowly varying part of the spectral flux so that only
// note onsets remain, then half-wave rectifies it.
func onsetEnvelope(flux []float64) []float64 {
	const half = 8
	out := make([]float64, len(flux))
	var sum float64
	for i := range flux {
		if i < len(flux) {
			sum += flux[min(i+half, len(flux)-1)]
		}
		if i > half {
			sum -= flux[i-half-1]
		}
		lo, hi := max(0, i-half), min(len(flux)-1, i+half)
		mean := sum / float64(hi-lo+1)
		out[i] = math.Max(0, flux[i]-mean)
	}
	return out
}

// detectTempo finds the beat period by autocorrelating the onset envelope and
// scoring each candidate tempo together with its multiples, weighted by a broad
// prior around 120 BPM (one octave wide on a log scale).
func detectTempo(env []float64) (bpm, strength float64) {
	maxLag := int(math.Ceil(framesPerSecond*60/minBPM)) * 4
	if len(env) < maxLag*2 {
		return 0, 0
	}
	ac := make([]float64, maxLag+2)
	for lag := range ac {
		var s float64
		for t := 0; t+lag < len(env); t++ {
			s += env[t] * env[t+lag]
		}
		ac[lag] = s / float64(len(env)-lag)
	}
	if ac[0] == 0 {
		return 0, 0
	}

	interp := func(lag float64) float64 {
		i := int(lag)
		if i+1 >= len(ac) {
			return 0
		}
		f := lag - float64(i)
		return ac[i]*(1-f) + ac[i+1]*f
	}

	bestScore := math.Inf(-1)
	for b := float64(minBPM); b <= maxBPM; b += 0.1 {
		period := framesPerSecond * 60 / b
		var s float64
		for m := 1; m <= 4; m++ {
			s += interp(period*float64(m)) / float64(m)
		}
		prior := math.Exp(-0.5 * math.Pow(math.Log2(b/preferredBPM), 2))
		if score := s * prior; score > bestScore {
			bestScore, bpm = score, b
		}
	}
	strength = math.Max(0, math.Min(1, interp(framesPerSecond*60/bpm)/ac[0]))
	return math.Round(bpm*10) / 10, strength
}

// onsetRate counts clear onsets per second: local maxima above mean + 1 stddev,
// at least 50 ms apart.
func onsetRate(env []float64) float64 {
	if len(env) < 3 {
		return 0
	}
	var sum, sumSq float64
	for _, v := range env {
		sum += v
		sumSq += v * v
	}
	n := float64(len(env))
	threshold := sum/n + stddev(sum, sumSq, n)
	minGap := int(math.Round(0.05 * framesPerSecond))

	count, last := 0, -minGap
	for i := 1; i < len(env)-1; i++ {
		if env[i] > threshold && env[i] >= env[i-1] && env[i] > env[i+1] && i-last >= minGap {
			count++
			last = i
		}
	}
	return float64(count) / (n / framesPerSecond)
}

// TempoSimilarity is 1 for equal tempos and falls off with the relative difference,
// treating half and double time as equal (90 BPM fits 180 BPM).
func TempoSimilarity(a, b float64) float64 {
	if a <= 0 || b <= 0 {
		return 0
	}
	d := math.Inf(1)
	for _, r := range []float64{0.5, 1, 2} {
		d = math.Min(d, math.Abs(math.Log2(b*r/a)))
	}
	return math.Exp(-math.Pow(d/0.08, 2))
}
