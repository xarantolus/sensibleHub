package music

import (
	"encoding/json"
	"fmt"
	"image"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// Term must be replaced with url.QueryEscape,
	appleMusicURL = "https://itunes.apple.com/search?term=%s&entity=song&limit=1"

	sizeLimit = 10000 // Always try to get the highest quality picture
)

var c = http.Client{
	Timeout: 30 * time.Second,
}

type SongData struct {
	Artist string
	Title  string
	Album  string

	// might be nil
	Artwork          image.Image
	ArtworkExtension string
}

// SearchITunes searches the song with the given cleaned title and cleaned artist on iTunes.
func SearchITunes(cleanTitle, cleanArtist string, currentExt string) (s SongData, err error) {
	searchTerm := cleanArtist + " " + cleanTitle
	if cleanArtist == "" {
		searchTerm = cleanTitle
	}
	if searchTerm == "" {
		err = fmt.Errorf("search term is empty")
		return
	}

	resp, err := c.Get(fmt.Sprintf(appleMusicURL, url.QueryEscape(searchTerm)))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected status code %d, wanted %d", resp.StatusCode, http.StatusOK)
		return
	}

	var search = new(searchResults)
	err = json.NewDecoder(resp.Body).Decode(search)
	if err != nil {
		return
	}

	if len(search.Results) == 0 {
		err = fmt.Errorf("cannot find this song on apple music")
		return
	}

	sres := search.Results[0]

	// use the same extension we already have *or* jpg
	aext := "jpg"
	tc := strings.TrimPrefix(strings.ToLower(currentExt), ".")
	if tc == "jpeg" {
		tc = "jpg"
	}
	if tc == "jpg" || tc == "png" {
		aext = tc
	}

	u, ok := sres.highResImageURL(aext)
	if ok {
		err = func() error {
			resp, err := c.Get(u)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			img, _, err := image.Decode(resp.Body)
			if err != nil {
				return err
			}

			s.Artwork = img
			s.ArtworkExtension = aext

			return nil
		}()
	}

	s.Artist = sres.ArtistName
	s.Album = sres.CollectionName
	s.Title = sres.TrackName

	return
}

type searchResults struct {
	ResultCount int             `json:"resultCount"`
	Results     []appleSongData `json:"results"`
}
type appleSongData struct {
	WrapperType            string    `json:"wrapperType"`
	Kind                   string    `json:"kind"`
	ArtistID               int       `json:"artistId"`
	CollectionID           int       `json:"collectionId"`
	TrackID                int       `json:"trackId"`
	ArtistName             string    `json:"artistName"`
	CollectionName         string    `json:"collectionName"`
	TrackName              string    `json:"trackName"`
	CollectionCensoredName string    `json:"collectionCensoredName"`
	TrackCensoredName      string    `json:"trackCensoredName"`
	ArtistViewURL          string    `json:"artistViewUrl"`
	CollectionViewURL      string    `json:"collectionViewUrl"`
	TrackViewURL           string    `json:"trackViewUrl"`
	PreviewURL             string    `json:"previewUrl"`
	ArtworkURL30           string    `json:"artworkUrl30"`
	ArtworkURL60           string    `json:"artworkUrl60"`
	ArtworkURL100          string    `json:"artworkUrl100"`
	CollectionPrice        float64   `json:"collectionPrice"`
	TrackPrice             float64   `json:"trackPrice"`
	ReleaseDate            time.Time `json:"releaseDate"`
	CollectionExplicitness string    `json:"collectionExplicitness"`
	TrackExplicitness      string    `json:"trackExplicitness"`
	DiscCount              int       `json:"discCount"`
	DiscNumber             int       `json:"discNumber"`
	TrackCount             int       `json:"trackCount"`
	TrackNumber            int       `json:"trackNumber"`
	TrackTimeMillis        int       `json:"trackTimeMillis"`
	Country                string    `json:"country"`
	Currency               string    `json:"currency"`
	PrimaryGenreName       string    `json:"primaryGenreName"`
	IsStreamable           bool      `json:"isStreamable"`
}

// highResImageURL returns a high-resolution url for this song item.
// format can be jpg or png
func (a *appleSongData) highResImageURL(format string) (url string, ok bool) {
	for _, u := range []string{a.ArtworkURL100, a.ArtworkURL60, a.ArtworkURL30} {
		li := strings.LastIndex(u, "/")
		if li < 0 {
			continue
		}

		// URLS end with something like `/source/100x100bb.jpg`
		// we can now just set a different filename after `/source/`
		// and it will be generated. (also works with different formats)
		sourceURL := u[:li]

		filename := fmt.Sprintf("%dx%d.%s", sizeLimit, sizeLimit, format)

		return sourceURL + "/" + filename, true
	}

	return "", false
}
