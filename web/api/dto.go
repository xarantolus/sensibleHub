package api

import (
	"errors"
	"math"
	"time"
	"xarantolus/sensibleHub/store"
	"xarantolus/sensibleHub/store/analysis"
	"xarantolus/sensibleHub/store/music"
)

type Cover struct {
	Size  int    `json:"size" doc:"Width and height in pixels; covers are square"`
	Color string `json:"color,omitempty" doc:"Dominant color as #rrggbb"`
}

type Playback struct {
	Start float64 `json:"start" minimum:"0" doc:"Seconds into the audio file where the song starts"`
	End   float64 `json:"end" minimum:"0" doc:"Seconds into the audio file where the song ends"`
}

type SongSummary struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Artist   string    `json:"artist,omitempty"`
	Album    string    `json:"album,omitempty"`
	Year     *int      `json:"year,omitempty"`
	Duration float64   `json:"duration" doc:"Length of the audio file in seconds"`
	Playback Playback  `json:"playback"`
	Sync     bool      `json:"sync"`
	Added    time.Time `json:"added"`
	LastEdit time.Time `json:"lastEdit"`
	Cover    *Cover    `json:"cover,omitempty"`
	Loudness *float64  `json:"loudness,omitempty" doc:"Integrated loudness in LUFS, once the song is analysed; used to play songs equally loud"`
}

type Analysis struct {
	Status        string    `json:"status" enum:"pending,done,failed"`
	Error         string    `json:"error,omitempty"`
	AnalyzedAt    time.Time `json:"analyzedAt,omitzero"`
	Key           string    `json:"key,omitempty" doc:"e.g. A minor"`
	Camelot       string    `json:"camelot,omitempty" doc:"Camelot wheel code, e.g. 8A"`
	KeyStrength   float64   `json:"keyStrength,omitempty"`
	BPM           float64   `json:"bpm,omitempty"`
	BeatStrength  float64   `json:"beatStrength,omitempty"`
	LoudnessLUFS  float64   `json:"loudnessLufs,omitempty"`
	LoudnessRange float64   `json:"loudnessRange,omitempty"`
	EnergyDB      float64   `json:"energyDb,omitempty"`
	OnsetRate     float64   `json:"onsetRate,omitempty" doc:"Note onsets per second"`
	Centroid      float64   `json:"centroid,omitempty" doc:"Spectral centroid (brightness) in Hz"`
	Flatness      float64   `json:"flatness,omitempty" doc:"Spectral flatness, 0 tonal … 1 noisy"`
}

type AnalysisStatus struct {
	Running bool `json:"running"`
	Done    int  `json:"done"`
	Total   int  `json:"total"`
}

type AudioFile struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type SongDetail struct {
	SongSummary
	SourceURL string    `json:"sourceUrl"`
	Imported  bool      `json:"imported"`
	File      AudioFile `json:"file"`
	Related   []string  `json:"related" doc:"IDs of similar songs"`
	Analysis  Analysis  `json:"analysis"`
	Listening Listening `json:"listening"`
}

type Listening struct {
	Plays       float64   `json:"plays" doc:"Plays, decaying over time (half-life 30 days)"`
	Skips       float64   `json:"skips" doc:"Skips, decaying over time (half-life 30 days)"`
	LastPlayed  time.Time `json:"lastPlayed,omitzero"`
	LastSkipped time.Time `json:"lastSkipped,omitzero"`
}

type Group struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Link        string   `json:"link,omitempty"`
	SongIDs     []string `json:"songIds"`
}

type Album struct {
	Title   string   `json:"title"`
	Artist  string   `json:"artist"`
	SongIDs []string `json:"songIds"`
}

type Artist struct {
	Name      string   `json:"name"`
	PlayTime  float64  `json:"playTime" doc:"Total length of all songs in seconds"`
	YearStart int      `json:"yearStart,omitempty"`
	YearEnd   int      `json:"yearEnd,omitempty"`
	Albums    []Album  `json:"albums"`
	Featured  []string `json:"featured" doc:"IDs of songs by other artists featuring this one"`
}

