package store

import (
	"sort"
	"strconv"
	"strings"
	"time"
	"xarantolus/sensiblehub/store/music"
)

// AllEntries returns an alphabetically sorted list of entries.
// Special characters and numbers will be at the beginning
func (m *Manager) AllEntries() (list []music.Entry) {
	// This is usally the second lock for one operation, but I don't care
	m.SongsLock.RLock()
	defer m.SongsLock.RUnlock()

	for _, e := range m.Songs {
		list = append(list, e)
	}

	sort.Slice(list, func(i, j int) bool {
		return strings.ToUpper(list[i].MusicData.Title) < strings.ToUpper(list[j].MusicData.Title)
	})

	return
}

// Group is a struct that stores songs that are grouped together,
// e.g. because they the same beginning or artist
type Group struct {
	Title string

	Songs []music.Entry
}

// GroupByTitle groups titles by the first letter of their title
func (m *Manager) GroupByTitle() (groups []Group) {
	m.SongsLock.RLock()
	defer m.SongsLock.RUnlock()

	list := m.AllEntries() // this must be sorted alphabetically, with special characters being first

	var g *Group

	var lastTitleStart rune = -1

	var appendedGroup bool
	for _, song := range list {
		title := strings.ToUpper(song.MusicData.Title)
		// This shouldn't happen, but treat it as a special character
		if len(title) == 0 {
			title = "!"
		}

		// Same start? Same group
		if rune(title[0]) == lastTitleStart || !isLetter(rune(title[0])) {
			if g == nil {
				groupTitle := rune(title[0])
				if !isLetter(groupTitle) {
					groupTitle = '#' // Special characters and numbers
				}

				g = &Group{
					Title: string(groupTitle),
					Songs: []music.Entry{song},
				}
			}
			g.Songs = append(g.Songs, song)
			continue
		}

		// Store the old group. On the first iteration, this will be nil
		if g != nil {
			groups = append(groups, *g)
			appendedGroup = true
		}

		// Create a new group

		groupTitle := rune(title[0])
		if !isLetter(groupTitle) {
			groupTitle = '#' // Special characters and numbers
		}
		g = &Group{
			Title: string(groupTitle),
			Songs: []music.Entry{song},
		}
		appendedGroup = false
		lastTitleStart = groupTitle
	}

	// Now store the last group that was created
	if g != nil && !appendedGroup {
		groups = append(groups, *g)
	}

	return
}

// GroupByArtist groups songs by their artist
func (m *Manager) GroupByArtist() (groups []Group) {
	m.SongsLock.RLock()
	defer m.SongsLock.RUnlock()

	var artMap = map[string][]music.Entry{}

	for _, song := range m.AllEntries() {
		artist := strings.ToUpper(song.MusicData.Artist)
		if strings.TrimSpace(artist) == "" {
			artist = "???"
		}

		artMap[artist] = append(artMap[artist], song)
	}

	for _, songs := range artMap {
		// since every len(songs) > 0
		groups = append(groups, Group{
			Title: songs[0].MusicData.Artist, // don't use the upper-case artist
			Songs: songs,
		})
	}

	// Sort artists alphabetically
	sort.Slice(groups, func(i, j int) bool {
		return strings.ToUpper(groups[i].Title) < strings.ToUpper(groups[j].Title)
	})

	return
}

// GroupByYear groups songs by their year
func (m *Manager) GroupByYear() (groups []Group) {
	m.SongsLock.RLock()
	defer m.SongsLock.RUnlock()

	var yMap = map[string][]music.Entry{}

	for _, song := range m.AllEntries() {
		var yearName string
		if song.MusicData.Year == nil {
			yearName = "#"
		} else {
			yearName = strconv.Itoa(*song.MusicData.Year)
		}

		yMap[yearName] = append(yMap[yearName], song)
	}

	for year, songs := range yMap {
		// Sort songs in year listing by title
		sort.Slice(songs, func(i, j int) bool {
			return strings.ToUpper(songs[i].MusicData.Title) < strings.ToUpper(songs[j].MusicData.Title)
		})

		// since every len(songs) > 0
		groups = append(groups, Group{
			Title: year,
			Songs: songs,
		})
	}

	// Sort years alphabetically, newest year at the top, unknown years at the bottom
	sort.Slice(groups, func(i, j int) bool {
		return strings.ToUpper(groups[i].Title) > strings.ToUpper(groups[j].Title)
	})

	return
}

func isLetter(r rune) bool {
	if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' {
		return true
	}
	return false
}

const (
	newSongs = 7
)

// Newest returns the newest entries
func (m *Manager) Newest() (list []music.Entry, today bool) {
	entries := m.AllEntries()
	if len(entries) == 0 {
		return
	}

	// Sort by date added, newest at the top
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Added.After(entries[j].Added)
	})

	// Try to get all songs that were added today
	date := time.Now()
	if date.Hour() > 12 {
		date = date.Add(-12 * time.Hour)
	}
	date = date.Round(24 * time.Hour)

	for _, song := range entries {
		if song.Added.Before(date) {
			break
		}

		list = append(list, song)
	}

	if len(list) >= (newSongs / 2) {
		return list, true
	}

	// If not enough were added today, we just take the first few - this can include songs that were *not* added today

	limit := newSongs
	if len(entries) < limit {
		limit = len(entries)
	}

	return entries[:limit], false
}

// NewestSong returns the most recently added song
func (m *Manager) NewestSong() (e music.Entry, ok bool) {
	m.SongsLock.RLock()
	defer m.SongsLock.RUnlock()

	for _, en := range m.Songs {
		if en.Added.After(e.Added) {
			e = en
			ok = true
		}
	}

	return
}
