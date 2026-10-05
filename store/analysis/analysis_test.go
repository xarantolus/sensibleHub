package analysis

import (
	"context"
	"math"
	"math/rand/v2"
	"os/exec"
	"path/filepath"
	"testing"
)

func chord(seconds float64, freqs ...float64) []float32 {
	n := int(seconds * SampleRate)
	out := make([]float32, n)
	for i := range out {
		t := float64(i) / SampleRate
		var v float64
		for _, f := range freqs {
			for h, amp := range []float64{1, 0.5, 0.25} {
				v += amp * math.Sin(2*math.Pi*f*float64(h+1)*t)
			}
		}
		out[i] = float32(0.1 * v)
	}
	return out
}

func clicks(seconds, bpm float64) []float32 {
	n := int(seconds * SampleRate)
	out := make([]float32, n)
	period := int(SampleRate * 60 / bpm)
	r := rand.New(rand.NewPCG(1, 2))
	for start := 0; start < n; start += period {
		for i := 0; i < 400 && start+i < n; i++ {
			out[start+i] = float32((r.Float64()*2 - 1) * math.Exp(-float64(i)/80))
		}
	}
	return out
}

func analyze(t *testing.T, samples []float32) Features {
	t.Helper()
	a := NewAnalyzer()
	for len(samples) > 0 {
		n := min(len(samples), 3000)
		a.Write(samples[:n])
		samples = samples[n:]
	}
	f, ok := a.Features()
	if !ok {
		t.Fatal("no features for non-silent audio")
	}
	return f
}

func TestDetectsMajorKey(t *testing.T) {
	f := analyze(t, chord(8, 261.63, 329.63, 392.00))
	if f.Key != 0 || f.Minor {
		t.Fatalf("C major triad detected as %s", KeyName(f.Key, f.Minor))
	}
	if got := Camelot(f.Key, f.Minor); got != "8B" {
		t.Errorf("Camelot %s, want 8B", got)
	}
}

func TestDetectsMinorKey(t *testing.T) {
	f := analyze(t, chord(8, 220.00, 261.63, 329.63))
	if f.Key != 9 || !f.Minor {
		t.Fatalf("A minor triad detected as %s", KeyName(f.Key, f.Minor))
	}
	if got := Camelot(f.Key, f.Minor); got != "8A" {
		t.Errorf("Camelot %s, want 8A", got)
	}
}

func TestDetectsTempo(t *testing.T) {
	for _, bpm := range []float64{90, 120, 128, 150} {
		f := analyze(t, clicks(30, bpm))
		if math.Abs(f.BPM-bpm) > 1 {
			t.Errorf("click track at %v BPM measured %v", bpm, f.BPM)
		}
		if f.BeatStrength < 0.3 {
			t.Errorf("click track at %v BPM has beat strength %v", bpm, f.BeatStrength)
		}
	}
}

func TestLouderAudioHasMoreEnergy(t *testing.T) {
	quiet := analyze(t, chord(4, 440))
	loud := chord(4, 440)
	for i := range loud {
		loud[i] *= 4
	}
	got := analyze(t, loud)
	if d := got.EnergyDB - quiet.EnergyDB; math.Abs(d-12.04) > 0.5 {
		t.Errorf("4x amplitude should add ~12 dB, added %.2f", d)
	}
}

func TestSilenceHasNoFeatures(t *testing.T) {
	a := NewAnalyzer()
	a.Write(make([]float32, SampleRate*5))
	if _, ok := a.Features(); ok {
		t.Fatal("silence produced features")
	}
}

func TestCamelotWheel(t *testing.T) {
	cases := []struct {
		tonic int
		minor bool
		want  string
	}{
		{0, false, "8B"}, {7, false, "9B"}, {5, false, "7B"}, {11, false, "1B"},
		{9, true, "8A"}, {4, true, "9A"}, {2, true, "7A"}, {0, true, "5A"},
	}
	for _, c := range cases {
		if got := Camelot(c.tonic, c.minor); got != c.want {
			t.Errorf("%s: %s, want %s", KeyName(c.tonic, c.minor), got, c.want)
		}
	}
}

func TestKeyCompatibility(t *testing.T) {
	const C, G, D, A, E = 0, 7, 2, 9, 4
	cases := []struct {
		name   string
		ta     int
		ma     bool
		tb     int
		mb     bool
		expect float64
	}{
		{"same key", C, false, C, false, 1},
		{"fifth up", C, false, G, false, 0.8},
		{"relative minor", C, false, A, true, 0.8},
		{"two fifths", C, false, D, false, 0.4},
		{"diagonal", C, false, E, true, 0.4},
		{"far", C, false, 6, false, 0},
	}
	for _, c := range cases {
		if got := KeyCompatibility(c.ta, c.ma, c.tb, c.mb); got != c.expect {
			t.Errorf("%s: %v, want %v", c.name, got, c.expect)
		}
		if got := KeyCompatibility(c.tb, c.mb, c.ta, c.ma); got != c.expect {
			t.Errorf("%s reversed: %v, want %v", c.name, got, c.expect)
		}
	}
}

func TestTempoSimilarityTreatsHalfAndDoubleTimeAsEqual(t *testing.T) {
	if s := TempoSimilarity(90, 180); s < 0.99 {
		t.Errorf("90 vs 180: %v", s)
	}
	if s := TempoSimilarity(120, 122); s < 0.9 {
		t.Errorf("120 vs 122: %v", s)
	}
	if s := TempoSimilarity(120, 150); s > 0.1 {
		t.Errorf("120 vs 150: %v", s)
	}
	if s := TempoSimilarity(0, 120); s != 0 {
		t.Errorf("unknown tempo: %v", s)
	}
}

func TestParseLoudness(t *testing.T) {
	out := []byte(`[Parsed_ebur128_0 @ 0x1] t: 1.0 M: -20.1 S:-120.7 I: -20.1 LUFS LRA: 0.0 LU
[Parsed_ebur128_0 @ 0x1] Summary:

  Integrated loudness:
    I:         -14.3 LUFS
    Threshold: -24.4 LUFS

  Loudness range:
    LRA:         6.2 LU
`)
	lufs, lra, err := parseLoudness(out)
	if err != nil || lufs != -14.3 || lra != 6.2 {
		t.Fatalf("got %v %v %v", lufs, lra, err)
	}
	if _, _, err := parseLoudness([]byte("    I:         -inf LUFS\n    LRA:         0.0 LU")); err == nil {
		t.Fatal("silent audio should be an error")
	}
}

func TestAnalyzeWithFFmpeg(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not installed")
	}
	path := filepath.Join(t.TempDir(), "tone.wav")
	gen := exec.Command(ffmpeg, "-v", "error", "-f", "lavfi", "-i", "sine=frequency=440:duration=12", "-ar", "44100", path)
	if out, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("generating audio: %v: %s", err, out)
	}

	a, err := Analyze(context.Background(), ffmpeg, path, 2, 10)
	if err != nil {
		t.Fatal(err)
	}
	if a.Key != 9 {
		t.Errorf("a 440 Hz tone should be in A, got %s", KeyName(a.Key, a.Minor))
	}
	if a.LoudnessLUFS > -10 || a.LoudnessLUFS < -40 {
		t.Errorf("implausible loudness %v", a.LoudnessLUFS)
	}
	if len(a.Timbre) != 2*mfccs || a.Version != Version || a.Start != 2 || a.End != 10 {
		t.Errorf("incomplete result %+v", a)
	}
}
