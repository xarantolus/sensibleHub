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

		art := cleanName(e.Artist())
		artistName, ok := normalizedArtists[strings.ToUpper(art)]
		if !ok {
			normalizedArtists[strings.ToUpper(e.Artist())] = e.Artist()
			artistName = art
		}

		_, ok = d.Artists[artistName]
		if !ok {
			d.Artists[artistName] = make(Album)
		}

		aname := cleanName(e.AlbumName())

		d.Artists[artistName][aname] = append(d.Artists[artistName][aname], fileInfoFromEntry(e))
	}

	return d, nil
}

func cleanName(n string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
			return r
		}

		if r >= '0' && r <= '9' {
			return r
		}

		if r == '-' || r == '.' || r == ' ' || r == '(' || r == ')' {
			return r
		}

		return -1
	}, n)
}
