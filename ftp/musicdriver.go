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

	"goftp.io/server/v2"
)

var (
	errReadOnly = fmt.Errorf("write/delete not allowed")
	errNotFound = fmt.Errorf("file/directory not found")
)

// The different depths an FTP path can have
const (
	artistDirectory int = 1
	albumDirectory  int = 2
	songFile        int = 3
)

// musicDriver serves the library snapshot of each session (see treeSnapshot)
type musicDriver struct {
	manager *store.Manager
}

func (m *musicDriver) Stat(ctx *server.Context, path string) (os.FileInfo, error) {
	artists := sessionTree(ctx)
	split := splitPath(path)

	switch len(split) {
	case artistDirectory:
		if split[0] == "" {
			// this one pretends to be a good root directory
			return &artistAlbumInfo{Album: "/"}, nil
		}
		if _, ok := artists[split[0]]; ok {
			return &artistAlbumInfo{Artist: split[0], isArtist: true}, nil
		}
	case albumDirectory:
		if _, ok := artists[split[0]][split[1]]; ok {
			return &artistAlbumInfo{Album: split[1]}, nil
		}
	case songFile:
		if fi := findSong(artists, split); fi != nil {
			return fi, nil
		}
	}

	return nil, errNotFound
}

func findSong(artists tree, split []string) *fileInfo {
	for _, fi := range artists[split[0]][split[1]] {
		if fi.Name() == split[2] {
			return fi
		}
	}
	return nil
}

// GetFile returns the MP3 file at path, starting at offset
func (m *musicDriver) GetFile(ctx *server.Context, path string, offset int64) (int64, io.ReadCloser, error) {
	split := splitPath(path)
	if len(split) != songFile {
		return 0, nil, errNotFound
	}

	file := findSong(sessionTree(ctx), split)
	if file == nil {
		return 0, nil, errNotFound
	}

	p, err := m.manager.MP3Path(file.Entry)
	if err != nil {
		return 0, nil, err
	}

	f, err := os.Open(p)
	if err != nil {
		return 0, nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return 0, nil, err
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

func (m *musicDriver) ListDir(ctx *server.Context, path string, f func(os.FileInfo) error) (err error) {
	artists := sessionTree(ctx)
	split := splitPath(path)

	switch len(split) {
	case artistDirectory:
		if split[0] == "" {
			// We want them in alphabetical order
			var keys []string
			for k := range artists {
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
			for k := range artists[split[0]] {
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
		// List all songs in album directory
		for _, song := range artists[split[0]][split[1]] {
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
func (m *musicDriver) PutFile(_ *server.Context, p string, f io.Reader, offset int64) (n int64, err error) {
	if offset > 0 {
		return 0, fmt.Errorf("resuming uploads is not supported")
	}

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

	_, err = m.manager.ImportFile(d.Name(), nil)
	if err != nil {
		log.Println("[Import] Error while importing file from FTP:", err.Error())
	}

	return
}

// Methods that are not implemented as the server is supposed to be mostly read-only:

func (m *musicDriver) DeleteDir(*server.Context, string) error {
	return errReadOnly
}

func (m *musicDriver) DeleteFile(*server.Context, string) error {
	return errReadOnly
}

func (m *musicDriver) Rename(*server.Context, string, string) error {
	return errReadOnly
}

func (m *musicDriver) MakeDir(*server.Context, string) error {
	return errReadOnly
}