type DownloadFailure struct {
	Reason  store.DownloadFailure `json:"reason" enum:"aborted,tool_failed,no_audio,invalid_audio,duplicate,internal"`
	Message string                `json:"message"`
	Output  string                `json:"output,omitempty" doc:"Downloader output, for diagnosis"`
	SongID  string                `json:"songId,omitempty" doc:"The existing song when the reason is duplicate"`
}

type DownloadStatus struct {
	Running   bool             `json:"running"`
	URL       string           `json:"url,omitempty" doc:"What the downloader is fetching right now"`
	Queued    int              `json:"queued" doc:"Downloads waiting after the current one"`
	LastError *DownloadFailure `json:"lastError,omitempty"`
}

func songSummary(e music.Entry) SongSummary {
	s := SongSummary{
		ID:       e.ID,
		Title:    e.MusicData.Title,
		Artist:   e.MusicData.Artist,
		Album:    e.MusicData.Album,
		Year:     e.MusicData.Year,
		Duration: e.MusicData.Duration,
		Playback: Playback{Start: max(e.AudioSettings.Start, 0), End: e.AudioSettings.End},
		Sync:     e.SyncSettings.Should,
		Added:    e.Added,
		LastEdit: e.LastEdit,
	}
	if s.Playback.End <= 0 || s.Playback.End > s.Duration {
		s.Playback.End = s.Duration
	}
	if e.PictureData.Filename != "" {
		s.Cover = &Cover{Size: e.PictureData.Size, Color: string(e.PictureData.DominantColorHEX)}
	}
	if a := e.Analysis; a.Usable() {
		lufs := a.LoudnessLUFS
		s.Loudness = &lufs
	}
	return s
}

func songDetail(e music.Entry, related []music.Entry) SongDetail {
	return SongDetail{
		SongSummary: songSummary(e),
		SourceURL:   e.SourceURL,
		Imported:    e.IsImported(),
		File:        AudioFile{Name: e.FileData.Filename, Size: e.FileData.Size},
		Related:     ids(related),
		Analysis:    analysisDTO(e.Analysis),
		Listening:   listeningDTO(e.Listening),
	}
}

func listeningDTO(l *music.Listening) Listening {
	if l == nil {
		return Listening{}
	}
	d := store.ListeningNow(*l)
	round := func(v float64) float64 { return math.Round(v*100) / 100 }
	return Listening{Plays: round(d.Plays), Skips: round(d.Skips), LastPlayed: l.LastPlayed, LastSkipped: l.LastSkipped}
}

func analysisDTO(a *music.Analysis) Analysis {
	switch {
	case a == nil || a.Version < analysis.Version:
		return Analysis{Status: "pending"}
	case !a.Usable():
		return Analysis{Status: "failed", Error: a.Error, AnalyzedAt: a.AnalyzedAt}
	}
	return Analysis{
		Status:        "done",
		AnalyzedAt:    a.AnalyzedAt,
		Key:           analysis.KeyName(a.Key, a.Minor),
		Camelot:       a.Camelot,
		KeyStrength:   a.KeyStrength,
		BPM:           a.BPM,
		BeatStrength:  a.BeatStrength,
		LoudnessLUFS:  a.LoudnessLUFS,
		LoudnessRange: a.LoudnessRange,
		EnergyDB:      a.EnergyDB,
		OnsetRate:     a.OnsetRate,
		Centroid:      a.Centroid,
		Flatness:      a.Flatness,
	}
}

func ids(entries []music.Entry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.ID
	}
	return out
}

func album(a store.Album) Album {
	return Album{Title: a.Title, Artist: a.Artist, SongIDs: ids(a.Songs)}
}

func downloadFailure(err *store.DownloadError) *DownloadFailure {
	if err == nil {
		return nil
	}
	f := &DownloadFailure{Reason: err.Reason, Message: err.Err.Error(), Output: err.Output}
	var dup *store.AlreadyDownloadedError
	if errors.As(err, &dup) {
		f.SongID = dup.SongID
	}
	return f
}
