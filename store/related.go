package store

import (
	"sort"
	"strings"
	"xarantolus/sensibleHub/store/music"
)

// This doesn't really work

// GetRelatedSongs returns some related songs for a song
func (m *Manager) GetRelatedSongs(e music.Entry) (out []music.Entry) {
	if e.MusicData.Artist == "" {
		return
	}

	e.MusicData.Artist = strings.ToUpper(CleanName(e.MusicData.Artist))

	m.SongsLock.RLock()
	defer m.SongsLock.RUnlock()

	type suggestion struct {
		score int
		s     music.Entry
	}
	var suggestions []suggestion

	for _, s := range m.Songs {
		if s.ID == e.ID {
			continue
		}
		var sc, mult int = 0, 2

		// Is there a song with the same title? definitely show it
		if strings.EqualFold(CleanName(s.MusicData.Title), CleanName(e.MusicData.Title)) {
			sc += 10000
		}

		if !strings.EqualFold(strings.ToUpper(CleanName(s.MusicData.Artist)), e.MusicData.Artist) {
			ti := strings.ToUpper(CleanName(s.MusicData.Title))

			firstBracket, lastBracket := strings.IndexByte(ti, '('), strings.LastIndexByte(ti, ')')

			// find the "feat." part in brackets:
			if firstBracket != -1 && lastBracket != -1 {
				// Both are uppercase, check if we can find the artist in there
				if strings.Contains(ti[firstBracket:lastBracket], e.MusicData.Artist) {
					sc += 25
					mult = 3
				}
			}

			cleanedET := CleanName(e.MusicData.Title)
			if b := strings.IndexByte(cleanedET, '('); b != -1 {
				cleanedET = strings.TrimSpace(cleanedET[:b])
			}
			cleanedST := CleanName(s.MusicData.Title)
			if b := strings.IndexByte(cleanedST, '('); b != -1 {
				cleanedST = strings.TrimSpace(cleanedST[:b])
			}

			// It's also the same title, but one of them has an (feat. abcd) somewhere in there
			if strings.EqualFold(cleanedST, cleanedET) {
				sc += 10000
			}
		}

		sc += score(strings.Fields(strings.ToUpper(s.MusicData.Title)), e.MusicData.Title, 2)
		sc += score(strings.Fields(strings.ToUpper(s.MusicData.Album)), e.MusicData.Album, 1)
		sc += score(strings.Fields(strings.ToUpper(s.MusicData.Artist)), e.MusicData.Artist, 2)

		// If add songs with the same artist as songs with a score of 0. Artists not featured anywhere will also get some similar songs then
		if sc > 0 {
			suggestions = append(suggestions, suggestion{sc * mult, s})
		}
	}

	if len(suggestions) == 0 {
		return
	}

	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].score > suggestions[j].score
	})

	// Return one
	if len(suggestions) < 2 {
		return []music.Entry{suggestions[0].s}
	}

	if len(suggestions) < 3 {
		return []music.Entry{suggestions[0].s, suggestions[1].s}
	}

	// Or two items
	return []music.Entry{suggestions[0].s, suggestions[1].s, suggestions[2].s}
}
