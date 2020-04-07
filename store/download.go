package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"xarantolus/sensiblehub/store/music"

	"image"
	"image/jpeg"
	"image/png"
	_ "image/png"
)

const (
	songDirTemplate = "data/songs/%s"
)

// Download downloads the song from the given URL using youtube-dl and saves it to the appropriate directory
func (m *Manager) Download(url string) (err error) {
	tmpDir, err := ioutil.TempDir("", "shub")
	if err != nil {
		return
	}
	// Delete temporary directory after downloading
	defer func() {
		derr := os.RemoveAll(tmpDir)
		if err == nil {
			err = derr
		}
	}()

	// Setup youtube-dl command and run it
	cmd := exec.Command("youtube-dl", "--write-info-json", "--write-thumbnail", "-f", "bestaudio/best", "--max-downloads", "1", "--no-playlist", "-x", "-o", "%(id)s.%(ext)s", url)
	cmd.Dir = tmpDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("while running youtube-dl: %s\nOutput: %s", err.Error(), string(out))
	}

	var (
		jsonPath string

		thumbPath string
		thumbMime string

		audioPath string
		audioSize int64
	)

	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		ext := strings.TrimPrefix(strings.ToUpper(filepath.Ext(path)), ".")

		switch ext {
		case "JSON":
			jsonPath = path
		case "JPG", "JPEG":
			thumbPath = path
			thumbMime = "image/jpeg"
		case "PNG":
			thumbPath = path
			thumbMime = "image/png"
		case "TEMP", "TMP":
			return fmt.Errorf("Invalid temporary file %s, assuming error", filepath.Base(path))
		default:
			// Assume this is the audio file
			audioPath = path
			audioSize = info.Size()
		}

		return nil
	})
	if err != nil {
		return
	}

	dur, err := getAudioDuration(audioPath)
	if err != nil {
		return fmt.Errorf("cannot get audio duration: %w", err)
	}

	minfo, jsonErr := readInfoFile(jsonPath)
	if jsonErr != nil {
		log.Println("Error while reading info file: ", jsonErr.Error())
		// Continue without data
	}

	m.SongsLock.Lock()
	defer m.SongsLock.Unlock()

	now := time.Now()
	var e = &music.Entry{
		ID:        m.generateID(),
		SourceURL: minfo.Webpage(url),

		LastEdit: now,
		Added:    now,

		// Assume that songs should be synced by default
		SyncSettings: music.SyncSettings{
			Should: true,
		},
		MetaFile: music.MetaFile{
			Filename: "info.json",
		},
		FileData: music.FileData{
			Filename: "original" + filepath.Ext(audioPath),
			Size:     audioSize,
		},
		AudioSettings: music.AudioSettings{
			Start: -1,
			End:   -1,
		},
		MusicData: music.MusicData{
			Title:    cascadeStrings(minfo.Track, minfo.Title, filepath.Base(minfo.Filename)),
			Artist:   cascadeStrings(minfo.Artist, minfo.Creator, minfo.Uploader),
			Album:    cascadeStrings(minfo.Album, minfo.Playlist, minfo.PlaylistTitle),
			Year:     minfo.Year(),
			Duration: dur,
		},
		PictureData: music.PictureData{
			MimeType: thumbMime,
			Filename: "cover" + filepath.Ext(thumbPath),
		},
	}

	// Create song dir
	songDir := fmt.Sprintf(songDirTemplate, e.ID)
	err = os.MkdirAll(songDir, 0644)
	if err != nil {
		return
	}
	// Delete songDir if something goes wrong
	defer func() {
		if err != nil {
			_ = os.RemoveAll(songDir)
		}
	}()

	// Move all kinds of files - this may not work on all platforms as they aren't in the same directory
	if jsonErr == nil {
		err = os.Rename(jsonPath, filepath.Join(songDir, e.MetaFile.Filename))
		if err != nil {
			return
		}
	}
	if thumbPath != "" {
		err = cropMoveCover(thumbPath, filepath.Join(songDir, e.PictureData.Filename))
		if err != nil {
			// reset image info
			e.PictureData.Filename = ""
			e.PictureData.MimeType = ""
		}
	}
	err = os.Rename(audioPath, filepath.Join(songDir, e.FileData.Filename))
	if err != nil {
		return
	}

	return m.Add(e)
}

