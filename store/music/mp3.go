package music

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"xarantolus/sensibleHub/store/config"

	"github.com/nfnt/resize"
	"golang.org/x/sync/singleflight"
)

var mp3Group singleflight.Group

// MP3Path returns the path for an mp3 file for this song. This might take some time
func (e *Entry) MP3Path(cfg config.Config) (p string, err error) {
	outName := filepath.Join("data", "songs", e.ID, "latest.mp3")

	// Re-create this mp3 file if it doesn't exist or doesn't have the latest details
	if fi, ferr := os.Stat(outName); !os.IsNotExist(ferr) && fi.ModTime().After(e.LastEdit) {
		return outName, ferr
	}

	ap, err := filepath.Abs(e.AudioPath())
	if err != nil {
		return
	}

	_, err, _ = mp3Group.Do(e.ID, func() (res interface{}, err error) {
		defer mp3Group.Forget(e.ID)

		td, err := ioutil.TempDir("", "sh-mp3")
		if err != nil {
			return
		}
		defer os.RemoveAll(td)

		tempAudio := filepath.Join(td, "temp.mp3")

		var outbuf bytes.Buffer

		cmd := exec.Command(cfg.Alternatives.FFmpeg, "-y", "-i", ap)

		// If we have a cover image, we add it
		if e.PictureData.Filename != "" {
			b, err := ioutil.ReadFile(e.CoverPath())
			if err != nil {
				return nil, err
			}

			var coverBuf bytes.Buffer

			coverImg, _, err := image.Decode(bytes.NewBuffer(b))
			if err == nil {
				maxImageSize := cfg.Cover.MaxSize
				if maxImageSize > 0 && coverImg.Bounds().Dx() > maxImageSize {
					resized := resize.Resize(uint(maxImageSize), 0, coverImg, resize.MitchellNetravali)
					coverImg = resized
				}

				// Yes, all png images are also converted to JPEG.
				// This might not be optimal for all use cases, but there are *some* audio players that support *only* JPEG
				err = jpeg.Encode(&coverBuf, coverImg, &jpeg.Options{Quality: 100})
				if err != nil {
					coverBuf.Reset()
				}
			} else {
				log.Println("error while decoding image:", err.Error())
				b = nil
			}

			// If everything worked, we can now add the cover image command-line arguments
			if len(b) > 0 && err == nil {
				// Use the buffer as cover input (stdin) and tell ffmpeg that it is a JPEG stream
				cmd.Args = append(cmd.Args, "-probesize", "64M", "-f", "jpeg_pipe", "-i", "-")
				cmd.Stdin = &coverBuf

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

			// Don't write a duration header to the output file.
			// This causes media players to display the correct duration.
			// https://superuser.com/questions/607703/wrong-audio-duration-with-ffmpeg
			"-write_xing", "0",

			// Write everything to tempAudio
			tempAudio)

		cmd.Stderr = os.Stderr // &outbuf // Collect error output in case we need it for error logging

		// Generate / Convert audio file
		err = cmd.Run()
		if err != nil {
			err = fmt.Errorf("%s\n\nStderr:%s", err.Error(), outbuf.String())
			log.Println("Error while running ffmpeg for mp3 generation:", err.Error())
			return
		}
		// if everything goes right, we can now move it to its destination
		err = os.Rename(tempAudio, outName)
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

func have(s *string) bool {
	return strings.TrimSpace(*s) != ""
}
