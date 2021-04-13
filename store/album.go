package store

import (
	"sort"
	"strings"

	"xarantolus/sensibleHub/store/music"
)

// Album represents a music album. Songs is guaranteed to have one or more items
type Album struct {
	Title  string
	Artist string

	Songs []music.Entry
}

func (a *Album) AlbumTitle() string {
	if a.Title == "" {
		return "Unknown"
	}
	return a.Title
}

// setupAlbum moves the song with the same title as the album name to the first place
func (a *Album) setupAlbum() (ret *Album) {
	a.Title = a.Songs[0].MusicData.Album
	a.Artist = a.Songs[0].MusicData.Artist

	if len(a.Songs) < 2 || a.Title == "" {
		return a
	}

	var firstSong int = -1

	for i, s := range a.Songs {
		title := s.MusicData.Title
		// Title (feat. something)
		if bi := strings.IndexRune(title, '('); bi != -1 {
			title = strings.TrimSpace(title[:bi])
		}

		// If song title without feature artist == album title
		if strings.EqualFold(CleanName(title), CleanName(a.Title)) {
			firstSong = i
			break
		}
	}

	// If it is -1, we couldn't find any song that should be at the front so we ignore it
	// If it is 0, we already have the correct order
	if firstSong < 1 {
		return a
	}

	newSongs := make([]music.Entry, len(a.Songs))
	newSongs[0] = a.Songs[firstSong]

	// Copy before first song
	n := copy(newSongs[1:], a.Songs[:firstSong])

	copy(newSongs[n+1:], a.Songs[firstSong+1:])

	return &Album{
		newSongs[0].MusicData.Album,
		newSongs[0].MusicData.Artist,
		newSongs,
	}
}

// GetAlbum gets the specified album for the given artist.
// Songs are sorted alphabetically as we don't store the title number
func (m *Manager) GetAlbum(artist, albumName string) (a Album, ok bool) {
	artist, albumName = CleanName(artist), CleanName(albumName)

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

	return *a.setupAlbum(), true
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

	cleanedArtist := CleanName(artist)

	artist = strings.ToUpper(artist)

	am := make(map[string]Album)

	unknownAlbum := Album{}

	for _, e := range m.AllEntries() {
		// Wrong artist?
		if !strings.EqualFold(CleanName(e.Artist()), cleanedArtist) {
			ti := strings.ToUpper(e.MusicData.Title)

			// We assume that the song title is something like
			//  Title (feat. FirstArtist & SecondArtist)
			firstBracket, lastBracket := strings.IndexByte(ti, '('), strings.LastIndexByte(ti, ')')

			// cannot find the "feat." part in brackets:
			if firstBracket == -1 || lastBracket == -1 {
				continue
			}

			// Both are uppercase, check if we can find the artist in there
			if strings.Contains(ti[firstBracket:lastBracket], artist) {
				ai.Featured = append(ai.Featured, e)
			}

			continue
		}

		aname := CleanName(e.MusicData.Album)

		combined := strings.ToUpper(aname)

		var album Album
		if aname == "" {
			album = unknownAlbum
		} else {
			album = am[combined]
		}

		// Album might be zero value, but that doesn't matter
		album.Songs = append(album.Songs, e)

		if aname == "" {
			unknownAlbum = album
		} else {
			am[combined] = album
		}

		ai.PlayTime += e.MusicData.Duration

		if e.MusicData.Year != nil {
			y := *e.MusicData.Year

			if y < ai.YearStart || ai.YearStart == 0 {
				ai.YearStart = y
			}
			if y > ai.YearEnd || ai.YearEnd == 0 {
				ai.YearEnd = y
			}
		}
	}

	if len(am) == 0 && len(unknownAlbum.Songs) == 0 {
		return ai, false
	}

	for _, a := range am {
		a = *a.setupAlbum()

		res = append(res, a)
	}

	ai.Albums = res

	sort.Slice(res, func(i, j int) bool {
		return strings.ToUpper(res[i].Title) < strings.ToUpper(res[j].Title)
	})

	sort.Slice(ai.Featured, func(i, j int) bool {
		return strings.ToUpper(ai.Featured[i].MusicData.Title) < strings.ToUpper(ai.Featured[j].MusicData.Title)
	})

	// Unknown albums should always be the last album
	if len(unknownAlbum.Songs) > 0 {
		unknownAlbum.Artist = unknownAlbum.Songs[0].MusicData.Artist

		ai.Albums = append(ai.Albums, unknownAlbum)
	}

	ai.Name = ai.Albums[0].Artist

	return ai, true
}

// GroupByAlbum groups songs by their artist and albums
func (m *Manager) GroupByAlbum() (res []Album) {
	normalizedArtists := make(map[string]string)

	// this code is very similar to `ftp/musicfactory.go`
	// in fact, I even copied most of it
	am := make(map[string]Album)

	for _, e := range m.AllEntries() {
		art := CleanName(e.Artist())
		artistName, ok := normalizedArtists[strings.ToUpper(art)]
		if !ok {
			normalizedArtists[strings.ToUpper(e.Artist())] = e.Artist()
			artistName = art
		}

		aname := CleanName(e.AlbumName())

		combined := strings.ToUpper(art + "/" + aname)

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
		a = *a.setupAlbum()
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

func equalBrackets(a, b string) bool {
	return strings.EqualFold(cleanBrackets(a), cleanBrackets(b))
}

func cleanBrackets(a string) string {
	for {
		var sb, eb = strings.IndexRune(a, '('), strings.IndexRune(a, ')')
		if sb == -1 || eb == -1 || sb > eb {
			break
		}

		a = a[:sb] + a[eb+1:]
	}

	return strings.Join(strings.Fields(a), " ")
}
