package music

import (
	"encoding/json"
	"fmt"
	"image"

	// Supported image formats
	"image/jpeg"
	"image/png"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/image/webp"
)

var (
	_ = jpeg.Decode
	_ = png.Decode
	_ = webp.Decode
)

const (
	// Term must be replaced with url.QueryEscape,
	appleMusicURL = "https://itunes.apple.com/search?term=%s&entity=song&limit=1"

	sizeLimit = 10000 // Always try to get the highest quality picture
)

var (
	c = http.Client{
		Timeout: 60 * time.Second,
	}

	possibleUserAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.246",
		"Mozilla/5.0 (X11; CrOS x86_64 8172.45.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.64 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_2) AppleWebKit/601.3.9 (KHTML, like Gecko) Version/9.0.2 Safari/601.3.9",
	}
)

type SongData struct {
	Artist string
	Title  string
	Album  string
	Year   int // don't use if it's 0

	// might be nil
	Artwork          image.Image
	ArtworkExtension string
}

// SearchITunes searches the song with the given cleaned title and cleaned artist on iTunes.
func SearchITunes(title, album, artist string, currentExt string) (s SongData, err error) {
	title, album, artist = relevantInfo(title), relevantInfo(album), relevantInfo(artist)

	// Sometimes it doesn't find any songs when we add the album to the query, so we just try again without it
	var secondTry bool
retry:
	var searchTerms []string
	if artist != "" {
		searchTerms = append(searchTerms, artist)
	}
	if album != "" && !secondTry && !strings.EqualFold(title, album) {
		if len(strings.Fields(album)) < 3 {
			searchTerms = append(searchTerms, album)
		}
	}
	if title != "" {
		searchTerms = append(searchTerms, title)
	}

	resp, err := c.Get(fmt.Sprintf(appleMusicURL, url.QueryEscape(strings.Join(searchTerms, " "))))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected status code %d, wanted %d", resp.StatusCode, http.StatusOK)
		return
	}

	search := new(searchResults)
	err = json.NewDecoder(resp.Body).Decode(search)
	if err != nil {
		return
	}

	if len(search.Results) == 0 {
		if !secondTry {
			secondTry = true
			goto retry
		}
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
		var fetchImageFunc = func() error {
			req, err := http.NewRequest(http.MethodGet, u, nil)
			if err != nil {
				return err
			}

			req.Header.Set("User-Agent", possibleUserAgents[rand.Intn(len(possibleUserAgents))])
			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

			resp, err := c.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			img, _, err := image.Decode(resp.Body)
			if err != nil {
				return fmt.Errorf("decode %s: %w", u, err)
			}

			s.Artwork = img
			s.ArtworkExtension = aext

			return nil
		}

		// I'm aware, that looks weird.
		// This step of downloading & decoding the image doesn't always work,
		// and retrying with the exact same parameters etc. seems to work.
		// As in the first step returns an error that the image cannot be decoded,
		// but when trying again it works?
		// It's confusing, but at least it works
		for i := 5; i >= 0; i-- {
			err = fetchImageFunc()
			if err == nil {
				break
			}
			log.Println(err)

			time.Sleep(5 * time.Second)
		}
	}

	s.Artist = sres.ArtistName
	s.Album = strings.TrimSuffix(sres.CollectionName, " - Single")
	s.Title = sres.TrackName

	if !sres.ReleaseDate.IsZero() {
		s.Year = sres.ReleaseDate.Year()
	}

	return
}

// relevantInfo returns the relevant info from a string
// e.g: Title (feat ...) => Title
//
//	Artist1 & Artist2 => Artist1
func relevantInfo(s string) string {
	if i := strings.IndexAny(s, "&("); i != -1 {
		s = s[:i]
	}

	return strings.TrimSpace(s)
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
