package store

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
	"time"
	"xarantolus/sensiblehub/store/music"

	"github.com/bogem/id3v2"
)

// ImportFiles imports files from the given directory. It tries to get as much metadata as possible
func (m *Manager) ImportFiles(directory string) (err error) {
	return filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		e, err := m.importFile(path, info)
		if err != nil {
			log.Printf("Error while importing %s: %s\n", path, err.Error())
			return nil
		}

		log.Printf("[Import] Added %s\n", e.SongName())
		return nil
	})
}

func (m *Manager) importFile(musicFile string, info os.FileInfo) (e *music.Entry, err error) {

	var md = music.MusicData{}

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
		md.Title = filepath.Base(musicFile)
	}

	// extract duration
	md.Duration, err = getAudioDuration(musicFile)
	if err != nil {
		return
	}
	if md.Duration < 1 {
		err = fmt.Errorf("Music file too short")
		return
	}

	var picBuf bytes.Buffer

	// try to extract image from the file
	cmd := exec.Command("ffmpeg", "-i", musicFile, "-an", "-vcodec", "copy", "-f", "image2pipe", "pipe:1")
	cmd.Stdout = &picBuf
	ffmpegErr := cmd.Run()
	if ffmpegErr != nil {
		picBuf.Reset()
	}

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

	if picBuf.Len() > 0 {
		err = CropCover(ioutil.NopCloser(&picBuf), "", e.CoverPath())
	} else {
		e.PictureData.Filename = ""
	}
	if ffmpegErr != nil || err != nil {
		e.PictureData.Filename = ""
	}

	err = os.Rename(musicFile, e.AudioPath())
	if err != nil {
		return
	}

	err = m.Add(e)
	if err != nil {
		return nil, err
	}
	return e, nil
}
