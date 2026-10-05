// Package analysis computes musical descriptors (key, tempo, energy, timbre)
// from decoded audio, without cgo or external models.
package analysis

import (
	"math"

	"gonum.org/v1/gonum/dsp/fourier"
)

const (
	SampleRate = 22050

	frameSize = 2048
	hopSize   = 256

	chromaFrameSize = 8192
	chromaHopSize   = 4096
	chromaMinHz     = 65
	chromaMaxHz     = 2100

	melBands = 40
	melMinHz = 20
	melMaxHz = 8000
	mfccs    = 13

	silenceDB = -60
)

// Features are the descriptors computed from samples alone; loudness comes from ffmpeg.
type Features struct {
	Key          int
	Minor        bool
	KeyStrength  float64
	BPM          float64
	BeatStrength float64

	EnergyDB       float64
	EnergyDBStdDev float64
	OnsetRate      float64

	Centroid float64
	Rolloff  float64
	Flatness float64

	Timbre []float64
}

// Analyzer consumes mono samples at SampleRate in chunks of any size, so a whole
// song never has to be held in memory.
type Analyzer struct {
	buf      []float64
	consumed int
	nextHop  int
	nextBig  int

	fft, bigFFT       *fourier.FFT
	window, bigWindow []float64
	coeffs            []complex128
	frame             []float64
	mags, prevLogMags []float64
	melFilters        [][]melWeight

	onset  []float64
	chroma [12]float64

	frames                  int
	energySum, energySumSq  float64
	centroidSum, rolloffSum float64
	flatnessSum             float64
	mfccSum, mfccSumSq      [mfccs]float64
}

type melWeight struct {
	bin    int
	weight float64
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		fft:         fourier.NewFFT(frameSize),
		bigFFT:      fourier.NewFFT(chromaFrameSize),
		window:      hann(frameSize),
		bigWindow:   hann(chromaFrameSize),
		frame:       make([]float64, chromaFrameSize),
		mags:        make([]float64, frameSize/2+1),
		prevLogMags: make([]float64, frameSize/2+1),
		melFilters:  melFilterbank(),
		nextHop:     frameSize,
		nextBig:     chromaFrameSize,
	}
}

func hann(n int) []float64 {
	w := make([]float64, n)
	for i := range w {
		w[i] = 0.5 - 0.5*math.Cos(2*math.Pi*float64(i)/float64(n-1))
	}
	return w
}

// Write adds samples. Frames are analysed as soon as enough samples are buffered.
func (a *Analyzer) Write(samples []float32) {
	for _, s := range samples {
		a.buf = append(a.buf, float64(s))
	}
	end := a.consumed + len(a.buf)

	for a.nextHop <= end {
		a.analyzeFrame(a.slice(a.nextHop-frameSize, frameSize))
		a.nextHop += hopSize
	}
	for a.nextBig <= end {
		a.analyzeChroma(a.slice(a.nextBig-chromaFrameSize, chromaFrameSize))
		a.nextBig += chromaHopSize
	}

	keep := chromaFrameSize
	if drop := len(a.buf) - keep; drop > 0 {
		a.buf = append(a.buf[:0], a.buf[drop:]...)
		a.consumed += drop
	}
}

func (a *Analyzer) slice(start, n int) []float64 {
	i := start - a.consumed
	return a.buf[i : i+n]
}

func (a *Analyzer) analyzeFrame(samples []float64) {
	var sumSq float64
	frame := a.frame[:frameSize]
	for i, s := range samples {
		sumSq += s * s
		frame[i] = s * a.window[i]
	}
	rmsDB := 10 * math.Log10(sumSq/frameSize+1e-12)

	a.coeffs = a.fft.Coefficients(a.coeffs, frame)
	var flux, total, weighted, power, logPower float64
	for k, c := range a.coeffs {
		m := math.Hypot(real(c), imag(c))
		a.mags[k] = m
		lm := math.Log1p(10 * m)
		if d := lm - a.prevLogMags[k]; d > 0 {
			flux += d
		}
		a.prevLogMags[k] = lm
		if k > 0 {
			f := float64(k) * SampleRate / frameSize
			total += m
			weighted += f * m
			p := m*m + 1e-12
			power += p
			logPower += math.Log(p)
		}
	}
	a.onset = append(a.onset, flux)

	if rmsDB < silenceDB {
		return
	}
	a.frames++
	a.energySum += rmsDB
	a.energySumSq += rmsDB * rmsDB
	if total > 0 {
		a.centroidSum += weighted / total
	}
	bins := float64(len(a.coeffs) - 1)
	a.flatnessSum += math.Exp(logPower/bins) / (power / bins)
	a.rolloffSum += a.rolloff(power)

	var mel [melBands]float64
	for b, filter := range a.melFilters {
		var e float64
		for _, w := range filter {
			m := a.mags[w.bin]
			e += m * m * w.weight
		}
		mel[b] = math.Log(e + 1e-10)
	}
	for c := range mfccs {
		var v float64
		for b := range melBands {
			v += mel[b] * math.Cos(math.Pi*float64(c)*(float64(b)+0.5)/melBands)
		}
		a.mfccSum[c] += v
		a.mfccSumSq[c] += v * v
	}
}

