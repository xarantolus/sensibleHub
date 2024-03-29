package store

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"xarantolus/sensibleHub/store/music"
)

// AllEntries returns an alphabetically sorted list of entries.
// Special characters and numbers will be at the beginning
func (m *Manager) AllEntries() (list []music.Entry) {
	// This is usually the second lock for one operation, but I don't care
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
// e.g. because they have the same prefix or artist
type Group struct {
	Title       string `json:"title"`
	Description string `json:"description"`

	Link string `json:"link"`

	Songs []music.Entry `json:"songs"`
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

	artMap := map[string][]music.Entry{}

	for _, song := range m.AllEntries() {
		artist := strings.ToUpper(CleanName(song.MusicData.Artist))
		if strings.TrimSpace(artist) == "" {
			artist = "???"
		}

		artMap[artist] = append(artMap[artist], song)
	}

	for _, songs := range artMap {
		// since every len(songs) > 0
		groups = append(groups, Group{
			Title:       songs[0].MusicData.Artist, // don't use the upper-case artist
			Description: songLenDescription(len(songs)),
			Songs:       songs,
			Link:        "/artist/" + CleanName(songs[0].Artist()),
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

	yMap := map[string][]music.Entry{}

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
			Title:       year,
			Description: songLenDescription(len(songs)),
			Songs:       songs,
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
	newSongs = 10
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

// musicExtensions is a list of music extensions that *should not* be at the end of a song title, see https://en.wikipedia.org/wiki/Audio_file_format#List_of_formats
// it also defines all audio formats that will be accepted after downloading
var musicExtensions = map[string]bool{
	"3gp":  true,
	"aa":   true,
	"aac":  true,
	"aax":  true,
	"act":  true,
	"aiff": true,
	"alac": true,
	"amr":  true,
	"ape":  true,
	"au":   true,
	"awb":  true,
	"dct":  true,
	"dss":  true,
	"dvf":  true,
	"flac": true,
	"gsm":  true,
	"ikla": true,
	"ivs":  true,
	"m4a":  true,
	"m4b":  true,
	"m4p":  true,
	"mmf":  true,
	"mp3":  true,
	"mpc":  true,
	"msv":  true,
	"nmf":  true,
	"nsf":  true,
	"ogg":  true,
	"oga":  true,
	"mogg": true,
	"opus": true,
	"ra":   true,
	"rm":   true,
	"raw":  true,
	"rf64": true,
	"sln":  true,
	"tta":  true,
	"voc":  true,
	"vox":  true,
	"wav":  true,
	"wma":  true,
	"wv":   true,
	"webm": true,
	"8svx": true,
	"cda":  true,
}

// Incomplete returns all entries with incomplete data
func (m *Manager) Incomplete() (groups []Group) {
	noArtist := Group{Title: "No Artist", Description: "We don't even know who made these songs"}
	noAlbum := Group{Title: "No Album", Description: "Missing album information"}
	noImage := Group{Title: "No Cover", Description: "There's no cover image for these songs"}
	noYear := Group{Title: "No Year", Description: "The year tag is missing"}
	weirdTitle := Group{Title: "Weird Title", Description: "Weird titles that should probably be changed"}
	smallCover := Group{Title: "Small Cover", Description: "Songs with a cover image less than 750 pixels in size"}

	// The conditions inside this loop must have the same order as the
	// lists in the for below it. That way, songs can cascade through these categories
	for _, e := range m.AllEntries() {
		// Maybe an import happened here and nothing has been done to the title yet
		if musicExtensions[strings.ToLower(strings.TrimPrefix(filepath.Ext(e.MusicData.Title), "."))] {
			weirdTitle.Songs = append(weirdTitle.Songs, e)
			continue
		}

		if strings.TrimSpace(e.MusicData.Artist) == "" {
			noArtist.Songs = append(noArtist.Songs, e)
			continue
		}

		if strings.TrimSpace(e.PictureData.Filename) == "" {
			noImage.Songs = append(noImage.Songs, e)
			continue
		}

		if strings.TrimSpace(e.MusicData.Album) == "" {
			noAlbum.Songs = append(noAlbum.Songs, e)
			continue
		}

		if e.PictureData.Size < 750 {
			smallCover.Songs = append(smallCover.Songs, e)
			continue
		}

		if e.MusicData.Year == nil || *e.MusicData.Year == 0 {
			noYear.Songs = append(noYear.Songs, e)
			continue
		}
	}

	for _, g := range []Group{weirdTitle, noArtist, noImage, noAlbum, noYear, smallCover} {
		if len(g.Songs) == 0 {
			continue
		}
		sort.Slice(g.Songs, func(i, j int) bool {
			return g.Songs[i].Added.After(g.Songs[j].Added)
		})
		groups = append(groups, g)
	}

	return
}

// Unsynced returns all songs that have syncing disabled
func (m *Manager) Unsynced() (groups []Group) {
	g := Group{
		Title: "Unsynced",
	}

	for _, song := range m.AllEntries() {
		if song.SyncSettings.Should {
			continue
		}

		g.Songs = append(g.Songs, song)
	}

	if len(g.Songs) > 0 {
		g.Description = fmt.Sprintf("%d songs are currently not synced to your devices", len(g.Songs))
		groups = []Group{g}
	}

	return
}

// RecentlyEdited returns a group of songs that were edited within
// the last two weeks, sorted by newest edit
func (m *Manager) RecentlyEdited() (groups []Group) {
	g := Group{
		Title:       "Recently edited",
		Description: "Songs edited within the last two weeks",
	}

	twoWeeksAgo := time.Now().Add(-14 * 24 * time.Hour)

	for _, song := range m.AllEntries() {
		if song.LastEdit.After(twoWeeksAgo) {
			g.Songs = append(g.Songs, song)
		}
	}

	if len(g.Songs) > 0 {
		sort.Slice(g.Songs, func(i, j int) bool {
			return g.Songs[i].LastEdit.After(g.Songs[j].LastEdit)
		})

		groups = []Group{g}
	}

	return
}

// SortedByAddDate returns all songs in a single group that is sorted by the date added (inverse)
func (m *Manager) SortedByAddDate() (groups []Group) {
	g := Group{
		Title:       "Date added",
		Description: "Songs ordered inversely by the date added",
	}

	entries := m.AllEntries()
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Added.After(entries[j].Added)
	})

	if len(entries) > 0 {
		g.Songs = entries
		groups = []Group{g}
	}

	return
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

func songLenDescription(count int) string {
	if count < 5 {
		return ""
	}
	return strconv.Itoa(count) + " songs"
}

// RandomSong returns a randomly chosen song. `ok` is false in case we have no songs at all
func (m *Manager) RandomSong() (e music.Entry, ok bool) {
	songs := m.AllEntries()
	if len(songs) == 0 {
		return
	}

	return songs[rand.Intn(len(songs))], true
}
