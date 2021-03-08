package ftp

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"xarantolus/sensibleHub/store"

	"goftp.io/server"
)

var (
	errReadOnly = fmt.Errorf("write/delete not allowed")
	errNotFound = fmt.Errorf("file/directory not found")
)

// The different depths an FTP path can have
const (
	rootDirectory   int = 0
	artistDirectory int = 1
	albumDirectory  int = 2
	songFile        int = 3
)

type musicDriver struct {
	Artists map[string]Album
}

// Album is part of the virtual file system
type Album map[string][]*fileInfo

func (m *musicDriver) Init(c *server.Conn) {
	log.Println("[FTP] Connected client from", c.RemoteAddr())
}

func (m *musicDriver) Stat(path string) (fi server.FileInfo, err error) {
	split := splitPath(path)

	switch len(split) {
	case rootDirectory:
		// this one pretends to be a good root directory
		return &artistAlbumInfo{
			Album: "/",
		}, nil
	case artistDirectory:
		return &artistAlbumInfo{
			Artist:   split[0],
			isArtist: true,
		}, nil
	case albumDirectory:
		return &artistAlbumInfo{
			Album: split[1],
		}, nil
	case songFile:
		{
			// Stat the actual mp3 files
			a, ok := m.Artists[split[0]]
			if ok {
				if files, ok := a[split[1]]; ok {
					for _, fi := range files {
						if fi.Name() == split[2] {
							return fi, nil
						}
					}
				}
			}
		}
	default:
		err = errNotFound
	}

	return
}

func (m *musicDriver) ChangeDir(path string) (err error) {
	split := splitPath(path)

	if len(split) == artistDirectory {
		// Exists or is root
		if _, ok := m.Artists[split[0]]; split[0] == "" || ok {
			return nil
		}
	} else if len(split) == albumDirectory {
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
func (m *musicDriver) GetFile(path string, offset int64) (i int64, content io.ReadCloser, err error) {
	split := splitPath(path)
	if len(split) != songFile {
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
			p, er := file.MP3Path(store.M.GetConfig())
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

			if offset > 0 {
				_, err = f.Seek(offset, io.SeekStart)
				if err != nil {
					f.Close()
					return 0, nil, err
				}
			}

			return fi.Size() - offset, f, nil
		}
	}

	// seems like we couldn't find the file
	err = errNotFound
	return
}

func (m *musicDriver) ListDir(path string, f func(server.FileInfo) error) (err error) {
	split := splitPath(path)

	switch len(split) {
	case artistDirectory:
		if split[0] == "" {
			// We want them in alphabetical order
			var keys []string
			for k := range m.Artists {
				keys = append(keys, k)
			}

			sort.Strings(keys)

			// List all artists in root directory
			for _, artist := range keys {
				err = f(&artistAlbumInfo{
					Artist:   artist,
					isArtist: true,
				})
				if err != nil {
					return
				}
			}
		} else {
			var keys []string
			for k := range m.Artists[split[0]] {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			// List all albums in artist directory
			for _, album := range keys {
				err = f(&artistAlbumInfo{
					Album: album,
				})
				if err != nil {
					return
				}
			}
		}
	case albumDirectory:
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

// PutFile implements putting files on the server while also importing them. That way, you can use FTP to import your music library
func (m *musicDriver) PutFile(p string, f io.Reader, overwrite bool) (n int64, err error) {
	dest := filepath.Join("import", store.CleanName(path.Base(strings.ReplaceAll(p, "\\", "/"))))

	// Try to create the import directory, but ignore if it doesn't work.
	_ = os.MkdirAll(filepath.Dir(dest), os.ModePerm)

	d, err := os.Create(dest)
	if err != nil {
		return
	}

	n, err = io.Copy(d, f)
	if err != nil {
		d.Close()

		os.Remove(d.Name())

		return
	}

	err = d.Close()
	if err != nil {
		return
	}

	_, err = store.M.ImportFile(d.Name(), nil)
	if err != nil {
		log.Println("[Import] Error while importing file from FTP:", err.Error())
	}

	return
}

// Methods that are not implemented as the server is supposed to be mostly read-only:

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
