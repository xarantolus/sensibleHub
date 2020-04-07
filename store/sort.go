package store

import (
	"sort"
	"strings"
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

func isLetter(r rune) bool {
	if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' {
		return true
	}
	return false
}
