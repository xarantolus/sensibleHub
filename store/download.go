package store

import (
	"context"
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
	"xarantolus/sensibleHub/store/music"

	"github.com/vitali-fedulov/images"
)

const (
	songDirTemplate = "data/songs/%s"
)

// Download downloads the song from the given URL using youtube-dl and saves it to the appropriate directory
func (m *Manager) download(url string) (err error) {
	log.Println("[Download] Start downloading", url)

	defer func() {
		if err != nil {
			log.Printf("[Download] Error while downloading %s: %s\n", url, err.Error())
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

	var cmdCtx, cancel = context.WithCancel(context.Background())

	m.downloadContextLock.Lock()
	m.downloadContext = cmdCtx
	m.downloadCancelFunc = cancel
	m.currentDownload = url
	m.downloadContextLock.Unlock()

	// Setup youtube-dl command and run it
	cmd := exec.CommandContext(cmdCtx, "youtube-dl", "--write-info-json", "--write-thumbnail", "-f", "bestaudio/best", "--max-downloads", "1", "--no-playlist", "-x", "-o", "song.%(ext)s")
	cmd.Dir = tmpDir

	// when searching for a specific song, we want to reject Instrumental versions.
	// This leads to youtube-dl selecting the second search result if the instrumental is first
	// Don't add it when we explicitly want them though
	if strings.Contains(url, "youtube.com/results?search_query") && !strings.Contains(strings.ToUpper(url), "INSTRUMENTAL") {
		cmd.Args = append(cmd.Args, "--reject-title", "(Instrumental)")
	}

	cmd.Args = append(cmd.Args, url)

	out, err := cmd.CombinedOutput()

	m.downloadContextLock.Lock()
	m.downloadContext = nil
	m.downloadCancelFunc = nil
	m.currentDownload = ""
	m.downloadContextLock.Unlock()

	// "exit status 101" means that the download limit has been reached (because of --max-downloads). We should just take this one song then, it's fine
	if err != nil && err.Error() != "exit status 101" {
		return fmt.Errorf("Error while running youtube-dl: %s\nOutput: %s", err.Error(), string(out))
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
		case "JPG", "JPEG", "PNG":
			thumbPath = path
		case "WEBP", "GIF", "TIFF", "RAW", "BMP":
			// These image formats are not supported and must be converted
			// The output format could be either PNG or JPG, but PNG is lossless
			outpath := filepath.Join(tmpDir, "song.png")
			err = exec.Command("ffmpeg", "-y", "-i", path, outpath).Run()
			if err != nil {
				return nil // Ignore error
			}
			thumbPath = outpath
		case "TEMP", "TMP":
			{
				// Do nothing; however, it could the sign of an error.
				// There is only a problem if audioPath == "", but that is handeled below
			}
		default:
			// Assume this is the audio file. This is not exact, as there might be some
			// image format that is not recognized above, which is why we need to have a whitelist
			if musicExtensions[strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))] {
				audioPath = path
				audioSize = info.Size()
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("%s\nyoutube-dl Output: %s", err.Error(), string(out))
	}

	if audioPath == "" {
		// Well, what can we do?
		return fmt.Errorf("invalid empty audio path, it seems like no audio was downloaded\nyoutube-dl Output: %s", string(out))
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
		log.Println("[Download] Error while reading info file: ", jsonErr.Error())
		// Continue without data
	}

	// For songs with multiple artists, there is a comma-separated list
	title, artist := cascadeStrings(minfo.Track, minfo.Title, filepath.Base(minfo.Filename)), cascadeStrings(minfo.Artist, minfo.Creator, strings.TrimSuffix(minfo.Uploader, " - Topic"))
	album := cascadeStrings(minfo.Album, minfo.Playlist, minfo.PlaylistTitle)

	// Split artists so we only have *one* in the artist field
	artists := strings.Split(artist, ", ")
	if len(artists) > 1 {
		artist = artists[0]

		feats := generateFeats(artists[1:])

		if !strings.HasSuffix(title, feats) {
			title += " " + feats
		}
	}

	now := time.Now()

	// Technically, we would need to lock `m.generateID`, but that doesn't play
	// nicely with `m.betterCover` below. Not great, but generating two
	// IDs at the same time just doesn't happen with downloads as they are sequential.
	// There is the risk that files are imported over FTP while downloading something new,
	// but the possibility of generating the same id is low
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
			Title:    title,
			Artist:   artist,
			Album:    album,
			Year:     minfo.Year(),
			Duration: dur,
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

	if thumbPath != "" {
		e.PictureData.Filename = "cover" + strings.ToLower(filepath.Ext(thumbPath))

		destPath := filepath.Join(songDir, e.PictureData.Filename)

		err = cropMoveCover(thumbPath, destPath)
		if err == nil {
			// See if we already have a higher-quality version of this cover
			better, err := m.betterCover(artist, album, destPath)
			if err == nil {
				// So, there is a better version.
				// Just copy it, if it fails then we ignore it.
				_ = copyOverwrite(better, destPath)
			}
		} else {
			e.PictureData.Filename = ""
		}
	}

	if m.cfg.AllowExternal.Apple {
		externalSongData, err := music.SearchITunes(title, album, artist, filepath.Ext(thumbPath))
		if err == nil && externalSongData.Artwork != nil {
			writeNewImage := func() {
				tmp, err := encodeImageToTemp(externalSongData.Artwork, e.PictureData.Filename)
				if err != nil {
					e.PictureData.Filename = ""
					return
				}

				// Everything went well, we can move it to its real destination
				destPath := filepath.Join(songDir, e.PictureData.Filename)

				err = os.Rename(tmp, destPath)
				if err != nil {
					e.PictureData.Filename = ""
					return
				}
			}

			if e.PictureData.Filename == "" {
				// Just encode our new image, no need to compare (as we have no image)
				e.PictureData.Filename = "cover." + externalSongData.ArtworkExtension
				writeNewImage()
			} else {
				// check if the downloaded cover is better than the one we already have
				currentCoverPath := filepath.Join(songDir, e.PictureData.Filename)

				currCov, err := images.Open(currentCoverPath)
				if err == nil {
					currSize := currCov.Bounds()
					newSize := externalSongData.Artwork.Bounds()

					// If it's not a square, we don't care
					if newSize.Max.X == newSize.Max.Y {
						// We only compare width, but height should be the same as width
						if currSize.Dx() < newSize.Dx() {
							// Now the new image is larger. Write it to the file
							writeNewImage()
						}
					}
				}
			}
		}
	}

	// Move all kinds of files - this may not work on all platforms as they aren't in the same directory
	if jsonErr == nil {
		err = os.Rename(jsonPath, filepath.Join(songDir, e.MetaFile.Filename))
		if err != nil {
			return
		}
	} else {
		e.MetaFile.Filename = ""
	}

	err = os.Rename(audioPath, filepath.Join(songDir, e.FileData.Filename))
	if err != nil {
		return
	}

	if e.PictureData.Filename != "" {
		hex, err := music.CalculateDominantColor(e.CoverPath())
		if err == nil {
			e.PictureData.DominantColorHEX = hex
		}
	}

	m.SongsLock.Lock()
	defer m.SongsLock.Unlock()

	err = m.Add(e)
	if err != nil {
		return
	}

	log.Printf("[Download] Added %s\n", e.SongName())

	return nil
}

// AbortDownload cancels the currently running download, returning an error if no download is running
func (m *Manager) AbortDownload() (err error) {
	m.downloadContextLock.Lock()
	defer m.downloadContextLock.Unlock()

	if m.downloadCancelFunc == nil {
		return fmt.Errorf("Cannot cancel downloading as no download is running")
	}

	m.downloadCancelFunc()

	return nil
}

// IsDownloading returns true if a download is running
func (m *Manager) IsDownloading() (string, bool) {
	m.downloadContextLock.Lock()
	defer m.downloadContextLock.Unlock()

	return m.currentDownload, m.downloadCancelFunc != nil
}

// info is the struct that stores data that can be read from a typical youtube-dl `.info.json` file
// If some fields duplicate information, only one of them will be used;
// however, there are clear preferences on which fields should be used
type info struct {
	Track    string `json:"track"`     // Preferred
	Title    string `json:"title"`     // Fallback
	Filename string `json:"_filename"` // Fallback

	Artist   string `json:"artist"`   // Preferred
	Creator  string `json:"creator"`  // Maybe
	Uploader string `json:"uploader"` // If nothing else has info

	Album         string `json:"album"`          // Preferred
	Playlist      string `json:"playlist"`       // Might be an album playlist
	PlaylistTitle string `json:"playlist_title"` // Same here

	ReleaseYear int    `json:"release_year"` // Preferred
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
func (i *info) Year() *int {
	if i.ReleaseYear != 0 {
		return &i.ReleaseYear
	}

	// this date format is typically used in the info file: YYYYMMDD
	const dFormat = "20060102"

	d, err := time.Parse(dFormat, i.ReleaseDate)
	if err == nil {
		y := d.Year()
		return &y
	}

	d, err = time.Parse(dFormat, i.UploadDate)
	if err == nil {
		y := d.Year()
		return &y
	}
	return nil
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

func generateFeats(artists []string) (out string) {
	switch len(artists) {
	case 1:
		return fmt.Sprintf("(feat. %s)", artists[0])
	case 2:
		return fmt.Sprintf("(feat. %s & %s)", artists[0], artists[1])
	case 3:
		return fmt.Sprintf("(feat. %s, %s & %s)", artists[0], artists[1], artists[2])
	default:
		return
	}
}
