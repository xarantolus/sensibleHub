package music

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"xarantolus/sensibleHub/store/config"
	"xarantolus/sensibleHub/store/file"

	"github.com/disintegration/imaging"
	"golang.org/x/sync/singleflight"
)

// MP3Version changes whenever MP3Path produces different files for the same
// entry, which invalidates every remembered MP3 size.
const MP3Version = 1

var mp3Group singleflight.Group

// MP3FilePath is where MP3Path keeps the generated MP3 of the song with this ID.
func MP3FilePath(id string) string {
	return filepath.Join("data", "songs", id, "latest.mp3")
}

// MP3Path returns the path for an mp3 file for this song, generating it if needed.
// Reusing a file marks it as recently used for the cleanup of generated files.
func (e *Entry) MP3Path(cfg config.Config) (p string, err error) {
	outName := MP3FilePath(e.ID)

	// Re-create this mp3 file if it doesn't exist or doesn't have the latest details
	if fi, ferr := os.Stat(outName); ferr == nil && fi.ModTime().After(e.LastEdit) {
		now := time.Now()
		if err := os.Chtimes(outName, now, now); err != nil && !os.IsNotExist(err) {
			log.Printf("Cannot mark %s as used: %v\n", outName, err)
		}
		return outName, nil
	}

	ap, err := filepath.Abs(e.AudioPath())
	if err != nil {
		return
	}

	_, err, _ = mp3Group.Do(e.ID, func() (res interface{}, err error) {
		defer mp3Group.Forget(e.ID)

		td, err := os.MkdirTemp("", "sh-mp3")
		if err != nil {
			return
		}
		defer os.RemoveAll(td)

		tempAudio := filepath.Join(td, "temp.mp3")

		cmd := exec.Command(cfg.Alternatives.FFmpeg, "-y", "-i", ap)

		if e.PictureData.Filename != "" {
			cover, err := coverJPEG(e.CoverPath(), cfg.Cover.MaxSize)
			if err != nil {
				log.Printf("Cannot embed cover of %s: %v\n", e.ID, err)
			} else {
				// Use the cover as second input (stdin) and tell ffmpeg that it is a JPEG stream
				cmd.Args = append(cmd.Args, "-probesize", "64M", "-f", "jpeg_pipe", "-i", "-")
				cmd.Stdin = bytes.NewReader(cover)

				cmd.Args = append(cmd.Args, "-c:v", "copy")

				cmd.Args = append(cmd.Args, "-map", "0", "-map", "1")

				// Make sure the image stream is recognized as cover image
				cmd.Args = append(cmd.Args,
					"-metadata:s:v", "title=Front cover",
					"-metadata:s:v", "comment=Cover (front)")

				cmd.Args = append(cmd.Args, "-flush_packets", "0")
			}
		}

		// Set whether to convert the stream or not - a mp3 stream can be kept
		// If this is not executed, then ffmpeg knows that it should be converted to MP3 because
		// tempAudio has an .mp3 extension
		if shouldCopy := strings.EqualFold(filepath.Ext(ap), ".MP3"); shouldCopy {
			// copy existing stream
			cmd.Args = append(cmd.Args, "-c:a", "copy")
		}

		// set start and end time depending on audio settings
		if e.AudioSettings.Start != -1 {
			cmd.Args = append(cmd.Args, "-ss", strconv.FormatFloat(e.AudioSettings.Start, 'f', 3, 64))
		}
		if e.AudioSettings.End != -1 {
			cmd.Args = append(cmd.Args, "-to", strconv.FormatFloat(e.AudioSettings.End, 'f', 3, 64))
		}

		cmd.Args = append(cmd.Args, "-id3v2_version", "3", "-write_id3v1", "1")

		// now set all kinds of metadata, e.g. like key="something"
		if have(&e.MusicData.Title) {
			cmd.Args = append(cmd.Args, "-metadata", "title="+e.MusicData.Title)
		}
		if have(&e.MusicData.Album) {
			cmd.Args = append(cmd.Args, "-metadata", "album="+e.MusicData.Album)
		}
		if have(&e.MusicData.Artist) {
			cmd.Args = append(cmd.Args, "-metadata", "artist="+e.MusicData.Artist)
		}
		if e.MusicData.Year != nil {
			cmd.Args = append(cmd.Args, "-metadata", "date="+strconv.Itoa(*e.MusicData.Year))
		}

		cmd.Args = append(cmd.Args,
			"-hide_banner", // don't show the ffmpeg banner, it's unnecessary noise for potential error output

			// Leave out the encoder version tags, so the file (and the size FTP
			// clients remember) stays the same across ffmpeg upgrades.
			"-fflags", "+bitexact", "-flags:a", "+bitexact",

			// Don't write a duration header to the output file.
			// This causes media players to display the correct duration.
			// https://superuser.com/questions/607703/wrong-audio-duration-with-ffmpeg
			"-write_xing", "0",

			// Write everything to tempAudio
			tempAudio)

		// An edit while ffmpeg runs must leave the result outdated.
		started := time.Now()

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		// Generate / Convert audio file
		err = cmd.Run()
		if err != nil {
			err = fmt.Errorf("%s\n\nStderr:%s", err.Error(), stderr.String())
			log.Println("Error while running ffmpeg for mp3 generation:", err.Error())
			return
		}
		// if everything goes right, we can now move it to its destination
		err = file.Move(tempAudio, outName)
		if err != nil {
			return
		}
		err = os.Chtimes(outName, started, started)
		if err != nil {
			return
		}

		return outName, nil
	})
	if err != nil {
		return "", err
	}

	return outName, nil
}

// coverJPEG returns the cover at path as JPEG no wider than maxSize (if
// positive). Some players only support JPEG covers, so PNGs are converted too.
// Stored covers usually qualify already and are passed through unchanged.
func coverJPEG(path string, maxSize int) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	tooLarge := maxSize > 0 && cfg.Width > maxSize
	if format == "jpeg" && !tooLarge {
		return b, nil
	}

	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	if tooLarge {
		img = imaging.Resize(img, maxSize, 0, imaging.Lanczos)
	}

	var out bytes.Buffer
	if err := jpeg.Encode(&out, img, &jpeg.Options{Quality: 95}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func have(s *string) bool {
	return strings.TrimSpace(*s) != ""
}
