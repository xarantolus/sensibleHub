package store

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"xarantolus/sensibleHub/store/file"
	"xarantolus/sensibleHub/store/music"

	"github.com/bogem/id3v2"
	"github.com/vitali-fedulov/images"
)

// ImportFiles imports files from the given directory. It tries to get as much metadata as possible
func (m *Manager) ImportFiles(directory string) (err error) {
	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		_, err = m.ImportFile(path, info)
		if err != nil {
			log.Printf("[Import] Error while importing %s: %s\n", path, err.Error())
			return nil
		}

		return nil
	})

	// If "directory" doesn't exist, we don't care
	if err != nil && os.IsNotExist(err) {
		err = nil
	}

	return
}

// ImportFile imports a file from the given path. `info` is optional
func (m *Manager) ImportFile(musicFile string, info os.FileInfo) (e *music.Entry, err error) {
	if info == nil {
		info, err = os.Stat(musicFile)
		if err != nil {
			return
		}
	}

	md := music.MusicData{}

	tag, err := id3v2.Open(musicFile, id3v2.Options{Parse: true})
	if err == nil {
		md.Title = tag.Title()
		md.Artist = tag.Artist()
		md.Album = tag.Album()

		y, nerr := strconv.Atoi(tag.Year())
		if nerr == nil {
			md.Year = &y
		}

		// Extracting the image doesn't really seem to work.
		// We'll do it with ffmpeg after this file is closed

		err = tag.Close()
		if err != nil {
			return
		}
	}

	if strings.TrimSpace(md.Title) == "" {
		b := filepath.Base(musicFile)
		if strings.Contains(b, " - ") {
			split := strings.Split(strings.TrimSuffix(b, filepath.Ext(b)), " - ")
			if len(split) == 2 {
				md.Artist, md.Title = split[0], split[1]
			}
		} else {
			md.Title = b
		}
	}

	// extract duration
	md.Duration, err = m.getAudioDuration(musicFile)
	if err != nil {
		return
	}
	if md.Duration < 1 {
		err = fmt.Errorf("Music file too short")
		return
	}

	var picBuf bytes.Buffer

	ex := filepath.Ext(musicFile)
	f := strings.TrimSuffix(musicFile, ex) + ".temp" + ex

	// try to extract image from the file: https://superuser.com/a/1328212
	cmd := exec.Command(m.cfg.Alternatives.FFmpeg, "-y", "-i", musicFile, "-map", "0:v", "-map", "-0:V", "-c", "copy", "-f", "image2pipe", "pipe:1")
	cmd.Stdout = &picBuf
	ffmpegErr := cmd.Run()
	if ffmpegErr != nil {
		picBuf.Reset()
	}

	// Try to remove metadata, mostly the cover image, as it will take space and will never be needed
	cmd = exec.Command(m.cfg.Alternatives.FFmpeg, "-i", musicFile, "-y", "-map_metadata", "-1", "-vn", "-acodec", "copy", f)
	err = cmd.Run()
	if err != nil {
		return
	}

	oldmfile := musicFile
	defer func() {
		if err == nil {
			err = os.Remove(oldmfile)
		}
	}()
	musicFile = f

	m.SongsLock.Lock()
	defer m.SongsLock.Unlock()

	now := time.Now()
	e = &music.Entry{
		ID:        m.generateID(),
		SourceURL: "Import",

		LastEdit: now,
		Added:    now,

		// Assume that songs should be synced by default
		SyncSettings: music.SyncSettings{
			Should: true,
		},
		FileData: music.FileData{
			Filename: "original" + strings.ToLower(filepath.Ext(musicFile)),
			Size:     info.Size(),
		},
		AudioSettings: music.AudioSettings{
			Start: -1,
			End:   -1,
		},
		MusicData: md,
		PictureData: music.PictureData{
			Filename: "cover.jpg",
		},
	}

	songDir := fmt.Sprintf(songDirTemplate, e.ID)
	err = os.MkdirAll(songDir, 0o644)
	if err != nil {
		return
	}

	// Delete songDir if something goes wrong
	defer func() {
		if err != nil {
			_ = os.RemoveAll(songDir)
		}
	}()

	if picBuf.Len() > 0 {
		err = cropCover(&picBuf, "", e.CoverPath())
	} else {
		e.PictureData.Filename = ""
	}
	if ffmpegErr != nil || err != nil {
		e.PictureData.Filename = ""
	}

	if e.PictureData.Filename != "" {
		hex, _ := music.CalculateDominantColor(e.CoverPath())
		e.PictureData.DominantColorHEX = music.Color(hex)

		i, err := images.Open(e.CoverPath())
		if err == nil {
			e.PictureData.Size = i.Bounds().Dx()
		}
	}

	err = file.Move(musicFile, e.AudioPath())
	if err != nil {
		return
	}

	err = m.Add(e)
	if err != nil {
		return nil, err
	}

	log.Printf("[Import] Added %s\n", e.SongName())

	return e, nil
}
