package music

import "time"

// Analysis holds audio descriptors computed from a song's (trimmed) audio.
// It is derived data: it may be missing, outdated (Version) or failed (Error),
// and everything using it must cope with that.
type Analysis struct {
	Version    int       `json:"version"`
	AnalyzedAt time.Time `json:"analyzed_at"`
	// Start and End are the trim bounds the analysis was computed for.
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Error string  `json:"error,omitempty"`

	// Key is the tonic pitch class (0 = C … 11 = B).
	Key         int     `json:"key"`
	Minor       bool    `json:"minor"`
	KeyStrength float64 `json:"key_strength"`
	Camelot     string  `json:"camelot"`

	BPM          float64 `json:"bpm"`
	BeatStrength float64 `json:"beat_strength"`

	LoudnessLUFS  float64 `json:"loudness_lufs"`
	LoudnessRange float64 `json:"loudness_range"`

	EnergyDB       float64 `json:"energy_db"`
	EnergyDBStdDev float64 `json:"energy_db_stddev"`
	OnsetRate      float64 `json:"onset_rate"`

	Centroid float64 `json:"centroid"`
	Rolloff  float64 `json:"rolloff"`
	Flatness float64 `json:"flatness"`

	// Timbre holds the means then the standard deviations of 13 MFCCs.
	Timbre []float64 `json:"timbre"`
}

// Usable reports whether a holds successfully computed descriptors.
func (a *Analysis) Usable() bool {
	return a != nil && a.Error == "" && len(a.Timbre) > 0
}
