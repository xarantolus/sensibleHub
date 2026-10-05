package store

import (
	"math"
	"time"
	"xarantolus/sensibleHub/store/music"
)

const (
	listeningHalfLife = 30 * 24 * time.Hour
	skipWeight        = 2.0
	minSkipFactor     = 0.15
	recentSkipWindow  = 24 * time.Hour
	recentSkipFactor  = 0.5
	maxClockSkew      = time.Minute
)

// PlayEvent reports how a song was listened to. Skipped means the listener moved
// on early by choice; a song that played to the end or long enough is a play.
type PlayEvent struct {
	SongID   string
	At       time.Time
	Listened float64
	Skipped  bool
}

// decayed returns the counters as of now: every count halves per listeningHalfLife,
// so old skips stop mattering.
func decayed(l music.Listening, now time.Time) music.Listening {
	if dt := now.Sub(l.Updated); dt > 0 && !l.Updated.IsZero() {
		k := math.Pow(0.5, float64(dt)/float64(listeningHalfLife))
		l.Plays *= k
		l.Skips *= k
	}
	l.Updated = now
	return l
}

// ListeningNow returns l's counters decayed to the current time.
func ListeningNow(l music.Listening) music.Listening {
	return decayed(l, time.Now())
}

// RecordPlays adds listening events to the songs' statistics. Events for unknown
// songs are ignored, since a song may have been deleted while events were queued
// offline. It returns how many events were stored.
func (m *Manager) RecordPlays(events []PlayEvent) (int, error) {
	now := time.Now()

	m.SongsLock.Lock()
	defer m.SongsLock.Unlock()

	stored := 0
	for _, ev := range events {
		e, ok := m.Songs[ev.SongID]
		if !ok {
			continue
		}
		at := ev.At
		if at.IsZero() || at.After(now.Add(maxClockSkew)) {
			at = now
		}

		var l music.Listening
		if e.Listening != nil {
			l = *e.Listening
		}
		if at.After(l.Updated) {
			l = decayed(l, at)
		}
		if ev.Skipped {
			l.Skips++
			if at.After(l.LastSkipped) {
				l.LastSkipped = at
			}
		} else {
			l.Plays++
			if at.After(l.LastPlayed) {
				l.LastPlayed = at
			}
		}
		e.Listening = &l
		m.Songs[ev.SongID] = e
		stored++
	}
	if stored == 0 {
		return 0, nil
	}
	return stored, m.Save(false)
}

// skipFactor lowers the weight of songs the listener tends to skip: a Bayesian
// play/skip ratio that starts neutral, plus a cooldown right after a skip.
func skipFactor(l *music.Listening, now time.Time) float64 {
	if l == nil {
		return 1
	}
	d := decayed(*l, now)
	f := math.Max((1+d.Plays)/(1+d.Plays+skipWeight*d.Skips), minSkipFactor)
	if !l.LastSkipped.IsZero() && now.Sub(l.LastSkipped) < recentSkipWindow {
		f *= recentSkipFactor
	}
	return f
}
