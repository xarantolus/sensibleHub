package store

import (
	"errors"
	"slices"
	"testing"
	"time"
	"xarantolus/sensibleHub/store/config"
	"xarantolus/sensibleHub/store/music"
)

func songBy(id, artist string, sync bool) music.Entry {
	e := song(id, sync, time.Now())
	e.MusicData.Artist = artist
	return e
}

func TestArtistExclusionOverridesSongsAndCanBeLifted(t *testing.T) {
	t.Chdir(t.TempDir())
	m := testManager(songBy("a", "AC/DC", true), songBy("b", "ac-dc", false), songBy("c", "Other", true))

	if err := m.SetArtistSync("acdc", false); err != nil {
		t.Fatal(err)
	}
	if m.ShouldSync(m.Songs["a"]) {
		t.Error("a song of an excluded artist must not sync, even when its own setting is on")
	}
	if !m.ShouldSync(m.Songs["c"]) {
		t.Error("other artists must be unaffected")
	}

	if err := m.SetArtistSync("AC/DC", true); err != nil {
		t.Fatal(err)
	}
	if !m.ShouldSync(m.Songs["a"]) || m.ShouldSync(m.Songs["b"]) {
		t.Error("lifting the exclusion must restore each song's own setting")
	}
	if len(m.UnsyncedArtists) != 0 {
		t.Errorf("unsynced artists = %v, want none", m.UnsyncedArtists)
	}
}

func TestSetArtistSyncUnknownArtist(t *testing.T) {
	t.Chdir(t.TempDir())
	m := testManager(songBy("a", "Someone", true))

	var nf *NotFoundError
	if err := m.SetArtistSync("Nobody", false); !errors.As(err, &nf) {
		t.Fatalf("got %v, want a NotFoundError", err)
	}
}

func TestArtistExclusionIsSaved(t *testing.T) {
	t.Chdir(t.TempDir())
	m := testManager(songBy("a", "Someone", true))
	if err := m.SetArtistSync("someone", false); err != nil {
		t.Fatal(err)
	}

	loaded, err := NewManager(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ShouldSync(loaded.Songs["a"]) {
		t.Error("the exclusion was lost on reload")
	}
}

func TestNextSongsSkipsExcludedArtists(t *testing.T) {
	t.Chdir(t.TempDir())
	m := testManager(songBy("a", "In", true), songBy("b", "Out", true), songBy("c", "In", true))
	if err := m.SetArtistSync("Out", false); err != nil {
		t.Fatal(err)
	}

	for range 50 {
		for _, s := range m.NextSongs("", 3, nil) {
			if s.ID == "b" {
				t.Fatal("suggested a song by an excluded artist")
			}
		}
	}
}

func TestUnsyncedGroupsExcludedArtistsSeparately(t *testing.T) {
	t.Chdir(t.TempDir())
	m := testManager(songBy("own", "Out", false), songBy("art", "Out", true), songBy("fine", "In", true))
	if err := m.SetArtistSync("Out", false); err != nil {
		t.Fatal(err)
	}

	groups := m.Unsynced()
	if len(groups) != 2 {
		t.Fatalf("got %d groups, want the song's own and the artist's", len(groups))
	}
	if !slices.Equal(ids(groups[0].Songs), []string{"own"}) {
		t.Errorf("own group = %v", ids(groups[0].Songs))
	}
	if groups[1].Title != "Out" || !slices.Equal(ids(groups[1].Songs), []string{"art"}) {
		t.Errorf("artist group = %q %v", groups[1].Title, ids(groups[1].Songs))
	}
}

func ids(entries []music.Entry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.ID
	}
	return out
}
