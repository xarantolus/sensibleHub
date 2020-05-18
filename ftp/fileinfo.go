package ftp

import (
	"os"
	"time"
	"xarantolus/sensiblehub/store/music"
)

// implements server.FileInfo
type fileInfo struct {
	music.Entry
}

func fileInfoFromEntry(e music.Entry) *fileInfo {
	return &fileInfo{e}
}

func (f *fileInfo) Group() string {
	return "music"
}

func (f *fileInfo) Owner() string {
	return "human"
}

func (f *fileInfo) Name() string {
	// base name
	return cleanName(f.Entry.SongName() + ".mp3")
}

func (f *fileInfo) Size() int64 {
	// We might not know the mp3 size just yet
	return 4096
}

func (f *fileInfo) Mode() os.FileMode {
	return os.ModePerm
}

func (f *fileInfo) ModTime() time.Time {
	return f.LastEdit
}

func (f *fileInfo) IsDir() bool {
	return false
}

func (f *fileInfo) Sys() interface{} {
	return nil
}

// implements server.FileInfo
type artistAlbumInfo struct {
	Artist, Album string

	isArtist bool
}

func (f *artistAlbumInfo) Group() string {
	return "music"
}

func (f *artistAlbumInfo) Owner() string {
	return "human"
}

func (f *artistAlbumInfo) Name() string {
	if f.isArtist {
		return f.Artist
	}
	return f.Album
}

func (f *artistAlbumInfo) Size() int64 {
	return 4096
}

func (f *artistAlbumInfo) Mode() os.FileMode {
	// From https://github.com/goftp/qiniu-driver/blob/master/fileinfo.go#L24-L29
	return os.ModeDir | os.ModePerm
}

func (f *artistAlbumInfo) ModTime() time.Time {
	return time.Time{}
}

// this is the most important bit
func (f *artistAlbumInfo) IsDir() bool {
	return true
}

func (f *artistAlbumInfo) Sys() interface{} {
	return nil
}
