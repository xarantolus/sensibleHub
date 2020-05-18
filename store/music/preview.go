package music

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"os"
	"sync"
	"time"

	"github.com/nfnt/resize"
	"golang.org/x/sync/singleflight"
)

var (
	// coverGroup manages the functions that generate cover previews.
	// They are started in HandleCover, and forgotten in HandleEditSong
	coverGroup singleflight.Group

	cstoreLock sync.RWMutex
	coverStore = make(map[string]cover)
)

type cover struct {
	date  time.Time
	bytes []byte
}

// CoverPreview generates a cover preview
func (e *Entry) CoverPreview() (c []byte, imageFormat string, err error) {
	imageFormat = "image/png"

	cstoreLock.RLock()
	cov, ok := coverStore[e.ID]
	if ok {
		if cov.date.Equal(e.LastEdit) {
			// we have to initialize it
			c = make([]byte, len(cov.bytes))

			copy(c, cov.bytes)
			cstoreLock.RUnlock()
			return
		}
		// we need to re-generate it
		coverGroup.Forget(e.ID)
	}
	cstoreLock.RUnlock()

	coverBytes, err, _ := coverGroup.Do(e.ID, func() (res interface{}, err error) {
		var b bytes.Buffer

		// always returns a png image
		err = resizeCover(e.CoverPath(), 60, &b)
		if err != nil {
			return
		}
		re := b.Bytes()

		cstoreLock.Lock()
		coverStore[e.ID] = cover{
			date:  e.LastEdit,
			bytes: re,
		}
		cstoreLock.Unlock()

		return re, nil
	})
	if err != nil {
		return
	}

	c = coverBytes.([]byte)

	return
}

func resizeCover(coverPath string, width uint, out io.Writer) (err error) {
	f, err := os.Open(coverPath)
	if err != nil {
		return
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return
	}

	resized := resize.Resize(width, 0, img, resize.Lanczos3)

	return png.Encode(out, resized)
}
