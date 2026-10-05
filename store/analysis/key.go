package analysis

import (
	"fmt"
	"math"
)

// Krumhansl–Kessler key profiles, starting at the tonic.
var (
	majorProfile = [12]float64{6.35, 2.23, 3.48, 2.33, 4.38, 4.09, 2.52, 5.19, 2.39, 3.66, 2.29, 2.88}
	minorProfile = [12]float64{6.33, 2.68, 3.52, 5.38, 2.60, 3.53, 2.54, 4.75, 3.98, 2.69, 3.34, 3.17}
)

// detectKey correlates a pitch-class profile with all 24 rotated key profiles.
// strength is the winning correlation, clamped to [0, 1].
func detectKey(chroma [12]float64) (tonic int, minor bool, strength float64) {
	best := math.Inf(-1)
	for t := range 12 {
		for _, m := range []bool{false, true} {
			profile := &majorProfile
			if m {
				profile = &minorProfile
			}
			var rotated [12]float64
			for i := range 12 {
				rotated[(i+t)%12] = profile[i]
			}
			if r := pearson(chroma[:], rotated[:]); r > best {
				best, tonic, minor = r, t, m
			}
		}
	}
	if math.IsNaN(best) || math.IsInf(best, 0) {
		return 0, false, 0
	}
	return tonic, minor, math.Max(0, math.Min(1, best))
}

func pearson(a, b []float64) float64 {
	var ma, mb float64
	for i := range a {
		ma += a[i]
		mb += b[i]
	}
	ma /= float64(len(a))
	mb /= float64(len(b))
	var cov, va, vb float64
	for i := range a {
		da, db := a[i]-ma, b[i]-mb
		cov += da * db
		va += da * da
		vb += db * db
	}
	if va == 0 || vb == 0 {
		return 0
	}
	return cov / math.Sqrt(va*vb)
}

// Camelot returns the Camelot wheel code of a key, e.g. 8B for C major and 8A for A minor.
func Camelot(tonic int, minor bool) string {
	major := tonic
	letter := "B"
	if minor {
		major = (tonic + 3) % 12
		letter = "A"
	}
	return fmt.Sprintf("%d%s", (major*7%12+7)%12+1, letter)
}

var keyNames = [12]string{"C", "C♯", "D", "E♭", "E", "F", "F♯", "G", "A♭", "A", "B♭", "B"}

// KeyName returns a readable key such as "A minor".
func KeyName(tonic int, minor bool) string {
	if minor {
		return keyNames[tonic%12] + " minor"
	}
	return keyNames[tonic%12] + " major"
}

// KeyCompatibility rates how consonant a transition between two keys is on the
// Camelot wheel: 1 for the same key, 0.8 for neighbours (a fifth apart or the
// relative major/minor), 0.4 for two steps or a diagonal, else 0.
func KeyCompatibility(tonicA int, minorA bool, tonicB int, minorB bool) float64 {
	na, nb := camelotNumber(tonicA, minorA), camelotNumber(tonicB, minorB)
	d := min((na-nb+12)%12, (nb-na+12)%12)
	sameMode := minorA == minorB
	switch {
	case d == 0 && sameMode:
		return 1
	case d == 0 || (d == 1 && sameMode):
		return 0.8
	case d == 1 || (d == 2 && sameMode):
		return 0.4
	}
	return 0
}

func camelotNumber(tonic int, minor bool) int {
	if minor {
		tonic = (tonic + 3) % 12
	}
	return tonic * 7 % 12
}
