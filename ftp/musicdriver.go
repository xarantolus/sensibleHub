package ftp

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/goftp/server"
)

var (
	errReadOnly = fmt.Errorf("write/delete not allowed")
	errNotFound = fmt.Errorf("file/directory not found")
)

type musicDriver struct {
	Artists map[string]Album
}

// Album is part of the virutal file system
type Album map[string][]*fileInfo

func (m *musicDriver) Init(c *server.Conn) {
	log.Println("Connected client from", c.PublicIp())
	return
}

func (m *musicDriver) Stat(path string) (fi server.FileInfo, err error) {
	split := splitPath(path)

	switch len(split) {
	case 0:
		// this one pretends to be a good root directory
		return &artistAlbumInfo{
			Album: "/",
		}, nil
	case 1:
		return &artistAlbumInfo{
			Artist:   split[0],
			isArtist: true,
		}, nil
	case 2:
		return &artistAlbumInfo{
			Album: split[1],
		}, nil
	default:
		err = errNotFound
	}

	return
}

func (m *musicDriver) ChangeDir(path string) (err error) {
	split := splitPath(path)

	if len(split) == 1 {
		// Exists or is root
		if _, ok := m.Artists[split[0]]; split[0] == "" || ok {
			return nil
		}
	} else if len(split) == 2 {
		a, ok := m.Artists[split[0]]
		if ok {
			if _, ok := a[split[1]]; ok {
				return nil
			}
		}
	}

	return errNotFound
}

// params  - path
// returns - a string containing the file data to send to the client
func (m *musicDriver) GetFile(path string, n int64) (i int64, content io.ReadCloser, err error) {
	split := splitPath(path)
	if len(split) != 3 {
		err = errNotFound
		return
	}

	art, ok := m.Artists[split[0]]
	if !ok {
		err = errNotFound
		return
	}

	al, ok := art[split[1]]
	if !ok {
		err = errNotFound
		return
	}

	for _, file := range al {
		// Serve the file with the given name
		if file.Name() == split[2] {
			p, er := file.MP3Path()
			if er != nil {
				err = er
				return
			}

			f, er := os.Open(p)
			if er != nil {
				err = er
				return
			}

			fi, er := f.Stat()
			if er != nil {
				f.Close()
				err = er
				return
			}

			return fi.Size(), f, nil
		}
	}

	// seems like we couldn't find the file
	err = errNotFound
	return
}

func (m *musicDriver) ListDir(path string, f func(server.FileInfo) error) (err error) {
	split := splitPath(path)

	switch len(split) {
	case 1:
		if split[0] == "" {
			// List all artists in root directory
			for artist := range m.Artists {
				err = f(&artistAlbumInfo{
					Artist:   artist,
					isArtist: true,
				})
				if err != nil {
					return
				}
			}
		} else {
			// List all albums in artist directory
			for album := range m.Artists[split[0]] {
				err = f(&artistAlbumInfo{
					Album: album,
				})
				if err != nil {
					return
				}
			}
		}
		break
	case 2:
		// List all albums in artist directory
		for _, song := range m.Artists[split[0]][split[1]] {
			err = f(song)
			if err != nil {
				return
			}
		}
	default:
		err = errNotFound
	}

	return
}

func splitPath(p string) []string {
	return strings.Split(strings.Trim(p, "/"), "/")
}

// Methods that are not implemented as the server is supposed to be read-only:

func (m *musicDriver) DeleteDir(path string) (err error) {
	return errReadOnly
}

func (m *musicDriver) DeleteFile(string) (err error) {
	return errReadOnly
}

func (m *musicDriver) Rename(src string, dest string) (err error) {
	return errReadOnly
}

func (m *musicDriver) MakeDir(path string) (err error) {
	return errReadOnly
}

func (m *musicDriver) PutFile(path string, f io.Reader, overwrite bool) (n int64, err error) {
	return 0, errReadOnly
}
