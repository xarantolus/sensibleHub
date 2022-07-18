package music

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"golang.org/x/sync/singleflight"
)

var (
	coverGroup singleflight.Group

	cstoreLock sync.RWMutex
	coverStore = make(map[string]cover)

	semaphore = make(chan struct{}, runtime.NumCPU()*2)
)

type cover struct {
	date  time.Time
	bytes []byte
}

// PreviewSize returns the size of all preview items in bytes
func PreviewSize() (size int64) {
	cstoreLock.RLock()
	defer cstoreLock.RUnlock()

	for _, cov := range coverStore {
		size += int64(len(cov.bytes))
	}

	return
}

// PreviewLen returns how many covers are currently in the cache
func PreviewLen() int {
	cstoreLock.RLock()
	defer cstoreLock.RUnlock()

	return len(coverStore)
}

// CoverPreview generates a cover preview
func (e *Entry) CoverPreview() (c []byte, imageFormat string, err error) {
	imageFormat = "image/jpeg"

	cstoreLock.RLock()
	cov, ok := coverStore[e.ID]
	if ok {
		if cov.date.Equal(e.LastEdit) {
			cstoreLock.RUnlock()
			return cov.bytes, imageFormat, nil
		}
		// we need to re-generate it
		coverGroup.Forget(e.ID)
	}
	cstoreLock.RUnlock()

	// Limit parallel processing
	semaphore <- struct{}{}
	defer func() {
		<-semaphore
	}()

	coverBytes, err, _ := coverGroup.Do(e.ID, func() (res interface{}, err error) {
		var b bytes.Buffer

		// always returns a jpeg image
		err = resizeCover(e.CoverPath(), 120, &b)
		if err != nil {
			return
		}

		var resBytes = make([]byte, b.Len())
		copy(resBytes, b.Bytes())

		cstoreLock.Lock()
		coverStore[e.ID] = cover{
			date:  e.LastEdit,
			bytes: resBytes,
		}
		cstoreLock.Unlock()

		b.Reset()

		return resBytes, nil
	})
	if err != nil {
		return
	}

	c = coverBytes.([]byte)

	runtime.GC()

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

	resized := imaging.Fit(img, int(width), int(width), imaging.Lanczos)

	return jpeg.Encode(out, resized, &jpeg.Options{
		Quality: jpeg.DefaultQuality,
	})
}
