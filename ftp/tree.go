package ftp

import (
	"strings"
	"xarantolus/sensibleHub/store"

	"goftp.io/server/v2"
)

// album is part of the virtual file system
type album map[string][]*fileInfo

// tree maps artist directories to their albums
type tree map[string]album

const treeKey = "sensibleHub.tree"

// treeSnapshot takes a snapshot of the library when a user logs in, so the
// files don't change while a client syncs.
type treeSnapshot struct {
	server.NullNotifier
	manager *store.Manager
}

func (t *treeSnapshot) AfterUserLogin(ctx *server.Context, _, _ string, passMatched bool, err error) {
	if passMatched && err == nil {
		ctx.Sess.Data[treeKey] = buildTree(t.manager)
	}
}

func sessionTree(ctx *server.Context) tree {
	t, _ := ctx.Sess.Data[treeKey].(tree)
	return t
}

func buildTree(m *store.Manager) tree {
	artists := make(tree)

	uniquePaths := make(map[string]bool)

	// Create the virtual file system
	normalizedArtists := make(map[string]string)
	normalizedAlbums := make(map[string]string)

	for _, e := range m.AllEntries() {
		// Entries that should not be synced will not appear in the listing
		if !m.ShouldSync(e) {
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

		_, ok = artists[artistName]
		if !ok {
			artists[artistName] = make(album)
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

		artists[artistName][aname] = append(artists[artistName][aname], f)

		uniquePaths[upath] = true
	}

	return artists
}
