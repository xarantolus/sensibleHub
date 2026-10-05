package store

import (
	"context"
	"errors"
	"log"
	"maps"
	"slices"
	"sync"
	"time"
	"xarantolus/sensibleHub/store/analysis"
	"xarantolus/sensibleHub/store/music"
)

const (
	analysisSaveEvery   = 10
	analysisIdlePoll    = 2 * time.Second
	analysisSongTimeout = 5 * time.Minute
)

type analysisQueue struct {
	mu      sync.Mutex
	pending map[string]bool
	wake    chan struct{}
	done    int
	total   int
	running bool
}

// AnalysisProgress counts analysed songs since the queue last became empty.
type AnalysisProgress struct {
	Done, Total int
	Running     bool
}

func (AnalysisProgress) isEvent() {}

func trimBounds(e music.Entry) (start, end float64) {
	start = max(e.AudioSettings.Start, 0)
	end = e.AudioSettings.End
	if end <= 0 || end >= e.MusicData.Duration {
		end = 0
	}
	return start, end
}

// needsAnalysis reports whether e lacks a current analysis. A failed analysis
// counts as current, so broken files are not retried on every start; it is
// retried when the algorithm version or the trim changes.
func needsAnalysis(e music.Entry) bool {
	a := e.Analysis
	if a == nil || a.Version < analysis.Version {
		return true
	}
	start, end := trimBounds(e)
	return a.Start != start || a.End != end
}

// AnalysisStatus returns how far the background analysis is.
func (m *Manager) AnalysisStatus() AnalysisProgress {
	m.analysis.mu.Lock()
	defer m.analysis.mu.Unlock()
	return AnalysisProgress{Done: m.analysis.done, Total: m.analysis.total, Running: m.analysis.running}
}

func (m *Manager) queueAnalysis(ids ...string) {
	if m.cfg.Analysis.Disabled || len(ids) == 0 {
		return
	}
	q := &m.analysis
	q.mu.Lock()
	if q.pending == nil {
		q.pending = map[string]bool{}
	}
	for _, id := range ids {
		if !q.pending[id] {
			q.pending[id] = true
			q.total++
		}
	}
	wake := q.wakeChan()
	q.mu.Unlock()

	select {
	case wake <- struct{}{}:
	default:
	}
}

func (q *analysisQueue) wakeChan() chan struct{} {
	if q.wake == nil {
		q.wake = make(chan struct{}, 1)
	}
	return q.wake
}

// StartAnalysis analyses every song that lacks a current analysis, newest first,
// and then keeps analysing songs as they are added or re-trimmed. It returns
// immediately; the work happens in the background, one song at a time.
func (m *Manager) StartAnalysis(ctx context.Context) {
	if m.cfg.Analysis.Disabled {
		log.Println("[Analysis] Disabled in config")
		return
	}

	m.analysis.mu.Lock()
	wake := m.analysis.wakeChan()
	m.analysis.mu.Unlock()

	var missing []string
	for _, e := range m.AllEntries() {
		if needsAnalysis(e) {
			missing = append(missing, e.ID)
		}
	}
	if len(missing) > 0 {
		log.Printf("[Analysis] %d songs need to be analysed\n", len(missing))
	}
	m.queueAnalysis(missing...)

	go func() {
		unsaved := 0
		for {
			id, ok := m.nextToAnalyse()
			if !ok {
				if unsaved > 0 {
					m.saveAnalyses()
					unsaved = 0
				}
				m.setAnalysisRunning(false)
				select {
				case <-ctx.Done():
					return
				case <-wake:
					continue
				}
			}
			m.setAnalysisRunning(true)

			for m.IsWorking() {
				select {
				case <-ctx.Done():
					return
				case <-time.After(analysisIdlePoll):
				}
			}

			if m.analyseOne(ctx, id) {
				unsaved++
			}
			if unsaved >= analysisSaveEvery {
				m.saveAnalyses()
				unsaved = 0
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()
}

// nextToAnalyse pops the most recently added pending song. It never holds the
// queue lock and SongsLock together: store mutations queue songs while holding SongsLock.
func (m *Manager) nextToAnalyse() (string, bool) {
	q := &m.analysis
	q.mu.Lock()
	pending := slices.Collect(maps.Keys(q.pending))
	q.mu.Unlock()

	var best string
	var bestAdded time.Time
	var gone []string
	m.SongsLock.RLock()
	for _, id := range pending {
		e, ok := m.Songs[id]
		if !ok {
			gone = append(gone, id)
			continue
		}
		if best == "" || e.Added.After(bestAdded) {
			best, bestAdded = id, e.Added
		}
	}
	m.SongsLock.RUnlock()

	q.mu.Lock()
	defer q.mu.Unlock()
	for _, id := range gone {
		if q.pending[id] {
			delete(q.pending, id)
			q.done++
		}
	}
	if best == "" {
		if len(q.pending) == 0 {
			q.done, q.total = 0, 0
		}
		return "", false
	}
	delete(q.pending, best)
	return best, true
}

func (m *Manager) setAnalysisRunning(running bool) {
	q := &m.analysis
	q.mu.Lock()
	changed := q.running != running
	q.running = running
	p := AnalysisProgress{Done: q.done, Total: q.total, Running: running}
	q.mu.Unlock()
	if changed {
		m.publish(p)
	}
}

// analyseOne analyses a song and stores the result; it reports whether the
// stored data changed.
func (m *Manager) analyseOne(ctx context.Context, id string) bool {
	e, ok := m.GetEntry(id)
	if !ok || !needsAnalysis(e) {
		m.finishAnalysis()
		return false
	}

	start, end := trimBounds(e)
	songCtx, cancel := context.WithTimeout(ctx, analysisSongTimeout)
	began := time.Now()
	result, err := analysis.Analyze(songCtx, m.cfg.Alternatives.FFmpeg, e.AudioPath(), start, end)
	cancel()
	if ctx.Err() != nil {
		return false
	}
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			err = errors.New("analysis took too long")
		}
		log.Printf("[Analysis] %s (%s): %v\n", e.SongName(), id, err)
		result.Error = err.Error()
	} else {
		log.Printf("[Analysis] %s: %s, %.1f BPM, %.1f LUFS (%s)\n", e.SongName(), result.Camelot, result.BPM, result.LoudnessLUFS, time.Since(began).Round(time.Millisecond))
	}

	m.SongsLock.Lock()
	current, ok := m.Songs[id]
	stillValid := ok
	if ok {
		cs, ce := trimBounds(current)
		stillValid = cs == start && ce == end
	}
	if stillValid {
		current.Analysis = &result
		m.Songs[id] = current
	}
	m.SongsLock.Unlock()

	m.finishAnalysis()
	if !ok {
		return false
	}
	if !stillValid {
		m.queueAnalysis(id)
		return false
	}
	m.publish(SongUpdated{Song: current})
	return true
}

func (m *Manager) finishAnalysis() {
	q := &m.analysis
	q.mu.Lock()
	q.done++
	p := AnalysisProgress{Done: q.done, Total: q.total, Running: true}
	q.mu.Unlock()
	m.publish(p)
}

func (m *Manager) saveAnalyses() {
	if err := m.Save(); err != nil {
		log.Printf("[Analysis] Saving results: %v\n", err)
	}
}
