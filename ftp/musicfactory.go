package ftp

import (
	"strings"
	"xarantolus/sensibleHub/store"

	"goftp.io/server"
)

// musicDriverFactory is the ftp driver factory for this program.
// It implements the server.DriverFactory
type musicDriverFactory struct {
	*store.Manager
}

func (m *musicDriverFactory) NewDriver() (server.Driver, error) {
	entries := m.AllEntries()

	d := &musicDriver{
		Artists: make(map[string]album),
		manager: m.Manager,
		cfg:     m.GetConfig(),
	}

	uniquePaths := make(map[string]bool)

	// Create the virtual file system
	normalizedArtists := make(map[string]string)
	normalizedAlbums := make(map[string]string)

	for _, e := range entries {
		// Entries that should not be synced will not appear in the listing
		if !e.SyncSettings.Should {
			continue
		}

		art := store.CleanName(e.Artist())
		if art == "" {
			art = "Other Artist"
		}
		artistName, ok := normalizedArtists[strings.ToUpper(art)]
		if !ok {
			normalizedArtists[strings.ToUpper(e.Artist())] = art
			artistName = art
		}

		_, ok = d.Artists[artistName]
		if !ok {
			d.Artists[artistName] = make(album)
		}

		aname := store.CleanName(e.AlbumName())
		if aname == "" {
			aname = "Other Album"
		}
		if a, ok := normalizedAlbums[strings.ToUpper(aname)]; !ok {
			normalizedAlbums[strings.ToUpper(aname)] = aname
		} else {
			aname = a
		}

		f := fileInfoFromEntry(e)

		upath := strings.ToUpper(artistName + "/" + aname + "/" + f.Name())
		// Don't allow duplicate paths
		if uniquePaths[upath] {
			continue
		}

		d.Artists[artistName][aname] = append(d.Artists[artistName][aname], f)

		uniquePaths[upath] = true
	}

	return d, nil
}