func (a *Analyzer) rolloff(power float64) float64 {
	threshold := 0.85 * power
	var acc float64
	for k := 1; k < len(a.mags); k++ {
		acc += a.mags[k]*a.mags[k] + 1e-12
		if acc >= threshold {
			return float64(k) * SampleRate / frameSize
		}
	}
	return SampleRate / 2
}

func (a *Analyzer) analyzeChroma(samples []float64) {
	frame := a.frame[:chromaFrameSize]
	var sumSq float64
	for i, s := range samples {
		sumSq += s * s
		frame[i] = s * a.bigWindow[i]
	}
	if 10*math.Log10(sumSq/chromaFrameSize+1e-12) < silenceDB {
		return
	}

	coeffs := a.bigFFT.Coefficients(nil, frame)
	var c [12]float64
	lo := int(chromaMinHz * chromaFrameSize / SampleRate)
	hi := int(chromaMaxHz * chromaFrameSize / SampleRate)
	for k := max(lo, 1); k <= hi && k < len(coeffs); k++ {
		f := float64(k) * SampleRate / chromaFrameSize
		pitch := 12*math.Log2(f/440) + 69
		pc := int(math.Round(pitch)) % 12
		if pc < 0 {
			pc += 12
		}
		c[pc] += math.Hypot(real(coeffs[k]), imag(coeffs[k]))
	}
	var peak float64
	for _, v := range c {
		peak = math.Max(peak, v)
	}
	if peak == 0 {
		return
	}
	for i, v := range c {
		a.chroma[i] += v / peak
	}
}

// Features returns the descriptors of everything written so far.
func (a *Analyzer) Features() (Features, bool) {
	if a.frames == 0 {
		return Features{}, false
	}
	n := float64(a.frames)
	f := Features{
		EnergyDB:       a.energySum / n,
		EnergyDBStdDev: stddev(a.energySum, a.energySumSq, n),
		Centroid:       a.centroidSum / n,
		Rolloff:        a.rolloffSum / n,
		Flatness:       a.flatnessSum / n,
		Timbre:         make([]float64, 2*mfccs),
	}
	for c := range mfccs {
		f.Timbre[c] = a.mfccSum[c] / n
		f.Timbre[mfccs+c] = stddev(a.mfccSum[c], a.mfccSumSq[c], n)
	}

	f.Key, f.Minor, f.KeyStrength = detectKey(a.chroma)
	env := onsetEnvelope(a.onset)
	f.BPM, f.BeatStrength = detectTempo(env)
	f.OnsetRate = onsetRate(env)
	return f, true
}

func stddev(sum, sumSq, n float64) float64 {
	mean := sum / n
	return math.Sqrt(math.Max(sumSq/n-mean*mean, 0))
}

func hzToMel(f float64) float64 { return 2595 * math.Log10(1+f/700) }
func melToHz(m float64) float64 { return 700 * (math.Pow(10, m/2595) - 1) }

func melFilterbank() [][]melWeight {
	lo, hi := hzToMel(melMinHz), hzToMel(melMaxHz)
	edges := make([]float64, melBands+2)
	for i := range edges {
		edges[i] = melToHz(lo + (hi-lo)*float64(i)/float64(melBands+1))
	}
	filters := make([][]melWeight, melBands)
	for b := range melBands {
		left, center, right := edges[b], edges[b+1], edges[b+2]
		for k := 1; k <= frameSize/2; k++ {
			f := float64(k) * SampleRate / frameSize
			var w float64
			switch {
			case f > left && f <= center:
				w = (f - left) / (center - left)
			case f > center && f < right:
				w = (right - f) / (right - center)
			}
			if w > 0 {
				filters[b] = append(filters[b], melWeight{bin: k, weight: w})
			}
		}
	}
	return filters
}
