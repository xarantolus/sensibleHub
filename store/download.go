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
)

const (
	songDirTemplate = "data/songs/%s"
)

// Download downloads the song from the given URL using youtube-dl and saves it to the appropriate directory
func (m *Manager) download(url string) (err error) {
	log.Println("Start downloading", url)
	defer func() {
		if err != nil {
			log.Printf("Error while downloading %s: %s\n", url, err.Error())
		}
	}()

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
		case "PNG":
			thumbPath = path
		case "TEMP", "TMP":
			{
				// Do nothing; however, it could the sign of an error.
				// There is only a problem if audioPath == "", but that is handeled below
			}
		default:
			// Assume this is the audio file.
			// Having only a whitelist of audio file formats would not be great as there are quite many
			// As long as it is supported by ffmpeg, we support it
			audioPath = path
			audioSize = info.Size()
		}

		return nil
	})
	if err != nil {
		return
	}

	if audioPath == "" {
		// Well, what can we do?
		return fmt.Errorf("invalid empty audio path, it seems like no audio was downloaded")
	}

	// the bad part about this is that still images also have a duration of 0
	dur, err := getAudioDuration(audioPath)
	if err != nil {
		return fmt.Errorf("cannot get audio duration: %w", err)
	}

	// this means that we have to assume. Also who would listen to a 0.5 seconds song?
	if dur < 1 {
		return fmt.Errorf("invalid audio (%s): duration too short", filepath.Base(audioPath))
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
			Filename: "original" + strings.ToLower(filepath.Ext(audioPath)),
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
			Filename: "cover" + strings.ToLower(filepath.Ext(thumbPath)),
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
	} else {
		e.MetaFile.Filename = ""
	}

	if thumbPath != "" {
		err = cropMoveCover(thumbPath, filepath.Join(songDir, e.PictureData.Filename))
		if err != nil {
			// reset image info
			e.PictureData.Filename = ""
		}
	} else {
		e.PictureData.Filename = ""
	}

	err = os.Rename(audioPath, filepath.Join(songDir, e.FileData.Filename))
	if err != nil {
		return
	}

	err = m.Add(e)
	if err != nil {
		return
	}

	log.Println("Download finished successfully")

	return nil
}

// info is the struct that stores data that can be read from a typical youtube-dl `.info.json` file
// If some fields duplicate information, only one of them will be used;
// however, there are clear preferences on which fields should be used
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

	// WebpageURL is used to "clean" the URL, e.g. to remove playlist parameters as they aren't used here
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

	// this date format is typically used in the info file: YYYYMMDD
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
