package store

import (
	"net/url"
	"sync"
	"testing"
	"xarantolus/sensibleHub/store/music"
)

func TestManager_hasLink(t *testing.T) {
	type args struct {
		u *url.URL
	}

	var m = Manager{
		SongsLock:    new(sync.RWMutex),
		enqueuedURLs: make(chan string, 25),

		Songs: map[string]music.Entry{
			"id": {
				SourceURL: "https://youtube.com/watch?v=videoid",
			},
		},
	}

	tests := map[string]bool{
		// Same url as above
		"https://www.youtube.com/watch?v=videoid": true,
		"https://youtube.com/watch?v=videoid":     true,
		"https://youtu.be/videoid":                true,
		"https://www.youtu.be/videoid":            true,

		"https://youtube.com/watch?v=videoidno":  false,
		"https://example.com":                    false,
		"https:/soundcloud.com/some-artist/song": false,
	}

	for turl, tok := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			parsed, err := url.ParseRequestURI(turl)
			if err != nil {
				panic(err)
			}

			_, gotOk := m.hasLink(parsed)
			if gotOk != tok {
				t.Errorf("Manager.hasLink() gotOk = %v, want %v for link %s", gotOk, tok, turl)
			}
		})
	}
}
