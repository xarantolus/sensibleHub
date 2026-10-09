package store

import (
	"slices"
	"strings"
	"xarantolus/sensibleHub/store/music"
)

type ArtistSyncChanged struct {
	Artist string
	Sync   bool
}

func (ArtistSyncChanged) isEvent() {}

func artistKey(artist string) string {
	return strings.ToUpper(CleanName(artist))
}

// ArtistSynced reports whether songs by artist may be synced; an artist
// excluded with SetArtistSync overrides the songs' own setting.
func (m *Manager) ArtistSynced(artist string) bool {
	key := artistKey(artist)

	m.artistsLock.RLock()
	defer m.artistsLock.RUnlock()

	return !slices.ContainsFunc(m.UnsyncedArtists, func(a string) bool { return artistKey(a) == key })
}

// ShouldSync reports whether e is synced to devices and played by autoplay.
func (m *Manager) ShouldSync(e music.Entry) bool {
	return e.SyncSettings.Should && m.ArtistSynced(e.Artist())
}

// SetArtistSync excludes all songs by artist from syncing, or lifts that
// exclusion so each song follows its own setting again.
func (m *Manager) SetArtistSync(artist string, sync bool) error {
	key := artistKey(artist)

	m.SongsLock.Lock()
	defer m.SongsLock.Unlock()

	name := ""
	for _, e := range m.Songs {
		if artistKey(e.Artist()) == key {
			name = e.Artist()
			break
		}
	}
	if name == "" {
		return &NotFoundError{Kind: "artist", Key: artist}
	}

	m.artistsLock.Lock()
	previous := m.UnsyncedArtists
	m.UnsyncedArtists = slices.DeleteFunc(slices.Clone(previous), func(a string) bool { return artistKey(a) == key })
	if !sync {
		m.UnsyncedArtists = append(m.UnsyncedArtists, name)
		slices.Sort(m.UnsyncedArtists)
	}
	m.artistsLock.Unlock()

	if err := m.Save(false); err != nil {
		m.artistsLock.Lock()
		m.UnsyncedArtists = previous
		m.artistsLock.Unlock()
		return err
	}

	m.publish(ArtistSyncChanged{Artist: name, Sync: sync})
	return nil
}
