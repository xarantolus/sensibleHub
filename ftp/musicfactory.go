package ftp

import (
	"strings"
	"xarantolus/sensiblehub/store"

	"github.com/goftp/server"
)

// musicDriverFactory is the ftp driver factory for this program.
// It implements the server.DriverFactory
type musicDriverFactory struct {
}

func (m *musicDriverFactory) NewDriver() (server.Driver, error) {
	entries := store.M.AllEntries()

	d := &musicDriver{
		Artists: make(map[string]Album),
	}

	// Create the virutal file system
	var normalizedArtists = make(map[string]string)

	for _, e := range entries {
		// Entries that should not be synced will not appear in the listing
		if !e.SyncSettings.Should {
			continue
		}

		art := store.CleanName(e.Artist())
		artistName, ok := normalizedArtists[strings.ToUpper(art)]
		if !ok {
			normalizedArtists[strings.ToUpper(e.Artist())] = e.Artist()
			artistName = art
		}

		_, ok = d.Artists[artistName]
		if !ok {
			d.Artists[artistName] = make(Album)
		}

		aname := store.CleanName(e.AlbumName())

		d.Artists[artistName][aname] = append(d.Artists[artistName][aname], fileInfoFromEntry(e))
	}

	return d, nil
}
