package store

import (
	"sort"
	"strings"
	"xarantolus/sensibleHub/store/music"
)

// Album represents a music album. Songs is gurranteed to have one or more items
type Album struct {
	Title  string
	Artist string

	Songs []music.Entry
}

// GetAlbum gets the specified album for the given artist.
// Songs are sorted alphabetically as we don't store the title number
func (m *Manager) GetAlbum(artist, albumName string) (a Album, ok bool) {
	for _, e := range m.AllEntries() {
		// Wrong artist?
		if !strings.EqualFold(CleanName(e.Artist()), artist) {
			continue
		}

		// Wrong album?
		if !strings.EqualFold(CleanName(e.AlbumName()), albumName) {
			continue
		}

		a.Songs = append(a.Songs, e)
	}

	if len(a.Songs) == 0 {
		return a, false
	}

	a.Title = a.Songs[0].AlbumName()
	a.Artist = a.Songs[0].Artist()

	return a, true
}

// ArtistInfo contains a summary of an artist and all albums
type ArtistInfo struct {
	Name     string
	PlayTime float64

	YearStart int
	YearEnd   int

	Albums []Album

	Featured []music.Entry
}

// Artist returns the albums for the specified artist
func (m *Manager) Artist(artist string) (ai ArtistInfo, ok bool) {
	var res []Album

	var cleanedArtist = CleanName(artist)

	artist = strings.ToUpper(artist)

	var am = make(map[string]Album)
	for _, e := range m.AllEntries() {
		// Wrong artist?
		if !strings.EqualFold(CleanName(e.Artist()), cleanedArtist) {
			if !strings.Contains(strings.ToUpper(e.MusicData.Title), artist) {
				continue
			}

			ai.Featured = append(ai.Featured, e)
			continue
		}

		aname := CleanName(e.AlbumName())

		var combined = strings.ToUpper(aname)

		album := am[combined]

		// Album might be zero value, but that doesn't matter
		album.Songs = append(album.Songs, e)

		am[combined] = album

		ai.PlayTime += e.MusicData.Duration

		if e.MusicData.Year != nil {
			y := *e.MusicData.Year

			if y < ai.YearStart || ai.YearStart == 0 {
				ai.YearStart = y
			} else if y > ai.YearEnd || ai.YearEnd == 0 {
				ai.YearEnd = y
			}
		}
	}

	if len(am) == 0 {
		return ai, false
	}

	for _, a := range am {
		a.Title = a.Songs[0].AlbumName()
		a.Artist = a.Songs[0].Artist()
		res = append(res, a)
	}

	ai.Albums = res

	sort.Slice(res, func(i, j int) bool {
		return strings.ToUpper(res[i].Title) < strings.ToUpper(res[j].Title)
	})

	sort.Slice(ai.Featured, func(i, j int) bool {
		return strings.ToUpper(ai.Featured[i].MusicData.Title) < strings.ToUpper(ai.Featured[j].MusicData.Title)
	})

	ai.Name = ai.Albums[0].Artist

	return ai, true
}

// GroupByAlbum groups songs by their artist and albums
func (m *Manager) GroupByAlbum() (res []Album) {
	var normalizedArtists = make(map[string]string)

	// this code is very similar to `ftp/musicfactory.go`
	// in fact, I even copied most of it
	var am = make(map[string]Album)

	for _, e := range m.AllEntries() {
		art := CleanName(e.Artist())
		artistName, ok := normalizedArtists[strings.ToUpper(art)]
		if !ok {
			normalizedArtists[strings.ToUpper(e.Artist())] = e.Artist()
			artistName = art
		}

		aname := CleanName(e.AlbumName())

		var combined = strings.ToUpper(art + "/" + aname)

		album, ok := am[combined]
		if !ok {
			album = Album{
				Title:  e.AlbumName(),
				Artist: artistName,
			}
		}

		album.Songs = append(album.Songs, e)

		am[combined] = album
	}

	for _, a := range am {
		res = append(res, a)
	}

	sort.Slice(res, func(i, j int) bool {
		if strings.EqualFold(res[i].Artist, res[j].Artist) {
			return strings.ToUpper(res[i].Title) < strings.ToUpper(res[j].Title)
		}

		return res[i].Artist < res[j].Artist
	})

	return
}

// CleanName converts `n` to an url-safe string
func CleanName(n string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
			return r
		}

		if r >= '0' && r <= '9' {
			return r
		}

		if r == '-' || r == '.' || r == ' ' || r == '(' || r == ')' {
			return r
		}

		return -1
	}, n)
}
