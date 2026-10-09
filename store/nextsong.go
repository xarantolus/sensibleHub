package store

import (
	"maps"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
	"time"
	"xarantolus/sensibleHub/store/analysis"
	"xarantolus/sensibleHub/store/music"
)

const (
	recentBoostCount = 50
	recentBoostMax   = 3.0
	sameArtistFactor = 0.5
)

// Suggestion is a song proposed to play next. Factors explain the score for debugging;
// a factor of 1 is neutral.
type Suggestion struct {
	ID      string
	Score   float64
	Factors map[string]float64
}

// NextSongs proposes up to count synced songs to play after current (which may be empty).
// Songs in exclude are skipped unless nothing else is left. The picks form a chain:
// each is drawn at random, weighted by how well it follows the previous pick
// (key, tempo, energy, timbre once analysed), by how recently it was added and
// by how rarely the listener skips it.
func (m *Manager) NextSongs(current string, count int, exclude []string) []Suggestion {
	m.SongsLock.RLock()
	prev, hasPrev := m.Songs[current]
	var pool []music.Entry
	for _, e := range m.Songs {
		if m.ShouldSync(e) && e.ID != current {
			pool = append(pool, e)
		}
	}
	m.SongsLock.RUnlock()

	excluded := make(map[string]bool, len(exclude))
	for _, id := range exclude {
		excluded[id] = true
	}
	candidates := slices.DeleteFunc(slices.Clone(pool), func(e music.Entry) bool { return excluded[e.ID] })
	if len(candidates) == 0 {
		candidates = pool
	}

	now := time.Now()
	recency := recencyWeights(pool)
	analyses := make([]*music.Analysis, 0, len(pool)+1)
	for _, e := range pool {
		analyses = append(analyses, e.Analysis)
	}
	if hasPrev {
		analyses = append(analyses, prev.Analysis)
	}
	stats := analysis.NewTimbreStats(analyses)

	var out []Suggestion
	for len(out) < count && len(candidates) > 0 {
		scored := make([]Suggestion, len(candidates))
		for i, e := range candidates {
			f := map[string]float64{"recency": recency[e.ID], "skips": skipFactor(e.Listening, now)}
			if hasPrev {
				if prev.MusicData.Artist != "" && strings.EqualFold(prev.MusicData.Artist, e.MusicData.Artist) {
					f["sameArtist"] = sameArtistFactor
				}
				maps.Copy(f, analysis.Transition(prev.Analysis, e.Analysis, stats))
			}
			score := 1.0
			for _, v := range f {
				score *= v
			}
			scored[i] = Suggestion{ID: e.ID, Score: score, Factors: f}
		}

		pick := weightedSample(scored, 1, rand.Float64)[0]
		out = append(out, pick)
		i := slices.IndexFunc(candidates, func(e music.Entry) bool { return e.ID == pick.ID })
		prev, hasPrev = candidates[i], true
		candidates = slices.Delete(candidates, i, i+1)
	}
	return out
}

// recencyWeights boosts the newest songs: the newest gets recentBoostMax, fading
// linearly to 1 at rank recentBoostCount. Older songs keep weight 1.
func recencyWeights(songs []music.Entry) map[string]float64 {
	sorted := slices.Clone(songs)
	slices.SortFunc(sorted, func(a, b music.Entry) int { return b.Added.Compare(a.Added) })

	w := make(map[string]float64, len(sorted))
	for rank, e := range sorted {
		w[e.ID] = 1
		if rank < recentBoostCount {
			w[e.ID] += (recentBoostMax - 1) * (1 - float64(rank)/recentBoostCount)
		}
	}
	return w
}

// weightedSample draws up to n items without replacement, each with probability
// proportional to its score. rnd returns values in [0, 1).
func weightedSample(items []Suggestion, n int, rnd func() float64) []Suggestion {
	items = slices.Clone(items)
	var out []Suggestion
	for len(out) < n && len(items) > 0 {
		var total float64
		for _, it := range items {
			total += math.Max(it.Score, 0)
		}
		i := len(items) - 1
		if total > 0 {
			target := rnd() * total
			for j, it := range items {
				target -= math.Max(it.Score, 0)
				if target < 0 {
					i = j
					break
				}
			}
		} else {
			i = int(rnd() * float64(len(items)))
		}
		out = append(out, items[i])
		items = slices.Delete(items, i, i+1)
	}
	return out
}