type info struct {
	Track    string `json:"track"`     // Prefered
	Title    string `json:"title"`     // Fallback
	Filename string `json:"_filename"` // Fallback

	Artist   string `json:"artist"`   // Prefered
	Creator  string `json:"creator"`  // Maybe
	Uploader string `json:"uploader"` // If nothing else has info

	Album         string `json:"album"`          // Prefered
	Playlist      string `json:"playlist"`       // Might be an album playlist
	PlaylistTitle string `json:"playlist_title"` // Same here

	ReleaseYear int    `json:"release_year"` // Prefered
	UploadDate  string `json:"upload_date"`  // Take year from here...
	ReleaseDate string `json:"release_date"` // ...or from here

	WebpageURL string `json:"webpage_url"` // This usually shouldn't be empty
}

func (i *info) Webpage(originalURL string) string {
	if _, err := url.ParseRequestURI(i.WebpageURL); err == nil {
		return i.WebpageURL
	}
	return originalURL
}

// Year returns the year recorded in the `.info.json` file
func (i *info) Year() int {
	if i.ReleaseYear != 0 {
		return i.ReleaseYear
	}

	const dFormat = "20060102"

	d, err := time.Parse(dFormat, i.ReleaseDate)
	if err == nil {
		return d.Year()
	}

	d, err = time.Parse(dFormat, i.UploadDate)
	if err == nil {
		return d.Year()
	}
	return 0
}

// readInfoFile reads a youtube-dl `.info.json` file and extracts some information
func readInfoFile(p string) (i info, err error) {
	f, err := os.Open(p)
	if err != nil {
		return
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&i)

	return
}

// cascadeStrings returns the first string in `s` that isn't empty
func cascadeStrings(s ...string) string {
	for _, val := range s {
		if strings.TrimSpace(val) == "" {
			continue
		}
		return val
	}

	return ""
}

// cropMoveCover tries to create a squared cover image from the image located at `sourceFile`.
// If no squared image can be generated, no image will be generated.
func cropMoveCover(sourceFile, destination string) (err error) {
	f, err := os.Open(sourceFile)
	if err != nil {
		return
	}

	img, _, err := image.Decode(f)
	if err != nil {
		f.Close()
		return
	}

	err = f.Close()
	if err != nil {
		return
	}

	bounds := img.Bounds()

	// If we already have a square, we can just use the source file
	if bounds.Max.X == bounds.Max.Y {
		return os.Rename(sourceFile, destination)
	}

	// Use the smaller dimension for cutting off
	smallerOne := bounds.Max.X
	if bounds.Max.Y < smallerOne {
		smallerOne = bounds.Max.Y
	}

	// Basically take the middle square. This works e.g. with youtube music video thumbnails
	defaultCrop := image.Rect(bounds.Max.X/2-smallerOne/2, 0, bounds.Max.X/2+smallerOne/2, bounds.Max.Y)

	type SubImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	subImg, ok := img.(SubImager)

	// if we cannot crop, we won't use an image at all
	if !ok {
		return fmt.Errorf("cannot crop image")
	}

	croppedImg := subImg.SubImage(defaultCrop)

	ext := strings.ToUpper(strings.TrimPrefix(filepath.Ext(destination), "."))

	f, err = os.Create(destination)
	if err != nil {
		return
	}

	switch ext {
	case "JPG", "JPEG":
		err = jpeg.Encode(f, croppedImg, &jpeg.Options{
			Quality: 100, // We don't care about file size, only quality
		})
	case "PNG":
		err = png.Encode(f, croppedImg)
	}
	if err != nil {
		f.Close()
		return
	}

	return f.Close()
}
