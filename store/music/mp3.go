package music

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bogem/id3v2"
	"golang.org/x/sync/singleflight"
)

var (
	mp3Group singleflight.Group
)

// MP3Path returns the path for an mp3 file for this song. This might take some time
func (e *Entry) MP3Path() (p string, err error) {
	var outName = filepath.Join("data", "songs", e.ID, "latest.mp3")

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

		var cmd *exec.Cmd
		if strings.ToUpper(filepath.Ext(ap)) == ".MP3" {
			// remove existing metadata, keep mp3 stream
			cmd = exec.Command("ffmpeg", "-i", ap, "-map_metadata", "-1", "-c:a", "copy", "-map", "a")
		} else {
			// Convert the given audio to mp3
			cmd = exec.Command("ffmpeg", "-i", ap, "-map_metadata", "-1", "-f", "mp3")
		}

		// Audio settings: start and end
		if e.AudioSettings.Start != -1 {
			cmd.Args = append(cmd.Args, "-ss", strconv.FormatFloat(e.AudioSettings.Start, 'f', 3, 64))
		}

		if e.AudioSettings.End != -1 {
			cmd.Args = append(cmd.Args, "-to", strconv.FormatFloat(e.AudioSettings.End, 'f', 3, 64))
		}

		// Set some options and output file
		cmd.Args = append(cmd.Args,
			"-hide_banner",

			// Don't write a duration header to the output file.
			// This causes media players to display the correct duration.
			// https://superuser.com/questions/607703/wrong-audio-duration-with-ffmpeg
			"-write_xing", "0",

			tempAudio)
		cmd.Stderr = &outbuf

		err = cmd.Run()
		if err != nil {
			err = fmt.Errorf("%s\n\nStderr:%s", err.Error(), outbuf.String())
			log.Println("Error while running ffmpeg for mp3 generation:", err.Error())
			return
		}

		// And open it
		tag, err := id3v2.Open(tempAudio, id3v2.Options{Parse: false})
		if err != nil {
			return
		}

		// Now we edit its tags

		// Set artist
		if have(&e.MusicData.Artist) {
			tag.SetArtist(e.MusicData.Artist)
		}

		// Set title
		if have(&e.MusicData.Title) {
			tag.SetTitle(e.MusicData.Title)
		}

		if have(&e.MusicData.Album) {
			tag.SetAlbum(e.MusicData.Album)
		}

		if e.MusicData.Year != nil {
			tag.SetYear(strconv.Itoa(*e.MusicData.Year))
		}

		// Set artwork
		if have(&e.PictureData.Filename) {
			b, err := ioutil.ReadFile(e.CoverPath())
			if err != nil {
				tag.Close()
				return nil, err
			}

			// Other mime types don't work as only jpeg and png images are accepted
			mimeType := "image/jpeg"
			if strings.ToUpper(filepath.Ext(e.PictureData.Filename)) == ".PNG" {
				mimeType = "image/png"
			}

			pic := id3v2.PictureFrame{
				Encoding:    id3v2.EncodingUTF8,
				MimeType:    mimeType,
				PictureType: id3v2.PTFrontCover,
				Description: "Front cover",
				Picture:     b,
			}

			tag.AddAttachedPicture(pic)
		}

		// and save it
		err = tag.Save()
		if err != nil {
			tag.Close()
			return
		}

		err = tag.Close()
		if err != nil {
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
